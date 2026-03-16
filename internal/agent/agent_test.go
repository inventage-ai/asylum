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
	projectsDir := filepath.Join(dir, ".asylum", "agents", "claude", "projects")

	if a.HasSession("/some/project") {
		t.Error("should be false when projects dir doesn't exist")
	}

	// Directory for a different project — should not match
	os.MkdirAll(filepath.Join(projectsDir, "-other-project"), 0755)
	os.WriteFile(filepath.Join(projectsDir, "-other-project", "session.jsonl"), []byte("data"), 0644)
	if a.HasSession("/some/project") {
		t.Error("should be false when only a different project has sessions")
	}

	// Matching project directory but no .jsonl files
	matchDir := filepath.Join(projectsDir, "-some-project")
	os.MkdirAll(matchDir, 0755)
	if a.HasSession("/some/project") {
		t.Error("should be false when matching project dir has no .jsonl files")
	}

	// Matching project with a .jsonl file
	os.WriteFile(filepath.Join(matchDir, "abc123.jsonl"), []byte("data"), 0644)
	if !a.HasSession("/some/project") {
		t.Error("should be true when matching project dir has .jsonl files")
	}
}

func TestGeminiHasSession(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	a := Gemini{}
	tmpDir := filepath.Join(dir, ".asylum", "agents", "gemini", "tmp")

	if a.HasSession("/some/project") {
		t.Error("should be false when tmp dir doesn't exist")
	}

	// Slug dir with .project_root pointing to a different project
	otherDir := filepath.Join(tmpDir, "other")
	os.MkdirAll(filepath.Join(otherDir, "chats"), 0755)
	os.WriteFile(filepath.Join(otherDir, ".project_root"), []byte("/other/project\n"), 0644)
	os.WriteFile(filepath.Join(otherDir, "chats", "session.json"), []byte("{}"), 0644)
	if a.HasSession("/some/project") {
		t.Error("should be false when .project_root points elsewhere")
	}

	// Matching .project_root but empty chats/
	matchDir := filepath.Join(tmpDir, "project")
	os.MkdirAll(filepath.Join(matchDir, "chats"), 0755)
	os.WriteFile(filepath.Join(matchDir, ".project_root"), []byte("/some/project\n"), 0644)
	if a.HasSession("/some/project") {
		t.Error("should be false when chats dir is empty")
	}

	// Matching .project_root with chat files
	os.WriteFile(filepath.Join(matchDir, "chats", "session.json"), []byte("{}"), 0644)
	if !a.HasSession("/some/project") {
		t.Error("should be true when matching project has chat files")
	}
}

func TestCodexHasSession(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	a := Codex{}
	sessionsDir := filepath.Join(dir, ".asylum", "agents", "codex", "sessions")

	if a.HasSession("/some/project") {
		t.Error("should be false when sessions dir doesn't exist")
	}

	// Sessions dir exists but no rollout files
	os.MkdirAll(filepath.Join(sessionsDir, "2026", "03", "16"), 0755)
	if a.HasSession("/some/project") {
		t.Error("should be false when sessions dir has no rollout files")
	}

	// Add a rollout file
	os.WriteFile(
		filepath.Join(sessionsDir, "2026", "03", "16", "rollout-2026-03-16T14-30-00-abc123.jsonl"),
		[]byte("data"), 0644,
	)
	if !a.HasSession("/some/project") {
		t.Error("should be true when rollout files exist")
	}
}
