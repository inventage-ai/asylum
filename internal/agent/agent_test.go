package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveConfigDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	a := Claude{}
	tests := []struct {
		isolation string
		cname     string
		want      string
	}{
		{"shared", "asylum-abc123-proj", filepath.Join(dir, ".claude")},
		{"project", "asylum-abc123-proj", filepath.Join(dir, ".asylum", "projects", "asylum-abc123-proj", "claude-config")},
		{"isolated", "asylum-abc123-proj", filepath.Join(dir, ".asylum", "agents", "claude")},
		{"", "asylum-abc123-proj", filepath.Join(dir, ".asylum", "agents", "claude")},
	}
	for _, tt := range tests {
		t.Run("isolation="+tt.isolation, func(t *testing.T) {
			got, err := ResolveConfigDir(a, tt.isolation, tt.cname)
			if err != nil {
				t.Fatalf("ResolveConfigDir: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGet(t *testing.T) {
	for _, name := range []string{"claude", "codex", "echo", "gemini", "opencode"} {
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
		{"with args resume", true, []string{"fix", "the", "bug"}, []string{"claude", "--dangerously-skip-permissions", "--continue", "'fix'", "'the'", "'bug'"}},
		{"with args no resume", false, []string{"fix", "bug"}, []string{"claude", "--dangerously-skip-permissions", "'fix'", "'bug'"}},
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
		{"with args", true, []string{"-p", "fix bug"}, []string{"gemini", "--yolo", "--resume", "'-p'", "'fix bug'"}},
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
		{"with args skips resume", true, []string{"fix", "bug"}, []string{"codex", "--yolo", "'fix'", "'bug'"}},
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

	a := Claude{}
	configDir := filepath.Join(dir, "config")

	if a.HasSession(configDir, "/some/project") {
		t.Error("should be false when projects dir doesn't exist")
	}

	projectsDir := filepath.Join(configDir, "projects")

	// Directory for a different project — should not match
	os.MkdirAll(filepath.Join(projectsDir, "-other-project"), 0755)
	os.WriteFile(filepath.Join(projectsDir, "-other-project", "session.jsonl"), []byte("data"), 0644)
	if a.HasSession(configDir, "/some/project") {
		t.Error("should be false when only a different project has sessions")
	}

	// Matching project directory but no .jsonl files
	matchDir := filepath.Join(projectsDir, "-some-project")
	os.MkdirAll(matchDir, 0755)
	if a.HasSession(configDir, "/some/project") {
		t.Error("should be false when matching project dir has no .jsonl files")
	}

	// Matching project with a .jsonl file
	os.WriteFile(filepath.Join(matchDir, "abc123.jsonl"), []byte("data"), 0644)
	if !a.HasSession(configDir, "/some/project") {
		t.Error("should be true when matching project dir has .jsonl files")
	}
}

func TestGeminiHasSession(t *testing.T) {
	dir := t.TempDir()

	a := Gemini{}
	configDir := filepath.Join(dir, "config")
	tmpDir := filepath.Join(configDir, "tmp")

	if a.HasSession(configDir, "/some/project") {
		t.Error("should be false when tmp dir doesn't exist")
	}

	// Slug dir with .project_root pointing to a different project
	otherDir := filepath.Join(tmpDir, "other")
	os.MkdirAll(filepath.Join(otherDir, "chats"), 0755)
	os.WriteFile(filepath.Join(otherDir, ".project_root"), []byte("/other/project\n"), 0644)
	os.WriteFile(filepath.Join(otherDir, "chats", "session.json"), []byte("{}"), 0644)
	if a.HasSession(configDir, "/some/project") {
		t.Error("should be false when .project_root points elsewhere")
	}

	// Matching .project_root but empty chats/
	matchDir := filepath.Join(tmpDir, "project")
	os.MkdirAll(filepath.Join(matchDir, "chats"), 0755)
	os.WriteFile(filepath.Join(matchDir, ".project_root"), []byte("/some/project\n"), 0644)
	if a.HasSession(configDir, "/some/project") {
		t.Error("should be false when chats dir is empty")
	}

	// Matching .project_root with chat files
	os.WriteFile(filepath.Join(matchDir, "chats", "session.json"), []byte("{}"), 0644)
	if !a.HasSession(configDir, "/some/project") {
		t.Error("should be true when matching project has chat files")
	}
}

func TestGeminiHasSessionContinuesAfterReadDirError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root ignores permission bits")
	}
	dir := t.TempDir()

	a := Gemini{}
	configDir := filepath.Join(dir, "config")
	tmpDir := filepath.Join(configDir, "tmp")

	// First dir: matches project but chats/ is unreadable
	badDir := filepath.Join(tmpDir, "bad")
	os.MkdirAll(filepath.Join(badDir, "chats"), 0755)
	os.WriteFile(filepath.Join(badDir, ".project_root"), []byte("/some/project\n"), 0644)
	os.Chmod(filepath.Join(badDir, "chats"), 0000)
	defer os.Chmod(filepath.Join(badDir, "chats"), 0755)

	// Second dir: also matches, chats/ is readable and has files
	goodDir := filepath.Join(tmpDir, "good")
	os.MkdirAll(filepath.Join(goodDir, "chats"), 0755)
	os.WriteFile(filepath.Join(goodDir, ".project_root"), []byte("/some/project\n"), 0644)
	os.WriteFile(filepath.Join(goodDir, "chats", "session.json"), []byte("{}"), 0644)

	if !a.HasSession(configDir, "/some/project") {
		t.Error("should be true: ReadDir error on first match should not stop scan")
	}
}

func TestCodexHasSession(t *testing.T) {
	dir := t.TempDir()

	a := Codex{}
	configDir := filepath.Join(dir, "config")
	projectsDir := filepath.Join(configDir, "projects")

	if a.HasSession(configDir, "/some/project") {
		t.Error("should be false when marker does not exist")
	}

	// Marker for a different project — should not match
	os.MkdirAll(filepath.Join(projectsDir, "-other-project"), 0755)
	os.WriteFile(filepath.Join(projectsDir, "-other-project", ".has_session"), []byte(""), 0644)
	if a.HasSession(configDir, "/some/project") {
		t.Error("should be false when only a different project has a marker")
	}

	// Marker for this project
	os.MkdirAll(filepath.Join(projectsDir, "-some-project"), 0755)
	os.WriteFile(filepath.Join(projectsDir, "-some-project", ".has_session"), []byte(""), 0644)
	if !a.HasSession(configDir, "/some/project") {
		t.Error("should be true when marker exists for this project")
	}
}

func TestEchoCommand(t *testing.T) {
	a := Echo{}
	t.Run("with args", func(t *testing.T) {
		cmd := a.Command(false, []string{"hello", "world"})
		if len(cmd) != 3 || cmd[0] != "echo" || cmd[1] != "hello" || cmd[2] != "world" {
			t.Errorf("got %v, want [echo hello world]", cmd)
		}
	})
	t.Run("without args", func(t *testing.T) {
		cmd := a.Command(false, nil)
		if len(cmd) != 1 || cmd[0] != "echo" {
			t.Errorf("got %v, want [echo]", cmd)
		}
	})
	t.Run("resume ignored", func(t *testing.T) {
		cmd := a.Command(true, []string{"test"})
		if len(cmd) != 2 || cmd[0] != "echo" {
			t.Errorf("resume should be ignored, got %v", cmd)
		}
	})
}

func TestCodexWriteMarker(t *testing.T) {
	dir := t.TempDir()

	a := Codex{}
	configDir := filepath.Join(dir, "config")

	if a.HasSession(configDir, "/some/project") {
		t.Fatal("precondition: should be false before WriteMarker")
	}

	if err := a.WriteMarker(configDir, "/some/project"); err != nil {
		t.Fatalf("WriteMarker: %v", err)
	}

	if !a.HasSession(configDir, "/some/project") {
		t.Error("should be true after WriteMarker")
	}

	// Writing again should be idempotent
	if err := a.WriteMarker(configDir, "/some/project"); err != nil {
		t.Errorf("WriteMarker second call: %v", err)
	}
}
