package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResumeByDefault(t *testing.T) {
	t.Run("unset is false", func(t *testing.T) {
		if (Config{}).ResumeByDefault() {
			t.Error("unset should resolve to false")
		}
	})
	t.Run("explicit false", func(t *testing.T) {
		f := false
		if (Config{DefaultResume: &f}).ResumeByDefault() {
			t.Error("explicit false should resolve to false")
		}
	})
	t.Run("explicit true", func(t *testing.T) {
		tr := true
		if !(Config{DefaultResume: &tr}).ResumeByDefault() {
			t.Error("explicit true should resolve to true")
		}
	})
}

func TestMergeDefaultResume(t *testing.T) {
	tr, f := true, false
	tests := []struct {
		name string
		base Config
		over Config
		want bool
	}{
		{"unset stays unset", Config{}, Config{}, false},
		{"global true, no override", Config{DefaultResume: &tr}, Config{}, true},
		{"project overrides global with false", Config{DefaultResume: &tr}, Config{DefaultResume: &f}, false},
		{"local overrides project with true", Config{DefaultResume: &f}, Config{DefaultResume: &tr}, true},
		{"empty overlay keeps base true", Config{DefaultResume: &tr}, Config{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Merge(tt.base, tt.over).ResumeByDefault()
			if got != tt.want {
				t.Errorf("ResumeByDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWriteDefaultResume_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	if err := WriteDefaultResume(dir, true); err != nil {
		t.Fatalf("WriteDefaultResume: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "default-resume: true") {
		t.Errorf("config.yaml = %q, want it to contain `default-resume: true`", data)
	}
}

func TestWriteDefaultResume_PreservesOtherKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	existing := "agent: gemini\nports:\n  - \"3000\"\n"
	if err := os.WriteFile(path, []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	if err := WriteDefaultResume(dir, true); err != nil {
		t.Fatalf("WriteDefaultResume: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	for _, want := range []string{"agent: gemini", "3000", "default-resume: true"} {
		if !strings.Contains(got, want) {
			t.Errorf("config.yaml missing %q in:\n%s", want, got)
		}
	}

	// Round-trip via LoadFile to ensure structure is intact.
	loaded, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	if loaded.Agent != "gemini" {
		t.Errorf("agent = %q, want gemini", loaded.Agent)
	}
	if !loaded.ResumeByDefault() {
		t.Error("ResumeByDefault() = false, want true")
	}
}

func TestWriteDefaultResume_UpdatesExistingKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("default-resume: true\nagent: claude\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := WriteDefaultResume(dir, false); err != nil {
		t.Fatalf("WriteDefaultResume: %v", err)
	}

	loaded, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	if loaded.ResumeByDefault() {
		t.Error("ResumeByDefault() = true after writing false")
	}
	if loaded.Agent != "claude" {
		t.Errorf("agent = %q, want claude", loaded.Agent)
	}
}

func TestWriteDefaultResume_RefusesNonMapping(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("not: [valid: yaml\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := WriteDefaultResume(dir, true); err == nil {
		t.Error("expected error for unparseable config.yaml")
	}
}
