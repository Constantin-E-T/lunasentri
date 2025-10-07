package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
)

// EnsureDataDir creates the data directory if it doesn't exist
func EnsureDataDir(dbPath string) error {
	dir := filepath.Dir(dbPath)
	return os.MkdirAll(dir, 0755)
}

// HashSecret computes SHA-256 hash of the given secret and returns it as hex string
func HashSecret(secret string) string {
	hash := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(hash[:])
}