package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/inventage-ai/asylum/internal/kit"
)

func TestSyncKitToConfig_InsertsNewKit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	initial := "version: \"0.2\"\nkits:\n  docker: {} # existing\n"
	os.WriteFile(path, []byte(initial), 0644)

	nodes := []*yaml.Node{
		kit.ScalarNode("rust", "Rust toolchain"),
		kit.MappingNode(),
	}
	if err := SyncKitToConfig(path, "rust", nodes); err != nil {
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

	nodes := []*yaml.Node{
		kit.ScalarNode("docker", "should not duplicate"),
		kit.MappingNode(),
	}
	if err := SyncKitToConfig(path, "docker", nodes); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	if strings.Count(string(data), "docker") != 1 {
		t.Error("docker should appear exactly once (not duplicated)")
	}
}

func TestSyncKitToConfig_CreatesKitsMapping(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	initial := "version: \"0.2\"\nagent: claude\n"
	os.WriteFile(path, []byte(initial), 0644)

	nodes := []*yaml.Node{
		kit.ScalarNode("docker", ""),
		kit.MappingNode(),
	}
	if err := SyncKitToConfig(path, "docker", nodes); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	text := string(data)
	if !strings.Contains(text, "kits:") {
		t.Error("kits mapping should be created")
	}
	if !strings.Contains(text, "docker") {
		t.Error("docker should be added under kits")
	}
}

func TestSyncKitToConfig_PreservesComments(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	initial := "version: \"0.2\"\n# My custom comment\nkits:\n  docker: {} # Docker support\n"
	os.WriteFile(path, []byte(initial), 0644)

	nodes := []*yaml.Node{
		kit.ScalarNode("rust", ""),
		kit.MappingNode(),
	}
	if err := SyncKitToConfig(path, "rust", nodes); err != nil {
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
	combined := "apt:                # System packages\n\nbrowser:            # Chromium browser"
	if err := SyncKitCommentToConfig(path, combined); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	text := string(data)
	if !strings.Contains(text, "apt:") || !strings.Contains(text, "browser:") {
		t.Errorf("expected both commented kits in output, got:\n%s", text)
	}
}

func TestSyncKitToConfig_BlankLinesBetweenMultipleKits(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	initial := "version: \"0.2\"\nkits:\n  docker: {}\n"
	os.WriteFile(path, []byte(initial), 0644)

	// Add two kits sequentially (each call reads, modifies, writes)
	nodes1 := []*yaml.Node{kit.ScalarNode("rust", ""), kit.MappingNode()}
	if err := SyncKitToConfig(path, "rust", nodes1); err != nil {
		t.Fatal(err)
	}
	nodes2 := []*yaml.Node{kit.ScalarNode("python", ""), kit.MappingNode()}
	if err := SyncKitToConfig(path, "python", nodes2); err != nil {
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
	synced, err := SyncNewKits(dir, false)
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

	synced, err := SyncNewKits(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if synced {
		t.Error("expected no sync when all kits are known")
	}
}

func TestSyncNewKits_NoStateFile(t *testing.T) {
	dir := t.TempDir()

	configPath := filepath.Join(dir, "config.yaml")
	os.WriteFile(configPath, []byte("version: \"0.2\"\nkits:\n  docker: {}\n"), 0644)

	// No state.json — all kits are new (first run after feature lands)
	synced, err := SyncNewKits(dir, false)
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
