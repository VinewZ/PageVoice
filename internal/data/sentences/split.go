package sentences

import (
	"fmt"
	"strings"

	"github.com/neurosnap/sentences"
	"github.com/vinewz/PageVoice/internal/data"
)

type Chunk struct {
	Index     int
	Sentences []string
	Text      string
}

func Split(text string, language string) ([]string, error) {
	lang := strings.ToLower(language)
	jsonPath := fmt.Sprintf("language-data/%s.json", lang)

	b, err := data.LangData.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("unsupported language %q: %w", language, err)
	}

	training, err := sentences.LoadTraining(b)
	if err != nil {
		return nil, fmt.Errorf("load training data for %q: %w", language, err)
	}

	tokenizer := sentences.NewSentenceTokenizer(training)
	raw := tokenizer.Tokenize(text)

	result := make([]string, len(raw))
	for i, s := range raw {
		result[i] = strings.TrimSpace(s.Text)
	}
	return result, nil
}

func GroupChunks(sentences []string, chunkLength int) []Chunk {
	if chunkLength <= 0 {
		chunkLength = 2500
	}

	var chunks []Chunk
	var current []string
	charCount := 0

	for _, s := range sentences {
		sLen := len(s) + 1
		if charCount+sLen > chunkLength && len(current) > 0 {
			chunks = append(chunks, Chunk{
				Index:     len(chunks),
				Sentences: current,
				Text:      strings.Join(current, " "),
			})
			current = nil
			charCount = 0
		}
		current = append(current, s)
		charCount += sLen
	}

	if len(current) > 0 {
		chunks = append(chunks, Chunk{
			Index:     len(chunks),
			Sentences: current,
			Text:      strings.Join(current, " "),
		})
	}

	return chunks
}
