package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCopilotBasics(t *testing.T) {
	c := Copilot{}
	if c.Name() != "copilot" {
		t.Fatalf("Name mismatch: %s", c.Name())
	}
	if c.Binary() != "copilot" {
		t.Fatalf("Binary mismatch: %s", c.Binary())
	}
	if c.NativeConfigDir() == "" {
		t.Fatal("NativeConfigDir empty")
	}
}

func TestHasSessionDetection(t *testing.T) {
	tmp := t.TempDir()
	c := Copilot{}
	const projA = "/work/project-a"

	if c.HasSession(tmp, projA) {
		t.Fatal("expected no session before WriteMarker")
	}
	if err := c.WriteMarker(tmp, projA); err != nil {
		t.Fatalf("WriteMarker: %v", err)
	}
	if !c.HasSession(tmp, projA) {
		t.Fatal("expected session after WriteMarker for same project")
	}
}

// Regression for cross-project resume leak: a marker written for project A
// must not make HasSession return true for project B even when both share the
// same config dir (the default in shared isolation).
func TestHasSessionIsProjectScoped(t *testing.T) {
	tmp := t.TempDir()
	c := Copilot{}
	const projA = "/work/project-a"
	const projB = "/work/project-b"

	// Simulate copilot having existing global session-state from another project.
	sessionDir := filepath.Join(tmp, "session-state", "session-from-project-a")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := c.WriteMarker(tmp, projA); err != nil {
		t.Fatalf("WriteMarker: %v", err)
	}

	// project-a was launched before → resume allowed
	if !c.HasSession(tmp, projA) {
		t.Error("project A should report a session")
	}
	// project-b shares the same configDir but has never been launched → no resume
	if c.HasSession(tmp, projB) {
		t.Error("project B must not inherit project A's session marker")
	}
}

func TestCommandResumeAndEnv(t *testing.T) {
	cmd := (Copilot{}).Command(true, []string{"--banner"}, CmdOpts{})
	if len(cmd) < 3 {
		t.Fatal("unexpected command structure")
	}
	joined := cmd[2]
	if !strings.Contains(joined, "--resume") {
		t.Fatalf("expected --resume in command, got %q", joined)
	}
	if !strings.Contains(joined, "GH_TOKEN") {
		t.Fatalf("expected GH_TOKEN export in command wrapper, got %q", joined)
	}
}
