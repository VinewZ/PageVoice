package piper

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/vinewz/PageVoice/internal/tts/wav"
)

type Piper struct {
	piperDir string
}

type voiceConfig struct {
	Audio struct {
		SampleRate int `json:"sample_rate"`
	} `json:"audio"`
}

func New(piperDir string) *Piper {
	return &Piper{piperDir: piperDir}
}

func (p *Piper) BinaryPath() string {
	return filepath.Join(p.piperDir, "piper")
}

func (p *Piper) EspeakDataDir() string {
	return filepath.Join(p.piperDir, "espeak-ng-data")
}

func (p *Piper) Synthesize(text string, modelPath string, configPath string) ([]byte, error) {
	sampleRate, err := readSampleRate(configPath)
	if err != nil {
		return nil, fmt.Errorf("read voice config: %w", err)
	}

	cmd := exec.Command(p.BinaryPath(),
		"--model", modelPath,
		"--espeak_data", p.EspeakDataDir(),
		"--output-raw",
	)
	cmd.Stdin = strings.NewReader(text)

	env := os.Environ()
	env = append(env, "LD_LIBRARY_PATH="+p.piperDir+":"+os.Getenv("LD_LIBRARY_PATH"))
	cmd.Env = env

	raw, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("run piper: %w", err)
	}

	return wav.Encode(raw, sampleRate)
}

func readSampleRate(configPath string) (int, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return 22050, nil
	}

	var cfg voiceConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return 22050, nil
	}

	if cfg.Audio.SampleRate <= 0 {
		return 22050, nil
	}

	return cfg.Audio.SampleRate, nil
}
