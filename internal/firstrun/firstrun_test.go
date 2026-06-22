package firstrun

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsFirstRun(t *testing.T) {
	home := t.TempDir()
	if !IsFirstRun(home) {
		t.Fatal("expected first-run when config.yaml is absent")
	}

	if err := os.MkdirAll(filepath.Join(home, ".asylum"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".asylum", "config.yaml"), []byte("version: \"0.2\"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if IsFirstRun(home) {
		t.Fatal("expected not-first-run when config.yaml exists")
	}
}
