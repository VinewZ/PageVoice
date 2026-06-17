package wav

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

func Encode(pcm []byte, sampleRate int) ([]byte, error) {
	var buf bytes.Buffer
	if err := EncodeWriter(pcm, sampleRate, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func EncodeWriter(pcm []byte, sampleRate int, w io.Writer) error {
	dataSize := len(pcm)
	fileSize := int32(36 + dataSize)

	write := func(v any) error {
		return binary.Write(w, binary.LittleEndian, v)
	}

	if err := write([]byte("RIFF")); err != nil {
		return fmt.Errorf("write riff header: %w", err)
	}
	if err := write(fileSize); err != nil {
		return fmt.Errorf("write file size: %w", err)
	}
	if err := write([]byte("WAVE")); err != nil {
		return fmt.Errorf("write wave marker: %w", err)
	}

	if err := write([]byte("fmt ")); err != nil {
		return fmt.Errorf("write fmt chunk id: %w", err)
	}
	if err := write(int32(16)); err != nil {
		return fmt.Errorf("write fmt chunk size: %w", err)
	}
	if err := write(int16(1)); err != nil {
		return fmt.Errorf("write audio format: %w", err)
	}
	if err := write(int16(1)); err != nil {
		return fmt.Errorf("write channels: %w", err)
	}
	if err := write(int32(sampleRate)); err != nil {
		return fmt.Errorf("write sample rate: %w", err)
	}
	if err := write(int32(sampleRate * 2)); err != nil {
		return fmt.Errorf("write byte rate: %w", err)
	}
	if err := write(int16(2)); err != nil {
		return fmt.Errorf("write block align: %w", err)
	}
	if err := write(int16(16)); err != nil {
		return fmt.Errorf("write bits per sample: %w", err)
	}

	if err := write([]byte("data")); err != nil {
		return fmt.Errorf("write data chunk id: %w", err)
	}
	if err := write(int32(dataSize)); err != nil {
		return fmt.Errorf("write data chunk size: %w", err)
	}
	if _, err := w.Write(pcm); err != nil {
		return fmt.Errorf("write pcm data: %w", err)
	}

	return nil
}
