package storage

import (
	"os"
	"path/filepath"
)

// EnsureDataDir creates the data directory if it doesn't exist
func EnsureDataDir(dbPath string) error {
	dir := filepath.Dir(dbPath)
	return os.MkdirAll(dir, 0755)
}