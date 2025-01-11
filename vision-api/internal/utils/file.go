package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// FileInfo contains file metadata and content information
type FileInfo struct {
	Path      string
	Name      string
	Extension string
	Size      int64
	Hash      string
	MimeType  string
}

// TempFileManager manages temporary file creation and cleanup
type TempFileManager struct {
	tempDir string
	files   map[string]string
	mu      sync.RWMutex
}

// NewTempFileManager creates a new temporary file manager
func NewTempFileManager(tempDir string) (*TempFileManager, error) {
	if tempDir == "" {
		tempDir = os.TempDir()
	}

	// Ensure temp directory exists
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	return &TempFileManager{
		tempDir: tempDir,
		files:   make(map[string]string),
	}, nil
}

// CreateTemp creates a temporary file and tracks it
func (tm *TempFileManager) CreateTemp(prefix string) (*os.File, error) {
	file, err := os.CreateTemp(tm.tempDir, prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	tm.mu.Lock()
	tm.files[file.Name()] = file.Name()
	tm.mu.Unlock()

	return file, nil
}

// Cleanup removes all tracked temporary files
func (tm *TempFileManager) Cleanup() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	var lastErr error
	for _, path := range tm.files {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			lastErr = fmt.Errorf("failed to remove temp file %s: %w", path, err)
		}
		delete(tm.files, path)
	}

	return lastErr
}

// GetFileInfo retrieves file information safely
func GetFileInfo(path string) (*FileInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file stats: %w", err)
	}

	// Calculate file hash
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return nil, fmt.Errorf("failed to calculate file hash: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	mimeType := getMimeType(ext)

	return &FileInfo{
		Path:      path,
		Name:      stat.Name(),
		Extension: ext,
		Size:      stat.Size(),
		Hash:      hex.EncodeToString(hash.Sum(nil)),
		MimeType:  mimeType,
	}, nil
}

// EnsureDirectory ensures a directory exists and is writeable
func EnsureDirectory(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(path, 0755)
		}
		return fmt.Errorf("failed to check directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path exists but is not a directory: %s", path)
	}

	// Check if directory is writeable
	testFile := filepath.Join(path, ".write_test")
	f, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("directory is not writeable: %w", err)
	}
	f.Close()
	os.Remove(testFile)

	return nil
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destination.Close()

	if _, err := io.Copy(destination, source); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

// getMimeType returns the MIME type for common image extensions
func getMimeType(ext string) string {
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".bmp":
		return "image/bmp"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}

// SafeFileName ensures a filename is safe for use
func SafeFileName(name string) string {
	// Remove any path components
	name = filepath.Base(name)

	// Replace unsafe characters
	replacer := strings.NewReplacer(
		" ", "_",
		"&", "_and_",
		"+", "_plus_",
		"=", "_eq_",
		"#", "_hash_",
		"%", "_pct_",
		"$", "_dollar_",
		"@", "_at_",
		"!", "_bang_",
		"?", "_qmark_",
		"<", "_lt_",
		">", "_gt_",
		":", "_colon_",
		";", "_semi_",
		"\"", "_quote_",
		"'", "_squote_",
		"\\", "_bslash_",
		"/", "_slash_",
		"|", "_pipe_",
	)
	return replacer.Replace(name)
}