package voices

type VoiceInfo struct {
	Name       string `json:"name"`
	Language   string `json:"language"`
	Quality    string `json:"quality"`
	Downloaded bool   `json:"downloaded"`
}

var PopularVoices = []VoiceInfo{
	{Name: "en_US-lessac-medium", Language: "english", Quality: "medium"},
	{Name: "en_US-libritts-high", Language: "english", Quality: "high"},
	{Name: "en_GB-alan-medium", Language: "english", Quality: "medium"},
	{Name: "de_DE-thorsten-medium", Language: "german", Quality: "medium"},
	{Name: "fr_FR-siwis-medium", Language: "french", Quality: "medium"},
	{Name: "es_ES-sharvard-medium", Language: "spanish", Quality: "medium"},
	{Name: "it_IT-riccardo-medium", Language: "italian", Quality: "medium"},
	{Name: "pt_BR-faber-medium", Language: "portuguese", Quality: "medium"},
	{Name: "nl_NLD-mls-medium", Language: "dutch", Quality: "medium"},
	{Name: "cs_CZ-jirka-medium", Language: "czech", Quality: "medium"},
}
