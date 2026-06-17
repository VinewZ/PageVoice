package tts

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/adrg/xdg"
	"github.com/vinewz/PageVoice/internal/data/sentences"
	"github.com/vinewz/PageVoice/internal/tts/piper"
	"github.com/vinewz/PageVoice/internal/tts/voices"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type Service struct {
	piperDir string
	mu       sync.Mutex
	stopCh   chan struct{}
	app      *application.App
}

func New(piperDir string) *Service {
	return &Service{
		piperDir: piperDir,
		stopCh:   make(chan struct{}),
	}
}

func (s *Service) ServiceStartup(ctx context.Context, opts application.ServiceOptions) error {
	s.app = application.Get()
	return nil
}

func (s *Service) SplitBook(dirName, language string) error {
	bookDir := filepath.Join(xdg.DataHome, "page-voice", "books", dirName)

	if _, err := os.Stat(filepath.Join(bookDir, "metadata.json")); err != nil {
		return fmt.Errorf("book %q not found", dirName)
	}

	text, err := os.ReadFile(filepath.Join(bookDir, "original.txt"))
	if err != nil {
		return fmt.Errorf("read original.txt: %w", err)
	}

	raw, err := sentences.Split(string(text), language)
	if err != nil {
		return fmt.Errorf("split sentences: %w", err)
	}

	chunkLength := 2500
	chunks := sentences.GroupChunks(raw, chunkLength)

	var sents []SentenceData
	for _, c := range chunks {
		for _, s := range c.Sentences {
			sents = append(sents, SentenceData{
				Index: len(sents),
				Text:  s,
				Chunk: c.Index,
			})
		}
	}

	sf := SentencesFile{
		Language:    language,
		ChunkLength: chunkLength,
		TotalChunks: len(chunks),
		Sentences:   sents,
	}

	data, err := json.MarshalIndent(sf, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal sentences.json: %w", err)
	}
	if err := os.WriteFile(filepath.Join(bookDir, "sentences.json"), data, 0644); err != nil {
		return fmt.Errorf("write sentences.json: %w", err)
	}

	if err := updateState(bookDir, func(st *state) {
		st.TotalChunks = len(chunks)
		st.UpdatedAt = nowStr()
	}); err != nil {
		return fmt.Errorf("update state: %w", err)
	}

	return nil
}

func (s *Service) StartSynthesis(dirName, voiceName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	bookDir := filepath.Join(xdg.DataHome, "page-voice", "books", dirName)
	audioDir := filepath.Join(bookDir, "audio")

	sf, err := readSentencesFile(bookDir)
	if err != nil {
		return fmt.Errorf("read sentences.json: %w", err)
	}

	st, err := readState(bookDir)
	if err != nil {
		return fmt.Errorf("read state.json: %w", err)
	}

	modelPath, configPath, err := voices.GetVoicePath(voiceName)
	if err != nil {
		return fmt.Errorf("voice: %w", err)
	}

	p := piper.New(s.piperDir)

	if st.TotalChunks == 0 {
		st.TotalChunks = sf.TotalChunks
	}

	if err := os.MkdirAll(audioDir, 0755); err != nil {
		return fmt.Errorf("create audio dir: %w", err)
	}

	stopCh := make(chan struct{})
	s.stopCh = stopCh

	if err := updateState(bookDir, func(st *state) {
		st.Status = "running"
		st.UpdatedAt = nowStr()
	}); err != nil {
		return err
	}

	go func() {
		defer func() {
			s.mu.Lock()
			s.stopCh = nil
			s.mu.Unlock()
		}()

		for chunkIdx := st.CurrentChunk; chunkIdx < st.TotalChunks; chunkIdx++ {
			select {
			case <-stopCh:
				updateState(bookDir, func(st *state) {
					st.Status = "paused"
					st.UpdatedAt = nowStr()
				})
				return
			default:
			}

			s.app.Event.Emit("tts:chunk-start", map[string]any{
				"dirName": dirName,
				"chunk":   chunkIdx,
				"total":   st.TotalChunks,
			})

			chunkText := buildChunkText(sf, chunkIdx)
			if chunkText == "" {
				continue
			}

			wavBytes, err := p.Synthesize(chunkText, modelPath, configPath)
			if err != nil {
				updateState(bookDir, func(st *state) {
					st.Status = "paused"
					st.UpdatedAt = nowStr()
				})
				s.app.Event.Emit("tts:chunk-error", map[string]any{
					"dirName": dirName,
					"chunk":   chunkIdx,
					"error":   err.Error(),
				})
				return
			}

			chunkFile := filepath.Join(audioDir, fmt.Sprintf("chunk_%03d.wav", chunkIdx+1))
			if err := os.WriteFile(chunkFile, wavBytes, 0644); err != nil {
				updateState(bookDir, func(st *state) {
					st.Status = "paused"
					st.UpdatedAt = nowStr()
				})
				s.app.Event.Emit("tts:chunk-error", map[string]any{
					"dirName": dirName,
					"chunk":   chunkIdx,
					"error":   err.Error(),
				})
				return
			}

			updateState(bookDir, func(st *state) {
				st.CurrentChunk = chunkIdx + 1
				st.UpdatedAt = nowStr()
			})

			s.app.Event.Emit("tts:chunk-complete", map[string]any{
				"dirName": dirName,
				"chunk":   chunkIdx,
				"total":   st.TotalChunks,
			})
		}

		updateState(bookDir, func(st *state) {
			st.Status = "completed"
			st.UpdatedAt = nowStr()
		})
	}()

	return nil
}

func (s *Service) StopSynthesis() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stopCh != nil {
		close(s.stopCh)
		s.stopCh = nil
	}
	return nil
}

func (s *Service) GetVoices() ([]VoiceInfo, error) {
	vl, err := voices.ListVoices()
	if err != nil {
		return nil, err
	}
	result := make([]VoiceInfo, len(vl))
	for i, v := range vl {
		result[i] = VoiceInfo{
			Name:       v.Name,
			Language:   v.Language,
			Quality:    v.Quality,
			Downloaded: v.Downloaded,
		}
	}
	return result, nil
}

func (s *Service) DownloadVoice(name string) error {
	return voices.DownloadVoice(name)
}

func (s *Service) GetSynthesisStatus(dirName string) (*SynthesisProgress, error) {
	bookDir := filepath.Join(xdg.DataHome, "page-voice", "books", dirName)
	st, err := readState(bookDir)
	if err != nil {
		return nil, fmt.Errorf("read state.json: %w", err)
	}
	return &SynthesisProgress{
		DirName:      dirName,
		CurrentChunk: st.CurrentChunk,
		TotalChunks:  st.TotalChunks,
		Status:       st.Status,
	}, nil
}

func (s *Service) GetSentences(dirName string) (string, error) {
	path := filepath.Join(xdg.DataHome, "page-voice", "books", dirName, "sentences.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (s *Service) GetAudio(dirName string, chunkIndex int) ([]byte, error) {
	path := filepath.Join(xdg.DataHome, "page-voice", "books", dirName, "audio", fmt.Sprintf("chunk_%03d.wav", chunkIndex))
	return os.ReadFile(path)
}

func (s *Service) GetGeneratedChunks(dirName string) ([]int, error) {
	audioDir := filepath.Join(xdg.DataHome, "page-voice", "books", dirName, "audio")
	entries, err := os.ReadDir(audioDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var chunks []int
	for _, e := range entries {
		var idx int
		if n, _ := fmt.Sscanf(e.Name(), "chunk_%d.wav", &idx); n == 1 {
			chunks = append(chunks, idx)
		}
	}
	return chunks, nil
}

func buildChunkText(sf *SentencesFile, chunkIdx int) string {
	var parts []string
	for _, s := range sf.Sentences {
		if s.Chunk == chunkIdx {
			parts = append(parts, s.Text)
		}
	}
	if len(parts) == 0 {
		return ""
	}
	result := ""
	for _, p := range parts {
		if result != "" {
			result += " "
		}
		result += p
	}
	return result
}

type state struct {
	Status       string `json:"status"`
	ChunkLength  int    `json:"chunkLength"`
	CurrentChunk int    `json:"currentChunk"`
	TotalChunks  int    `json:"totalChunks"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

func readState(bookDir string) (*state, error) {
	data, err := os.ReadFile(filepath.Join(bookDir, "state.json"))
	if err != nil {
		return nil, err
	}
	var st state
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, err
	}
	return &st, nil
}

func updateState(bookDir string, fn func(st *state)) error {
	st, err := readState(bookDir)
	if err != nil {
		return err
	}
	fn(st)
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(bookDir, "state.json"), data, 0644)
}

func readSentencesFile(bookDir string) (*SentencesFile, error) {
	data, err := os.ReadFile(filepath.Join(bookDir, "sentences.json"))
	if err != nil {
		return nil, err
	}
	var sf SentencesFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return nil, err
	}
	return &sf, nil
}

func nowStr() string {
	return timeNow().UTC().Format("2006-01-02T15:04:05Z")
}

var timeNow = time.Now
