package firstrun

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunSkipsExistingUser(t *testing.T) {
	home := t.TempDir()
	// Simulate existing user: agents/ directory exists
	os.MkdirAll(filepath.Join(home, ".asylum", "agents"), 0755)

	if err := Run(home); err != nil {
		t.Fatal(err)
	}
	// Should not create config for existing user
	if _, err := os.Stat(filepath.Join(home, ".asylum", "config.yaml")); err == nil {
		t.Fatal("config.yaml should not exist for existing user")
	}
}

func TestRunNewUser(t *testing.T) {
	home := t.TempDir()
	// No agents dir — this is a new user, but Run should succeed
	// (credential prompting moved to onboarding wizard)
	if err := Run(home); err != nil {
		t.Fatal(err)
	}
}
