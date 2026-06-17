package textupload

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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

	return &UploadResult{
		FileName:      fileName,
		FileType:      ext[1:],
		Language:      language,
		Metadata:      meta,
		Sentences:     sentences,
		TotalChars:    len(text),
		SentenceCount: len(sentences),
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
