package voices

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
)

const hfBase = "https://huggingface.co/rhasspy/piper-voices/resolve/main"

func VoicesDir() string {
	return filepath.Join(xdg.DataHome, "page-voice", "piper", "voices")
}

func EnsureVoicesDir() error {
	return os.MkdirAll(VoicesDir(), 0755)
}

func ListVoices() ([]VoiceInfo, error) {
	if err := EnsureVoicesDir(); err != nil {
		return nil, fmt.Errorf("ensure voices dir: %w", err)
	}

	result := make([]VoiceInfo, len(PopularVoices))
	copy(result, PopularVoices)

	entries, err := os.ReadDir(VoicesDir())
	if err != nil {
		return result, nil
	}

	downloaded := make(map[string]bool)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		modelPath := filepath.Join(VoicesDir(), e.Name(), e.Name()+".onnx")
		if _, err := os.Stat(modelPath); err == nil {
			downloaded[e.Name()] = true
		}
	}

	for i := range result {
		result[i].Downloaded = downloaded[result[i].Name]
	}

	return result, nil
}

func VoiceExists(name string) bool {
	path := filepath.Join(VoicesDir(), name, name+".onnx")
	_, err := os.Stat(path)
	return err == nil
}

func GetVoicePath(name string) (string, string, error) {
	modelPath := filepath.Join(VoicesDir(), name, name+".onnx")
	configPath := filepath.Join(VoicesDir(), name, name+".onnx.json")

	if _, err := os.Stat(modelPath); err != nil {
		return "", "", fmt.Errorf("voice %q not downloaded", name)
	}

	return modelPath, configPath, nil
}

func DownloadVoice(name string) error {
	if VoiceExists(name) {
		return nil
	}

	voiceDir := filepath.Join(VoicesDir(), name)
	if err := os.MkdirAll(voiceDir, 0755); err != nil {
		return fmt.Errorf("create voice dir: %w", err)
	}

	parts := strings.SplitN(name, "-", 3)
	if len(parts) < 3 {
		return fmt.Errorf("invalid voice name %q: expected format lang_region-voice-quality", name)
	}
	lang := parts[0]
	langShort := strings.SplitN(lang, "_", 2)[0]
	voiceName := parts[1]
	quality := parts[2]

	for _, ext := range []string{".onnx", ".onnx.json"} {
		url := fmt.Sprintf("%s/%s/%s/%s/%s/%s%s", hfBase, langShort, lang, voiceName, quality, name, ext)
		dst := filepath.Join(voiceDir, name+ext)

		if err := downloadFile(url, dst); err != nil {
			os.RemoveAll(voiceDir)
			return fmt.Errorf("download %s: %w", url, err)
		}
	}

	return nil
}

func downloadFile(url, dst string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	f, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}
