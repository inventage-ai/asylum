package versions

import (
	"os"
	"testing"
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

	// Test 4: file written atomically (temp file cleaned up)
	if _, err := os.Stat(path + ".tmp"); err == nil {
		t.Fatal("temp file not cleaned up")
	}
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
