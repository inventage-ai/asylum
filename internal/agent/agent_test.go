package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	for _, name := range []string{"claude", "gemini", "codex"} {
		a, err := Get(name)
		if err != nil {
			t.Errorf("Get(%q) error: %v", name, err)
			continue
		}
		if a.Name() != name {
			t.Errorf("Get(%q).Name() = %q", name, a.Name())
		}
	}

	_, err := Get("unknown")
	if err == nil {
		t.Error("Get(unknown) should return error")
	}
}

func TestClaudeCommand(t *testing.T) {
	a := Claude{}
	tests := []struct {
		name      string
		resume    bool
		extra     []string
		wantParts []string
	}{
		{"default with session", true, nil, []string{"claude", "--dangerously-skip-permissions", "--continue"}},
		{"default no session", false, nil, []string{"claude", "--dangerously-skip-permissions"}},
		{"new session", false, nil, []string{"claude", "--dangerously-skip-permissions"}},
		{"with args resume", true, []string{"fix", "the", "bug"}, []string{"claude", "--dangerously-skip-permissions", "--continue", "fix", "the", "bug"}},
		{"with args no resume", false, []string{"fix", "bug"}, []string{"claude", "--dangerously-skip-permissions", "fix", "bug"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := a.Command(tt.resume, tt.extra)
			assertZshWrapped(t, cmd, tt.wantParts)
		})
	}
}

func TestGeminiCommand(t *testing.T) {
	a := Gemini{}
	tests := []struct {
		name      string
		resume    bool
		extra     []string
		wantParts []string
	}{
		{"default with session", true, nil, []string{"gemini", "--yolo", "--resume"}},
		{"default no session", false, nil, []string{"gemini", "--yolo"}},
		{"with args", true, []string{"-p", "fix bug"}, []string{"gemini", "--yolo", "--resume", "-p", "fix bug"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := a.Command(tt.resume, tt.extra)
			assertZshWrapped(t, cmd, tt.wantParts)
		})
	}
}

func TestCodexCommand(t *testing.T) {
	a := Codex{}
	tests := []struct {
		name      string
		resume    bool
		extra     []string
		wantParts []string
	}{
		{"default with session", true, nil, []string{"codex", "resume", "--last", "--yolo"}},
		{"default no session", false, nil, []string{"codex", "--yolo"}},
		{"with args skips resume", true, []string{"fix", "bug"}, []string{"codex", "--yolo", "fix", "bug"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := a.Command(tt.resume, tt.extra)
			assertZshWrapped(t, cmd, tt.wantParts)
		})
	}
}

func assertZshWrapped(t *testing.T, cmd []string, wantParts []string) {
	t.Helper()
	if len(cmd) != 3 || cmd[0] != "zsh" || cmd[1] != "-c" {
		t.Fatalf("command not zsh-wrapped: %v", cmd)
	}
	inner := cmd[2]
	want := "source ~/.zshrc && exec " + strings.Join(wantParts, " ")
	if inner != want {
		t.Errorf("inner command:\n  got:  %q\n  want: %q", inner, want)
	}
}

func TestClaudeHasSession(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	a := Claude{}
	configDir := filepath.Join(dir, ".asylum", "agents", "claude", "projects")

	if a.HasSession("/some/project") {
		t.Error("should be false when projects dir doesn't exist")
	}

	os.MkdirAll(configDir, 0755)
	if a.HasSession("/some/project") {
		t.Error("should be false when projects dir is empty")
	}

	os.MkdirAll(filepath.Join(configDir, "abc123"), 0755)
	if !a.HasSession("/some/project") {
		t.Error("should be true when projects dir has entries")
	}
}

func TestGeminiHasSession(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	a := Gemini{}
	configDir := filepath.Join(dir, ".asylum", "agents", "gemini", "tmp")

	if a.HasSession("/some/project") {
		t.Error("should be false when tmp dir doesn't exist")
	}

	os.MkdirAll(configDir, 0755)
	if a.HasSession("/some/project") {
		t.Error("should be false when tmp dir is empty")
	}

	os.MkdirAll(filepath.Join(configDir, "hash123", "chats"), 0755)
	if !a.HasSession("/some/project") {
		t.Error("should be true when tmp dir has entries")
	}
}

func TestCodexHasSession(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	a := Codex{}
	configDir := filepath.Join(dir, ".asylum", "agents", "codex")

	if a.HasSession("/some/project") {
		t.Error("should be false when history.jsonl doesn't exist")
	}

	os.MkdirAll(configDir, 0755)
	historyFile := filepath.Join(configDir, "history.jsonl")

	os.WriteFile(historyFile, []byte(""), 0644)
	if a.HasSession("/some/project") {
		t.Error("should be false when history.jsonl is empty")
	}

	os.WriteFile(historyFile, []byte(`{"session": "data"}`), 0644)
	if !a.HasSession("/some/project") {
		t.Error("should be true when history.jsonl is non-empty")
	}
}
