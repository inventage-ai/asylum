package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/inventage-ai/asylum/internal/kit"
)

func TestSyncKitToConfig_InsertsNewKit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	initial := "version: \"0.2\"\nkits:\n  docker: {} # existing\n"
	os.WriteFile(path, []byte(initial), 0644)

	if err := SyncKitToConfig(path, "rust", "  rust:               # Rust toolchain"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	text := string(data)

	// Existing content preserved
	if !strings.Contains(text, "docker") {
		t.Error("existing docker kit should be preserved")
	}
	// New kit added with blank line separator
	if !strings.Contains(text, "rust") {
		t.Error("rust kit should be inserted")
	}
	if !strings.Contains(text, "docker: {} # existing\n\n") {
		t.Errorf("expected blank line between kits, got:\n%s", text)
	}
}

func TestSyncKitToConfig_SkipsExistingKit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	initial := "version: \"0.2\"\nkits:\n  docker: {}\n"
	os.WriteFile(path, []byte(initial), 0644)

	if err := SyncKitToConfig(path, "docker", "  docker:"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	if strings.Count(string(data), "docker") != 1 {
		t.Error("docker should appear exactly once (not duplicated)")
	}
}

func TestSyncKitToConfig_PreservesComments(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	initial := "version: \"0.2\"\n# My custom comment\nkits:\n  docker: {} # Docker support\n"
	os.WriteFile(path, []byte(initial), 0644)

	if err := SyncKitToConfig(path, "rust", "  rust:"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	text := string(data)
	if !strings.Contains(text, "My custom comment") {
		t.Error("user comments should be preserved")
	}
	if !strings.Contains(text, "Docker support") {
		t.Error("line comments should be preserved")
	}
}

func TestSyncKitToConfig_PreservesIndentation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	initial := "version: \"0.2\"\nkits:\n  python:\n    # versions:\n    #   - 3.14\n\n  # apt:                # System packages\n  #   packages:\n  #     - imagemagick\n"
	os.WriteFile(path, []byte(initial), 0644)

	if err := SyncKitToConfig(path, "docker", "  docker:               # Docker-in-Docker support"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	text := string(data)

	// Original indentation must be preserved (2-space, not 4-space)
	if strings.Contains(text, "    python:") {
		t.Error("indentation should be preserved at 2 spaces, not changed to 4")
	}
	// Commented children of python must stay intact
	if !strings.Contains(text, "    # versions:") {
		t.Errorf("commented children should be preserved, got:\n%s", text)
	}
	// Commented-out kit section must stay intact
	if !strings.Contains(text, "  # apt:") {
		t.Errorf("commented-out kit entries should be preserved, got:\n%s", text)
	}
	// New kit inserted before commented section
	dockerIdx := strings.Index(text, "docker:")
	aptIdx := strings.Index(text, "# apt:")
	if dockerIdx < 0 || aptIdx < 0 || dockerIdx > aptIdx {
		t.Errorf("docker should be inserted before commented-out kits, got:\n%s", text)
	}
}

func TestSyncKitCommentToConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	initial := "version: \"0.2\"\nkits:\n  docker: {}\n"
	os.WriteFile(path, []byte(initial), 0644)

	if err := SyncKitCommentToConfig(path, "apt:                # System packages"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	text := string(data)
	if !strings.Contains(text, "# apt:") {
		t.Error("commented kit should appear in output")
	}
}

func TestSyncKitCommentToConfig_BlankLineBetweenComments(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	initial := "version: \"0.2\"\nkits:\n  docker: {}\n"
	os.WriteFile(path, []byte(initial), 0644)

	// Add two comments in one call to test the \n\n separator
	combined := "apt:                # System packages\n\nagent-browser:      # Browser automation via agent-browser"
	if err := SyncKitCommentToConfig(path, combined); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	text := string(data)
	if !strings.Contains(text, "apt:") || !strings.Contains(text, "agent-browser:") {
		t.Errorf("expected both commented kits in output, got:\n%s", text)
	}
}

func TestSyncKitToConfig_BlankLinesBetweenMultipleKits(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	initial := "version: \"0.2\"\nkits:\n  docker: {}\n"
	os.WriteFile(path, []byte(initial), 0644)

	// Add two kits sequentially (each call reads, modifies, writes)
	if err := SyncKitToConfig(path, "rust", "  rust:"); err != nil {
		t.Fatal(err)
	}
	if err := SyncKitToConfig(path, "python", "  python:"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	text := string(data)

	// Each kit should be separated by a blank line
	for _, name := range []string{"docker", "rust", "python"} {
		if !strings.Contains(text, name) {
			t.Errorf("expected %s in output", name)
		}
	}

	// Count blank lines within the kits block — should have 2 (between 3 kits)
	kitsIdx := strings.Index(text, "kits:")
	if kitsIdx < 0 {
		t.Fatal("kits: not found")
	}
	kitsBlock := text[kitsIdx:]
	blankLines := strings.Count(kitsBlock, "\n\n")
	if blankLines < 2 {
		t.Errorf("expected at least 2 blank line separators between 3 kits, got %d in:\n%s", blankLines, text)
	}
}

func TestSyncNewKits_NonInteractive(t *testing.T) {
	dir := t.TempDir()

	// Create config with existing kits
	configPath := filepath.Join(dir, "config.yaml")
	os.WriteFile(configPath, []byte("version: \"0.2\"\nkits:\n  docker: {}\n"), 0644)

	// Create state with only "docker" known
	SaveState(dir, State{KnownKits: []string{"docker"}})

	// SyncNewKits should detect all other registered kits as new.
	// Non-interactive mode: TierDefault kits added as comments, not active.
	synced, err := SyncNewKits(dir, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !synced {
		t.Error("expected sync to process new kits")
	}

	// State should now contain all registered kits
	state, err := LoadState(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(state.KnownKits) < 2 {
		t.Errorf("expected state to contain all registered kits, got %v", state.KnownKits)
	}

	// Config should still parse correctly
	data, _ := os.ReadFile(configPath)
	text := string(data)
	if !strings.Contains(text, "docker") {
		t.Error("existing docker kit should be preserved")
	}
}

func TestSyncNewKits_AllKnown(t *testing.T) {
	dir := t.TempDir()

	configPath := filepath.Join(dir, "config.yaml")
	os.WriteFile(configPath, []byte("version: \"0.2\"\nkits:\n  docker: {}\n"), 0644)

	// State already knows all kits
	SaveState(dir, State{KnownKits: kit.All()})

	synced, err := SyncNewKits(dir, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	if synced {
		t.Error("expected no sync when all kits are known")
	}
}

func TestSyncNewKits_NoConfigFile(t *testing.T) {
	dir := t.TempDir()

	// No config.yaml, no state.json — simulates upgrade from pre-config version.
	// Should silently mark all kits as known without showing messages.
	synced, err := SyncNewKits(dir, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	if synced {
		t.Error("expected no sync when config.yaml doesn't exist")
	}

	// State should be populated with all kits so next run doesn't re-prompt
	state, _ := LoadState(dir)
	if len(state.KnownKits) != len(kit.All()) {
		t.Errorf("expected state to contain all %d kits, got %d", len(kit.All()), len(state.KnownKits))
	}
}

func TestSyncNewKits_NoStateFile(t *testing.T) {
	dir := t.TempDir()

	configPath := filepath.Join(dir, "config.yaml")
	os.WriteFile(configPath, []byte("version: \"0.2\"\nkits:\n  docker: {}\n"), 0644)

	// No state.json — all kits are new (first run after feature lands)
	synced, err := SyncNewKits(dir, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !synced {
		t.Error("expected sync when state file is missing")
	}

	// State should now exist with all kits
	state, _ := LoadState(dir)
	if len(state.KnownKits) == 0 {
		t.Error("state should be populated after sync")
	}
}

func TestRemoveKitComment_WithOptions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	initial := `version: "0.2"
kits:
  docker: {}

  # python:
  #   versions:
  #     - 3.14
  # packages:
  #   - ansible
`
	os.WriteFile(path, []byte(initial), 0644)

	if err := RemoveKitComment(path, "python"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	text := string(data)

	if strings.Contains(text, "python") {
		t.Errorf("commented python block should be removed, got:\n%s", text)
	}
	if !strings.Contains(text, "docker") {
		t.Error("docker should be preserved")
	}
}

func TestRemoveKitComment_NotPresent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	initial := "version: \"0.2\"\nkits:\n  docker: {}\n"
	os.WriteFile(path, []byte(initial), 0644)

	if err := RemoveKitComment(path, "python"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	if string(data) != initial {
		t.Errorf("file should be unchanged, got:\n%s", string(data))
	}
}

func TestRemoveKitComment_SingleLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	initial := `version: "0.2"
kits:
  docker: {}

  # apt:                # System packages
`
	os.WriteFile(path, []byte(initial), 0644)

	if err := RemoveKitComment(path, "apt"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	text := string(data)

	if strings.Contains(text, "apt") {
		t.Errorf("commented apt line should be removed, got:\n%s", text)
	}
}

func TestSetKitDisabled(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	initial := "version: \"0.2\"\nkits:\n  ast-grep:\n  java:\n    versions:\n      - 17\n"
	os.WriteFile(path, []byte(initial), 0644)

	if err := SetKitDisabled(path, "ast-grep"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	text := string(data)

	if !strings.Contains(text, "  ast-grep:\n    disabled: true\n") {
		t.Errorf("expected disabled: true under ast-grep, got:\n%s", text)
	}
	if !strings.Contains(text, "java:") {
		t.Error("java should be preserved")
	}
}

func TestSetKitDisabled_WithExistingProperties(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	initial := "version: \"0.2\"\nkits:\n  apt:\n    packages:\n      - imagemagick\n"
	os.WriteFile(path, []byte(initial), 0644)

	if err := SetKitDisabled(path, "apt"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	text := string(data)

	if !strings.Contains(text, "  apt:\n    disabled: true\n    packages:\n") {
		t.Errorf("disabled should be first property, got:\n%s", text)
	}
}

func TestSetKitDisabled_AlreadyDisabled(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	initial := "version: \"0.2\"\nkits:\n  ast-grep:\n    disabled: true\n"
	os.WriteFile(path, []byte(initial), 0644)

	if err := SetKitDisabled(path, "ast-grep"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	if string(data) != initial {
		t.Errorf("file should be unchanged, got:\n%s", string(data))
	}
}

func TestSetKitDisabled_OverridesDisabledFalse(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	initial := "version: \"0.2\"\nkits:\n  ast-grep:\n    disabled: false\n    packages:\n      - foo\n"
	os.WriteFile(path, []byte(initial), 0644)

	if err := SetKitDisabled(path, "ast-grep"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	text := string(data)

	if !strings.Contains(text, "disabled: true") {
		t.Errorf("disabled: false should become disabled: true, got:\n%s", text)
	}
	if strings.Contains(text, "disabled: false") {
		t.Errorf("disabled: false should be gone, got:\n%s", text)
	}
	if !strings.Contains(text, "packages:") {
		t.Errorf("other properties should be preserved, got:\n%s", text)
	}
}

func TestRemoveKitDisabled(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	initial := "version: \"0.2\"\nkits:\n  ast-grep:\n    disabled: true\n  java:\n"
	os.WriteFile(path, []byte(initial), 0644)

	if err := RemoveKitDisabled(path, "ast-grep"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	text := string(data)

	if strings.Contains(text, "disabled") {
		t.Errorf("disabled line should be removed, got:\n%s", text)
	}
	if !strings.Contains(text, "ast-grep:") {
		t.Error("ast-grep entry should remain")
	}
	if !strings.Contains(text, "java:") {
		t.Error("java should be preserved")
	}
}

func TestRemoveKitDisabled_NotPresent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	initial := "version: \"0.2\"\nkits:\n  ast-grep:\n"
	os.WriteFile(path, []byte(initial), 0644)

	if err := RemoveKitDisabled(path, "ast-grep"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	if string(data) != initial {
		t.Errorf("file should be unchanged, got:\n%s", string(data))
	}
}

func TestRemoveKitDisabled_PreservesOtherProperties(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	initial := "version: \"0.2\"\nkits:\n  apt:\n    disabled: true\n    packages:\n      - imagemagick\n"
	os.WriteFile(path, []byte(initial), 0644)

	if err := RemoveKitDisabled(path, "apt"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	text := string(data)

	if strings.Contains(text, "disabled") {
		t.Errorf("disabled line should be removed, got:\n%s", text)
	}
	if !strings.Contains(text, "packages:") {
		t.Errorf("packages should be preserved, got:\n%s", text)
	}
}

func TestKitExistsInFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	initial := "version: \"0.2\"\nkits:\n  docker: {}\n\n  # ast-grep:\n"
	os.WriteFile(path, []byte(initial), 0644)

	if !KitExistsInFile(path, "docker") {
		t.Error("docker should exist as active entry")
	}
	if KitExistsInFile(path, "ast-grep") {
		t.Error("ast-grep is a comment, should not be found")
	}
	if KitExistsInFile(path, "python") {
		t.Error("python is absent, should not be found")
	}
}

