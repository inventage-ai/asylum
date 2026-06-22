package firstrun

import (
	"os"
	"path/filepath"

	"github.com/inventage-ai/asylum/internal/config"
	"github.com/inventage-ai/asylum/internal/log"
	"github.com/inventage-ai/asylum/internal/term"
	"github.com/inventage-ai/asylum/internal/tui"
)

// IsExistingInstall reports whether the user has run asylum before. The signal
// is `<home>/.asylum/agents/`, the same directory the first-run code uses —
// it's created only when an agent's config is materialised
// (`container.EnsureAgentConfig`), not by the early `~/.asylum/config.yaml`
// write that happens on every invocation. That makes it a reliable
// existing-vs-new probe even when called after default-config initialisation.
func IsExistingInstall(home string) bool {
	_, err := os.Stat(filepath.Join(home, ".asylum", "agents"))
	return err == nil
}

// ResumePromptDecision is the outcome of the one-time resume-migration dialog.
type ResumePromptDecision int

const (
	// ResumePromptSkipped means the dialog was not shown (new user, flag already
	// set, non-agent mode, or no TTY). The caller does not need to take any
	// session-related action.
	ResumePromptSkipped ResumePromptDecision = iota
	// ResumePromptKeptNewDefault means the dialog was shown and the user
	// elected to keep the new "start a new session by default" behaviour.
	ResumePromptKeptNewDefault
	// ResumePromptOptedIntoLegacy means the dialog was shown, the user opted
	// back into auto-resume, and `default-resume: true` has been written to
	// the global config file.
	ResumePromptOptedIntoLegacy
)

// MaybeShowResumeMigrationPrompt runs the one-time upgrade dialog for users
// who had asylum installed before the default-new-session behaviour change.
//
// It is a no-op when any of the following hold:
//   - The dialog has already been shown (`state.ResumeMigrationPromptShown`).
//   - This is not an agent-mode invocation (`agentMode == false`).
//   - stdin/stdout is not a terminal.
//   - This is a brand-new install (`existingInstall == false`). In that case
//     the prompt-shown flag is set immediately so future asylum versions also
//     skip the dialog.
//
// On a shown-and-decided run, the state is updated in place and the caller is
// expected to persist it via `config.SaveState`. When the user opts into
// legacy behaviour, `default-resume: true` is written to `<asylumDir>/config.yaml`
// and the resolved config is mutated in place so the current invocation honours
// the new value.
//
// Returns the outcome and whether `state` was mutated and needs saving.
func MaybeShowResumeMigrationPrompt(
	asylumDir string,
	state *config.State,
	cfg *config.Config,
	agentMode bool,
	existingInstall bool,
) (ResumePromptDecision, bool) {
	if state.ResumeMigrationPromptShown {
		return ResumePromptSkipped, false
	}

	// Pre-mark new users so they never see the dialog on a future asylum
	// version. Detection must run before any code creates ~/.asylum/.
	if !existingInstall {
		state.ResumeMigrationPromptShown = true
		return ResumePromptSkipped, true
	}

	if !agentMode || !term.IsTerminal() {
		// Suppress without marking — the next interactive agent run should
		// still surface the dialog.
		return ResumePromptSkipped, false
	}

	idx, err := tui.Select(
		"Asylum default behaviour has changed: each `asylum` invocation now starts a new session.\n"+
			"Use --continue or --resume (forwarded to the agent) to resume the previous session.\n"+
			"You can restore the old auto-resume behaviour as a one-time choice below.",
		[]tui.Option{
			{
				Label:       "Keep the new default (start a new session)",
				Description: "Pass --continue or --resume when you want to resume.",
			},
			{
				Label:       "Restore previous behaviour (auto-resume)",
				Description: "Writes `default-resume: true` to ~/.asylum/config.yaml.",
			},
		},
		0,
	)
	if err != nil {
		log.Warn("resume-migration prompt: %v", err)
		return ResumePromptSkipped, false
	}

	if idx != 1 {
		state.ResumeMigrationPromptShown = true
		return ResumePromptKeptNewDefault, true
	}

	// Only mark the prompt as shown when the opt-in actually persists. If the
	// config write fails (read-only filesystem, full disk, parse error on an
	// existing file), leave the flag clear so the next interactive run shows
	// the dialog again — silently switching the user to the new default
	// against their stated preference would be worse than re-prompting.
	if err := config.WriteDefaultResume(asylumDir, true); err != nil {
		log.Error("write default-resume: %v — will re-prompt on next run", err)
		return ResumePromptSkipped, false
	}
	state.ResumeMigrationPromptShown = true
	tr := true
	cfg.DefaultResume = &tr
	return ResumePromptOptedIntoLegacy, true
}
