package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadState_Missing(t *testing.T) {
	dir := t.TempDir()
	s, err := LoadState(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.KnownKits) != 0 {
		t.Errorf("expected empty KnownKits, got %v", s.KnownKits)
	}
}

func TestLoadSaveState_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	original := State{KnownKits: []string{"docker", "java", "node"}}
	if err := SaveState(dir, original); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadState(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.KnownKits) != 3 {
		t.Fatalf("expected 3 kits, got %v", loaded.KnownKits)
	}
	for i, name := range []string{"docker", "java", "node"} {
		if loaded.KnownKits[i] != name {
			t.Errorf("KnownKits[%d] = %q, want %q", i, loaded.KnownKits[i], name)
		}
	}
}

func TestSaveState_SortsKits(t *testing.T) {
	dir := t.TempDir()
	if err := SaveState(dir, State{KnownKits: []string{"python", "apt", "java"}}); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadState(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"apt", "java", "python"}
	for i, name := range want {
		if loaded.KnownKits[i] != name {
			t.Errorf("KnownKits[%d] = %q, want %q", i, loaded.KnownKits[i], name)
		}
	}
}

func TestNewKits(t *testing.T) {
	state := State{KnownKits: []string{"docker", "java"}}
	registered := []string{"docker", "java", "rust", "zig"}
	got := NewKits(registered, state)
	if len(got) != 2 || got[0] != "rust" || got[1] != "zig" {
		t.Errorf("NewKits = %v, want [rust zig]", got)
	}
}

func TestNewKits_AllKnown(t *testing.T) {
	state := State{KnownKits: []string{"docker", "java"}}
	got := NewKits([]string{"docker", "java"}, state)
	if len(got) != 0 {
		t.Errorf("expected no new kits, got %v", got)
	}
}

func TestNewKits_DeletedState(t *testing.T) {
	// Empty state = all kits are new
	got := NewKits([]string{"docker", "java"}, State{})
	if len(got) != 2 {
		t.Errorf("expected 2 new kits, got %v", got)
	}
}

func TestState_ResumeMigrationPromptShown_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	if err := SaveState(dir, State{ResumeMigrationPromptShown: true}); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadState(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !loaded.ResumeMigrationPromptShown {
		t.Error("ResumeMigrationPromptShown lost in round-trip")
	}

	// Default (missing field) is false.
	loaded2, err := LoadState(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if loaded2.ResumeMigrationPromptShown {
		t.Error("missing field should default to false")
	}
}

func TestLoadState_CorruptFile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "state.json"), []byte("not json"), 0644)
	_, err := LoadState(dir)
	if err == nil {
		t.Fatal("expected error for corrupt state file")
	}
}
