package piper

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/adrg/xdg"
)

const (
	piperSubDir      = "page-voice/piper"
	versionMarker    = ".version"
	markerFileExists = "extracted"
)

func EnsureExtracted() (string, error) {
	piperDir := filepath.Join(xdg.DataHome, piperSubDir)

	if isExtracted(piperDir) {
		return piperDir, nil
	}

	if err := os.MkdirAll(piperDir, 0755); err != nil {
		return "", fmt.Errorf("create piper dir: %w", err)
	}

	if len(archive) == 0 {
		return "", fmt.Errorf("piper binary not embedded for this architecture (GOARCH=%s). Only amd64 is supported.", runtime.GOARCH)
	}

	if err := extractTarGz(archive, piperDir); err != nil {
		return "", fmt.Errorf("extract piper archive: %w", err)
	}

	binPath := filepath.Join(piperDir, "piper")
	if err := os.Chmod(binPath, 0755); err != nil {
		return "", fmt.Errorf("chmod piper binary: %w", err)
	}

	if err := writeMarker(piperDir); err != nil {
		return "", fmt.Errorf("write version marker: %w", err)
	}

	return piperDir, nil
}

func isExtracted(piperDir string) bool {
	marker := filepath.Join(piperDir, versionMarker)
	data, err := os.ReadFile(marker)
	if err != nil {
		return false
	}
	return string(data) == markerFileExists
}

func writeMarker(piperDir string) error {
	marker := filepath.Join(piperDir, versionMarker)
	return os.WriteFile(marker, []byte(markerFileExists), 0644)
}

func extractTarGz(data []byte, destDir string) error {
	gzr, err := gzip.NewReader(&byteReader{data: data})
	if err != nil {
		return fmt.Errorf("create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar entry: %w", err)
		}

		rel, err := stripPrefix(header.Name, "piper/")
		if err != nil {
			continue
		}

		target := filepath.Join(destDir, rel)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)&os.ModePerm); err != nil {
				return fmt.Errorf("create dir %s: %w", target, err)
			}

		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("create parent dir for %s: %w", target, err)
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode)&os.ModePerm)
			if err != nil {
				return fmt.Errorf("create file %s: %w", target, err)
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return fmt.Errorf("write file %s: %w", target, err)
			}
			f.Close()

		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("create parent dir for symlink %s: %w", target, err)
			}
			if err := os.Symlink(header.Linkname, target); err != nil {
				return fmt.Errorf("create symlink %s -> %s: %w", target, header.Linkname, err)
			}
		}
	}

	return nil
}

func stripPrefix(name, prefix string) (string, error) {
	if len(name) < len(prefix) || name[:len(prefix)] != prefix {
		return "", fmt.Errorf("unexpected path %q (missing prefix %q)", name, prefix)
	}
	return name[len(prefix):], nil
}

type byteReader struct {
	data []byte
	pos  int
}

func (b *byteReader) Read(p []byte) (int, error) {
	if b.pos >= len(b.data) {
		return 0, io.EOF
	}
	n := copy(p, b.data[b.pos:])
	b.pos += n
	return n, nil
}
