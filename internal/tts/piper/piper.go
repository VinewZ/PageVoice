package piper

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

	return encodeWAV(raw, sampleRate), nil
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

func encodeWAV(pcm []byte, sampleRate int) []byte {
	dataSize := len(pcm)
	fileSize := 36 + dataSize

	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, []byte("RIFF"))
	binary.Write(buf, binary.LittleEndian, int32(fileSize))
	binary.Write(buf, binary.LittleEndian, []byte("WAVE"))

	binary.Write(buf, binary.LittleEndian, []byte("fmt "))
	binary.Write(buf, binary.LittleEndian, int32(16))
	binary.Write(buf, binary.LittleEndian, int16(1))
	binary.Write(buf, binary.LittleEndian, int16(1))
	binary.Write(buf, binary.LittleEndian, int32(sampleRate))
	binary.Write(buf, binary.LittleEndian, int32(sampleRate*2))
	binary.Write(buf, binary.LittleEndian, int16(2))
	binary.Write(buf, binary.LittleEndian, int16(16))

	binary.Write(buf, binary.LittleEndian, []byte("data"))
	binary.Write(buf, binary.LittleEndian, int32(dataSize))
	buf.Write(pcm)

	return buf.Bytes()
}
