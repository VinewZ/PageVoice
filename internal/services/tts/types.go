package tts

type VoiceInfo struct {
	Name       string `json:"name"`
	Language   string `json:"language"`
	Quality    string `json:"quality"`
	Downloaded bool   `json:"downloaded"`
}

type SynthesisProgress struct {
	DirName      string `json:"dirName"`
	CurrentChunk int    `json:"currentChunk"`
	TotalChunks  int    `json:"totalChunks"`
	Status       string `json:"status"`
	Error        string `json:"error"`
}

type SentenceData struct {
	Index int    `json:"index"`
	Text  string `json:"text"`
	Chunk int    `json:"chunk"`
}

type SentencesFile struct {
	Language    string         `json:"language"`
	ChunkLength int            `json:"chunkLength"`
	TotalChunks int            `json:"totalChunks"`
	Sentences   []SentenceData `json:"sentences"`
}
