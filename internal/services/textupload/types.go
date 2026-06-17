package textupload

type TOCEntry struct {
	Title string `json:"title"`
	Depth int    `json:"depth"`
}

type Metadata struct {
	Title      string     `json:"title,omitempty"`
	Author     string     `json:"author,omitempty"`
	Language   string     `json:"language"`
	SourceFile string     `json:"sourceFile"`
	ImportedAt string     `json:"importedAt"`
	TOC        []TOCEntry `json:"toc,omitempty"`
}

type State struct {
	Status       string `json:"status"`
	ChunkLength  int    `json:"chunkLength"`
	CurrentChunk int    `json:"currentChunk"`
	TotalChunks  int    `json:"totalChunks"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

type LibraryEntry struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	DirName string `json:"dirName"`
}

type Sentence struct {
	Index int    `json:"index"`
	Text  string `json:"text"`
}

type BookDetail struct {
	ID       string   `json:"id"`
	DirName  string   `json:"dirName"`
	Metadata Metadata `json:"metadata"`
	State    State    `json:"state"`
}

type UploadResult struct {
	FileName      string     `json:"fileName"`
	FileType      string     `json:"fileType"`
	Language      string     `json:"language"`
	Metadata      Metadata   `json:"metadata"`
	Sentences     []Sentence `json:"sentences"`
	TotalChars    int        `json:"totalChars"`
	SentenceCount int        `json:"sentenceCount"`
	BookID        string     `json:"bookId"`
	DataDir       string     `json:"dataDir"`
}
