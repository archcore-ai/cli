package sync

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"archcore-cli/templates"
)

// FileState represents the current on-disk state of a single document file.
type FileState struct {
	RelPath string // relative to .archcore/, e.g. "auth/jwt-strategy.adr.md"
	AbsPath string
	Hash    string
}

// HashFile computes the SHA-256 hex digest of the file at the given path.
func HashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("opening file for hash: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("hashing file: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// ScanFiles walks .archcore/ recursively and returns the current state of
// every .md file found. Skips hidden directories, settings.json, and
// .sync-state.json.
func ScanFiles(baseDir string) ([]FileState, error) {
	archcoreDir := filepath.Join(baseDir, ".archcore")
	var files []FileState

	err := templates.WalkArchcoreFiles(archcoreDir, func(path string, d fs.DirEntry) error {
		relPath, _ := filepath.Rel(archcoreDir, path)
		relPath = filepath.ToSlash(relPath)

		hash, err := HashFile(path)
		if err != nil {
			return fmt.Errorf("hashing %s: %w", relPath, err)
		}

		files = append(files, FileState{
			RelPath: relPath,
			AbsPath: path,
			Hash:    hash,
		})
		return nil
	})

	if os.IsNotExist(err) {
		return nil, nil
	}
	return files, err
}
