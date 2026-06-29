package versions

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Read reads and parses versions.json from the given path.
// Returns an empty VersionMap if the file does not exist or contains invalid JSON.
func Read(path string) (VersionMap, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	vm := make(VersionMap)
	if err := json.Unmarshal(data, &vm); err != nil {
		return nil, nil
	}
	return vm, nil
}

// Write marshals the version map to JSON and writes it atomically
// (write to temp file, then rename) in the same directory as the target path.
func Write(path string, vm VersionMap) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(vm, "", "  ")
	if err != nil {
		return err
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
}

// StaleSince returns the time when the file was last updated, or nil if the
// file does not exist. The caller can use this to determine if a background
// refresh is needed (e.g., if time.Since() > 24 hours).
func StaleSince(path string) (*time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	t := info.ModTime()
	return &t, nil
}

// IsStale returns true if the file exists and was modified more than
// stalenessDur ago. Returns false if the file does not exist.
func IsStale(path string, stalenessDur time.Duration) (bool, error) {
	t, err := StaleSince(path)
	if err != nil {
		return false, err
	}
	if t == nil {
		return false, nil
	}
	return time.Since(*t) > stalenessDur, nil
}
