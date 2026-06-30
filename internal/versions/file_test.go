package versions

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestReadWrite(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/versions.json"

	// Test 1: missing file returns nil, nil (not an error)
	vm, err := Read(path)
	if err != nil {
		t.Fatalf("Read(missing) returned error: %v", err)
	}
	if vm != nil {
		t.Fatalf("Read(missing) returned non-nil map: %v", vm)
	}

	// Test 2: round-trip
	original := VersionMap{"claude": "v2.1.195", "gemini": "0.8.0"}
	if err := Write(path, original); err != nil {
		t.Fatalf("Write() returned error: %v", err)
	}
	loaded, err := Read(path)
	if err != nil {
		t.Fatalf("Read() returned error: %v", err)
	}
	for k, v := range original {
		if loaded[k] != v {
			t.Errorf("Read() [%q] = %q, want %q", k, loaded[k], v)
		}
	}

	// Test 3: corrupt file returns nil, nil (not an error)
	if err := os.WriteFile(path, []byte("not valid json{"), 0644); err != nil {
		t.Fatal(err)
	}
	loaded, err = Read(path)
	if err != nil {
		t.Fatalf("Read(corrupt) returned error: %v", err)
	}
	if loaded != nil {
		t.Fatalf("Read(corrupt) returned non-nil map: %v", loaded)
	}

	// Test 4: file written atomically (no temp file left behind)
	leftovers, _ := filepath.Glob(filepath.Join(dir, "versions-*.json"))
	if len(leftovers) != 0 {
		t.Fatalf("temp files not cleaned up: %v", leftovers)
	}
}

func TestWriteConcurrent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "versions.json")

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := Write(path, VersionMap{"claude": "1.0.0", "gemini": "2.0.0"}); err != nil {
				t.Errorf("Write() returned error: %v", err)
			}
		}()
	}
	wg.Wait()

	// The file must always be a complete, valid JSON object — never truncated.
	loaded, err := Read(path)
	if err != nil {
		t.Fatalf("Read() returned error: %v", err)
	}
	if loaded["claude"] != "1.0.0" || loaded["gemini"] != "2.0.0" {
		t.Fatalf("Read() = %v, want complete map", loaded)
	}

	leftovers, _ := filepath.Glob(filepath.Join(dir, "versions-*.json"))
	if len(leftovers) != 0 {
		t.Fatalf("temp files not cleaned up: %v", leftovers)
	}
}

func TestWriteFailureLeavesExistingIntact(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root bypasses directory permissions")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "versions.json")

	original := VersionMap{"claude": "1.0.0"}
	if err := Write(path, original); err != nil {
		t.Fatalf("Write() returned error: %v", err)
	}

	// Make the directory read-only so the temp-file creation fails.
	if err := os.Chmod(dir, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0755) })

	if err := Write(path, VersionMap{"claude": "2.0.0"}); err == nil {
		t.Fatal("expected Write() to fail on a read-only directory")
	}

	loaded, err := Read(path)
	if err != nil {
		t.Fatalf("Read() returned error: %v", err)
	}
	if loaded["claude"] != "1.0.0" {
		t.Errorf("existing file changed after failed write: got %q, want %q", loaded["claude"], "1.0.0")
	}
}

func TestNeedsRefresh(t *testing.T) {
	full := VersionMap{}
	for _, name := range AgentNames() {
		full[name] = "1.0.0"
	}
	partial := VersionMap{"claude": "1.0.0"} // missing other tracked agents

	t.Run("fresh and complete is not stale", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "versions.json")
		if err := Write(path, full); err != nil {
			t.Fatal(err)
		}
		stale, err := NeedsRefresh(path, full, 24*time.Hour)
		if err != nil {
			t.Fatal(err)
		}
		if stale {
			t.Error("NeedsRefresh() = true, want false for fresh complete map")
		}
	})

	t.Run("fresh but missing agent is stale", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "versions.json")
		if err := Write(path, partial); err != nil {
			t.Fatal(err)
		}
		stale, err := NeedsRefresh(path, partial, 24*time.Hour)
		if err != nil {
			t.Fatal(err)
		}
		if !stale {
			t.Error("NeedsRefresh() = false, want true when a tracked agent is missing")
		}
	})

	t.Run("old file is stale", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "versions.json")
		if err := Write(path, full); err != nil {
			t.Fatal(err)
		}
		old := time.Now().Add(-48 * time.Hour)
		if err := os.Chtimes(path, old, old); err != nil {
			t.Fatal(err)
		}
		stale, err := NeedsRefresh(path, full, 24*time.Hour)
		if err != nil {
			t.Fatal(err)
		}
		if !stale {
			t.Error("NeedsRefresh() = false, want true for a file older than the interval")
		}
	})
}

func TestWriteCreatesDir(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/subdir/versions.json"

	vm := VersionMap{"claude": "1.0.0"}
	if err := Write(path, vm); err != nil {
		t.Fatalf("Write() returned error: %v", err)
	}

	loaded, err := Read(path)
	if err != nil {
		t.Fatalf("Read() returned error: %v", err)
	}
	if loaded["claude"] != "1.0.0" {
		t.Errorf("Read() = %q, want %q", loaded["claude"], "1.0.0")
	}
}
