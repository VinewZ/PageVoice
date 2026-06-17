package textupload

import (
	"bytes"
	"crypto/rand"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/kapmahc/epub"
	"github.com/ledongthuc/pdf"
	"github.com/neurosnap/sentences"
	"golang.org/x/net/html"
)

//go:embed language-data/*.json
var languageData embed.FS

type Service struct{}

func New() *Service {
	return &Service{}
}

func (s *Service) ProcessFile(fileName string, fileData []byte, language string) (*UploadResult, error) {
	ext := strings.ToLower(filepath.Ext(fileName))

	tmpFile, err := os.CreateTemp("", "pagevoice-*"+ext)
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.Write(fileData); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("write temp file: %w", err)
	}
	tmpFile.Close()

	var text string
	var meta Metadata

	switch ext {
	case ".pdf":
		text, meta, err = extractPDF(tmpPath)
	case ".epub":
		text, meta, err = extractEPUB(tmpPath)
	case ".txt":
		text = string(fileData)
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}
	if err != nil {
		return nil, err
	}

	sentences, err := splitSentences(text, language)
	if err != nil {
		return nil, err
	}

	title := meta.Title
	if title == "" {
		title = strings.TrimSuffix(fileName, ext)
	}

	id, dir, err := saveToDisk(text, meta, language, title, fileName)
	if err != nil {
		return nil, err
	}

	return &UploadResult{
		FileName:      fileName,
		FileType:      ext[1:],
		Language:      language,
		Metadata:      meta,
		Sentences:     sentences,
		TotalChars:    len(text),
		SentenceCount: len(sentences),
		BookID:        id,
		DataDir:       dir,
	}, nil
}

func extractPDF(path string) (string, Metadata, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", Metadata{}, fmt.Errorf("open pdf: %w", err)
	}
	defer f.Close()

	meta := Metadata{}
	trailer := r.Trailer()
	info := trailer.Key("Info")
	if info.Kind() != pdf.Null {
		meta.Title = info.Key("Title").Text()
		meta.Author = info.Key("Author").Text()
	}

	textReader, err := r.GetPlainText()
	if err != nil {
		return "", Metadata{}, fmt.Errorf("extract pdf text: %w", err)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, textReader); err != nil {
		return "", Metadata{}, fmt.Errorf("read pdf text: %w", err)
	}

	return buf.String(), meta, nil
}

func extractEPUB(path string) (string, Metadata, error) {
	bk, err := epub.Open(path)
	if err != nil {
		return "", Metadata{}, fmt.Errorf("open epub: %w", err)
	}
	defer bk.Close()

	meta := Metadata{}
	if len(bk.Opf.Metadata.Title) > 0 {
		meta.Title = bk.Opf.Metadata.Title[0]
	}
	if len(bk.Opf.Metadata.Creator) > 0 {
		meta.Author = bk.Opf.Metadata.Creator[0].Data
	}

	var textBuilder strings.Builder
	for _, item := range bk.Opf.Spine.Items {
		var href string
		for _, mf := range bk.Opf.Manifest {
			if mf.ID == item.IDref {
				href = mf.Href
				break
			}
		}
		if href == "" {
			continue
		}

		rc, err := bk.Open(href)
		if err != nil {
			continue
		}

		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			continue
		}

		t := extractTextFromHTML(string(content))
		textBuilder.WriteString(t)
		textBuilder.WriteString("\n")
	}

	return textBuilder.String(), meta, nil
}

func extractTextFromHTML(content string) string {
	doc, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return ""
	}

	var buf strings.Builder
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			t := strings.TrimSpace(n.Data)
			if t != "" {
				buf.WriteString(t)
				buf.WriteString(" ")
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return buf.String()
}

func splitSentences(text string, language string) ([]Sentence, error) {
	lang := strings.ToLower(language)
	jsonPath := fmt.Sprintf("language-data/%s.json", lang)

	b, err := languageData.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("unsupported language %q: %w", language, err)
	}

	training, err := sentences.LoadTraining(b)
	if err != nil {
		return nil, fmt.Errorf("load training data for %q: %w", language, err)
	}

	tokenizer := sentences.NewSentenceTokenizer(training)
	raw := tokenizer.Tokenize(text)

	result := make([]Sentence, len(raw))
	for i, s := range raw {
		result[i] = Sentence{
			Index: i + 1,
			Text:  strings.TrimSpace(s.Text),
		}
	}
	return result, nil
}

// --- disk persistence ---

var nonAlpha = regexp.MustCompile(`[^a-z0-9]+`)

func genID() string {
	n, err := rand.Int(rand.Reader, big.NewInt(0x1000000))
	if err != nil {
		return "000000"
	}
	return fmt.Sprintf("%06x", n.Int64())
}

func dirName(title string) string {
	s := strings.TrimSpace(strings.ToLower(title))
	s = nonAlpha.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 48 {
		s = s[:48]
	}
	if s == "" {
		s = "untitled"
	}
	return s + "-" + genID()
}

type saveResult struct {
	ID  string
	Dir string
}

func saveToDisk(text string, meta Metadata, language, title, fileName string) (string, string, error) {
	id := genID()
	dir := dirName(title)
	bookDir := filepath.Join(xdg.DataHome, "page-voice", "books", dir)

	if err := os.MkdirAll(bookDir, 0755); err != nil {
		return "", "", fmt.Errorf("create book dir: %w", err)
	}

	if err := os.WriteFile(filepath.Join(bookDir, "original.txt"), []byte(text), 0644); err != nil {
		return "", "", fmt.Errorf("write original.txt: %w", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)

	meta.Language = language
	meta.SourceFile = fileName
	meta.ImportedAt = now

	metaBytes, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return "", "", fmt.Errorf("marshal metadata: %w", err)
	}
	if err := os.WriteFile(filepath.Join(bookDir, "metadata.json"), metaBytes, 0644); err != nil {
		return "", "", fmt.Errorf("write metadata.json: %w", err)
	}

	state := State{
		Status:       "pending",
		ChunkLength:  250,
		CurrentChunk: 0,
		TotalChunks:  0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	stateBytes, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return "", "", fmt.Errorf("marshal state: %w", err)
	}
	if err := os.WriteFile(filepath.Join(bookDir, "state.json"), stateBytes, 0644); err != nil {
		return "", "", fmt.Errorf("write state.json: %w", err)
	}

	if err := appendLibrary(id, title, dir); err != nil {
		return "", "", fmt.Errorf("update library.json: %w", err)
	}

	return id, bookDir, nil
}

func (s *Service) GetLibrary() ([]LibraryEntry, error) {
	libPath := libraryPath()
	data, err := os.ReadFile(libPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []LibraryEntry{}, nil
		}
		return nil, fmt.Errorf("read library.json: %w", err)
	}
	var entries []LibraryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parse library.json: %w", err)
	}
	return entries, nil
}

func libraryPath() string {
	return filepath.Join(xdg.DataHome, "page-voice", "library.json")
}

func (s *Service) GetBook(dirName string) (*BookDetail, error) {
	bookDir := filepath.Join(xdg.DataHome, "page-voice", "books", dirName)

	metaBytes, err := os.ReadFile(filepath.Join(bookDir, "metadata.json"))
	if err != nil {
		return nil, fmt.Errorf("read metadata.json: %w", err)
	}
	var meta Metadata
	if err := json.Unmarshal(metaBytes, &meta); err != nil {
		return nil, fmt.Errorf("parse metadata.json: %w", err)
	}

	stateBytes, err := os.ReadFile(filepath.Join(bookDir, "state.json"))
	if err != nil {
		return nil, fmt.Errorf("read state.json: %w", err)
	}
	var state State
	if err := json.Unmarshal(stateBytes, &state); err != nil {
		return nil, fmt.Errorf("parse state.json: %w", err)
	}

	id := dirName[strings.LastIndex(dirName, "-")+1:]

	return &BookDetail{
		ID:       id,
		DirName:  dirName,
		Metadata: meta,
		State:    state,
	}, nil
}

func appendLibrary(id, title, dir string) error {
	libPath := libraryPath()

	var entries []LibraryEntry

	if data, err := os.ReadFile(libPath); err == nil {
		json.Unmarshal(data, &entries)
	}

	entries = append(entries, LibraryEntry{ID: id, Title: title, DirName: dir})

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(libPath), 0755); err != nil {
		return err
	}

	return os.WriteFile(libPath, data, 0644)
}
