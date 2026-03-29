package firstrun

import (
	"os"
	"path/filepath"
	"strings"
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

func TestSetKitCredentials(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	// Start with a config that has a kits section
	initial := `version: "2"
kits:
  java:
    versions:
      - 17
      - 21
`
	os.WriteFile(cfgPath, []byte(initial), 0644)

	if err := SetKitCredentials(cfgPath, "java", "auto"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, "credentials: auto") {
		t.Errorf("expected 'credentials: auto' in config, got:\n%s", content)
	}
	// Verify existing content preserved
	if !strings.Contains(content, "versions:") {
		t.Errorf("existing content should be preserved, got:\n%s", content)
	}
}

func TestSetKitCredentials_NewKit(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	initial := `version: "2"
kits:
  node:
`
	os.WriteFile(cfgPath, []byte(initial), 0644)

	if err := SetKitCredentials(cfgPath, "java", "auto"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, "java:") {
		t.Errorf("expected java kit added, got:\n%s", content)
	}
	if !strings.Contains(content, "credentials: auto") {
		t.Errorf("expected credentials: auto, got:\n%s", content)
	}
}

func TestSetKitCredentials_NoKitsSection(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	initial := `version: "2"
`
	os.WriteFile(cfgPath, []byte(initial), 0644)

	if err := SetKitCredentials(cfgPath, "java", "auto"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, "kits:") {
		t.Errorf("expected kits section added, got:\n%s", content)
	}
	if !strings.Contains(content, "credentials: auto") {
		t.Errorf("expected credentials: auto, got:\n%s", content)
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
