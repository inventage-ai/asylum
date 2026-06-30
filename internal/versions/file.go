package versions

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Read reads and parses versions.json from the given path.
// Returns nil if the file does not exist or contains invalid JSON, which the
// caller uses to trigger a blocking fetch.
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

// Write marshals the version map to JSON and writes it atomically: it writes to
// a per-write unique temp file in the target directory, then renames it onto the
// target path. The unique name keeps concurrent asylum invocations (which share
// one ~/.asylum/versions.json) from clobbering each other's temp file.
func Write(path string, vm VersionMap) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(vm, "", "  ")
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, "versions-*.json")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if err := tmp.Chmod(0644); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}

	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return nil
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

// NeedsRefresh reports whether the version file should be refetched: either it
// is older than stalenessDur, or the loaded map is missing an entry for a
// tracked agent (e.g. one whose fetch failed during an earlier partial fetch),
// so missing agents are retried before the staleness interval elapses.
func NeedsRefresh(path string, vm VersionMap, stalenessDur time.Duration) (bool, error) {
	stale, err := IsStale(path, stalenessDur)
	if err != nil {
		return false, err
	}
	if stale {
		return true, nil
	}
	for _, name := range AgentNames() {
		if _, ok := vm[name]; !ok {
			return true, nil
		}
	}
	return false, nil
}
