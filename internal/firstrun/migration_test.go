package firstrun

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/inventage-ai/asylum/internal/config"
)

// MaybeShowResumeMigrationPrompt's interactive branch needs a TTY. These
// tests cover the non-interactive branches (which are the ones that exist
// regardless of how the dialog is rendered) and the new-user pre-mark path,
// since `tui.Select` returns the default index without prompting when not a
// terminal — which is what `term.IsTerminal()` reports in `go test`.

func TestIsExistingInstall(t *testing.T) {
	t.Run("fresh home is new", func(t *testing.T) {
		if IsExistingInstall(t.TempDir()) {
			t.Error("fresh home should be reported as new")
		}
	})

	t.Run("only ~/.asylum/config.yaml is not enough", func(t *testing.T) {
		home := t.TempDir()
		// Default config is written eagerly on every invocation — it must NOT
		// flip the existing-install probe, otherwise brand-new users would be
		// shown the migration dialog.
		if err := os.MkdirAll(filepath.Join(home, ".asylum"), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(home, ".asylum", "config.yaml"), []byte("agent: claude\n"), 0644); err != nil {
			t.Fatal(err)
		}
		if IsExistingInstall(home) {
			t.Error("home with only ~/.asylum/config.yaml should still be reported as new")
		}
	})

	t.Run("~/.asylum/agents/ flips the signal", func(t *testing.T) {
		home := t.TempDir()
		if err := os.MkdirAll(filepath.Join(home, ".asylum", "agents"), 0755); err != nil {
			t.Fatal(err)
		}
		if !IsExistingInstall(home) {
			t.Error("home with ~/.asylum/agents/ should be reported as existing")
		}
	})
}

func TestMaybeShowResumeMigrationPrompt_NewUserPreMarks(t *testing.T) {
	dir := t.TempDir()
	state := config.State{}
	cfg := config.Config{}

	decision, dirty := MaybeShowResumeMigrationPrompt(dir, &state, &cfg, true, false)
	if decision != ResumePromptSkipped {
		t.Errorf("decision = %v, want skipped", decision)
	}
	if !dirty {
		t.Error("state should be marked dirty for a new user")
	}
	if !state.ResumeMigrationPromptShown {
		t.Error("new user should have ResumeMigrationPromptShown set")
	}
	if cfg.DefaultResume != nil {
		t.Errorf("cfg.DefaultResume = %v, want nil", *cfg.DefaultResume)
	}
}

func TestMaybeShowResumeMigrationPrompt_FlagAlreadySetIsNoop(t *testing.T) {
	dir := t.TempDir()
	state := config.State{ResumeMigrationPromptShown: true}
	cfg := config.Config{}

	decision, dirty := MaybeShowResumeMigrationPrompt(dir, &state, &cfg, true, true)
	if decision != ResumePromptSkipped {
		t.Errorf("decision = %v, want skipped", decision)
	}
	if dirty {
		t.Error("state should not be dirty when flag already set")
	}
}

func TestMaybeShowResumeMigrationPrompt_NonAgentModeSuppressed(t *testing.T) {
	dir := t.TempDir()
	state := config.State{}
	cfg := config.Config{}

	decision, dirty := MaybeShowResumeMigrationPrompt(dir, &state, &cfg, false, true)
	if decision != ResumePromptSkipped {
		t.Errorf("decision = %v, want skipped", decision)
	}
	if dirty {
		t.Error("non-agent mode should not mark state dirty")
	}
	if state.ResumeMigrationPromptShown {
		t.Error("non-agent mode should not flip the flag")
	}
}

func TestMaybeShowResumeMigrationPrompt_NonTTYExistingUserSuppressed(t *testing.T) {
	dir := t.TempDir()
	state := config.State{}
	cfg := config.Config{}

	// In `go test` there is no TTY, so existing-user + agent-mode falls into
	// the "suppress without marking" branch.
	decision, dirty := MaybeShowResumeMigrationPrompt(dir, &state, &cfg, true, true)
	if decision != ResumePromptSkipped {
		t.Errorf("decision = %v, want skipped", decision)
	}
	if dirty {
		t.Error("non-TTY existing-user should not mark state dirty")
	}
	if state.ResumeMigrationPromptShown {
		t.Error("non-TTY suppression must NOT flip the flag — next interactive run should still prompt")
	}

	// And nothing was written to disk.
	if _, err := os.Stat(filepath.Join(dir, "config.yaml")); err == nil {
		t.Error("config.yaml should not be created by suppressed dialog")
	}
}

// TestMaybeShowResumeMigrationPrompt_OptInWritesConfig covers the legacy-opt-in
// path by calling WriteDefaultResume directly via the same helper the dialog
// invokes — there is no piped-TTY in `go test`, but we still want a black-box
// assertion that the end state matches what the prompt produces. This is the
// integration-style coverage referenced in tasks 4.5 and 4.8.
func TestMaybeShowResumeMigrationPrompt_OptInWritesConfig(t *testing.T) {
	dir := t.TempDir()

	if err := config.WriteDefaultResume(dir, true); err != nil {
		t.Fatalf("WriteDefaultResume: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "default-resume: true") {
		t.Errorf("config.yaml does not contain `default-resume: true`:\n%s", data)
	}

	loaded, err := config.LoadFile(filepath.Join(dir, "config.yaml"))
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	if !loaded.ResumeByDefault() {
		t.Error("ResumeByDefault() = false after dialog-equivalent write")
	}
}
