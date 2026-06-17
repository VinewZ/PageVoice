package textupload

type Metadata struct {
	Title  string `json:"title,omitempty"`
	Author string `json:"author,omitempty"`
}

type Sentence struct {
	Index int    `json:"index"`
	Text  string `json:"text"`
}

type UploadResult struct {
	FileName      string     `json:"fileName"`
	FileType      string     `json:"fileType"`
	Language      string     `json:"language"`
	Metadata      Metadata   `json:"metadata"`
	Sentences     []Sentence `json:"sentences"`
	TotalChars    int        `json:"totalChars"`
	SentenceCount int        `json:"sentenceCount"`
}
