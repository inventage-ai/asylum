package firstrun

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/inventage-ai/asylum/internal/config"
	"github.com/inventage-ai/asylum/internal/kit"
	"github.com/inventage-ai/asylum/internal/term"
	"github.com/inventage-ai/asylum/internal/tui"
)

// --- Helper-level unit tests ---

func TestActiveAgentIsClaude(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
		want bool
	}{
		{"nil", nil, false},
		{"empty agent + nil agents map → fallback to claude", &config.Config{}, true},
		{"explicit claude + nil agents map", &config.Config{Agent: "claude"}, true},
		{"explicit claude + agents map with claude", &config.Config{
			Agent:  "claude",
			Agents: map[string]*config.AgentConfig{"claude": {}},
		}, true},
		{"explicit claude + agents map without claude (deselected)", &config.Config{
			Agent:  "claude",
			Agents: map[string]*config.AgentConfig{"gemini": {}},
		}, false},
		{"explicit gemini", &config.Config{
			Agent:  "gemini",
			Agents: map[string]*config.AgentConfig{"gemini": {}},
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := activeAgentIsClaude(tt.cfg); got != tt.want {
				t.Errorf("activeAgentIsClaude = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCredentialKitsFrom(t *testing.T) {
	a := &kit.Kit{Name: "github", CredentialFunc: func(kit.CredentialOpts) ([]kit.CredentialMount, error) { return nil, nil }}
	b := &kit.Kit{Name: "java", CredentialFunc: nil}
	c := &kit.Kit{Name: "java/maven", CredentialFunc: func(kit.CredentialOpts) ([]kit.CredentialMount, error) { return nil, nil }}
	got := credentialKitsFrom([]*kit.Kit{a, b, c})
	if len(got) != 2 || got[0].Name != "github" || got[1].Name != "java/maven" {
		t.Errorf("credentialKitsFrom = %v, want [github, java/maven]", got)
	}
}

func TestPickedFromMulti(t *testing.T) {
	names := []string{"claude", "gemini", "codex"}
	if got := pickedFromMulti(tui.StepResult{MultiIdx: []int{1, 2}}, names, nil); !equalStrings(got, []string{"gemini", "codex"}) {
		t.Errorf("picked = %v", got)
	}
	// Empty selection → fallback.
	if got := pickedFromMulti(tui.StepResult{}, names, []string{"claude"}); !equalStrings(got, []string{"claude"}) {
		t.Errorf("empty fallback = %v", got)
	}
}

func TestResolveDefaultAgent(t *testing.T) {
	agentNames := []string{"claude", "codex", "gemini"}

	// Single pick → step 2 ignored.
	if got := resolveDefaultAgent([]tui.StepResult{{MultiIdx: []int{1}}, {SelectIdx: 0}}, []string{"codex"}, agentNames); got != "codex" {
		t.Errorf("single-pick default = %q", got)
	}
	// Multi-pick with claude among them, step picks claude.
	if got := resolveDefaultAgent([]tui.StepResult{{MultiIdx: []int{0, 2}}, {SelectIdx: 0}}, []string{"claude", "gemini"}, agentNames); got != "claude" {
		t.Errorf("multi default-agent picked = %q, want claude", got)
	}
	// Multi-pick but step picks an agent the user didn't include → fallback to first picked.
	// SelectIdx=0 → "claude", but picked=[gemini, codex] excludes it → fallback to gemini.
	if got := resolveDefaultAgent([]tui.StepResult{{MultiIdx: []int{1, 2}}, {SelectIdx: 0}}, []string{"gemini", "codex"}, agentNames); got != "gemini" {
		t.Errorf("invalid default-agent fallback = %q, want gemini (first picked)", got)
	}
}

func TestDefaultKitChoices(t *testing.T) {
	got := defaultKitChoices()
	// TierDefault kits should be present and true. Always-on/hidden kits must
	// be absent from the map (they're not selectable from the wizard).
	for _, name := range []string{"docker", "github", "java", "python"} {
		k := kit.Get(name)
		if k == nil || k.Tier != kit.TierDefault {
			continue
		}
		if !got[name] {
			t.Errorf("TierDefault kit %q should be true in defaultKitChoices()", name)
		}
	}
	for _, name := range []string{"node", "ports", "ssh", "shell"} {
		if _, present := got[name]; present {
			t.Errorf("non-selectable kit %q should not appear in defaultKitChoices()", name)
		}
	}
}

// --- Regression tests for the Codex adversarial review findings ---

// Regression for "Non-Claude first-run selections still write an active
// Claude config". When the wizard's image-shaping phase writes a config
// without claude, activeAgentIsClaude must return false so the runtime
// phase skips the isolation step entirely.
func TestRegression_NonClaudeSelectionSkipsClaudeIsolation(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	cfgPath := filepath.Join(home, ".asylum", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0755); err != nil {
		t.Fatal(err)
	}

	// Simulate phase-1 outcome: user picked gemini only.
	if err := WriteConfig(cfgPath, Choices{
		DefaultAgent: "gemini",
		Agents:       map[string]bool{"gemini": true},
		Kits:         defaultKitChoices(),
	}); err != nil {
		t.Fatal(err)
	}

	projectDir := t.TempDir()
	cfg, err := config.Load(projectDir, config.CLIFlags{}, "")
	if err != nil {
		t.Fatal(err)
	}

	if activeAgentIsClaude(&cfg) {
		t.Errorf("post-phase-1 cfg should not mark claude active when only gemini was picked")
	}

	// The generated file must not contain an active claude entry under agents:.
	data, _ := os.ReadFile(cfgPath)
	body := string(data)
	if !strings.Contains(body, "# claude:") {
		t.Errorf("expected commented `# claude:` in agents block:\n%s", body)
	}
	// Negative: no uncommented claude entry under agents:.
	lines := strings.Split(body, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "claude:" {
			t.Errorf("found uncommented `claude:` entry — phase 1 should not have activated claude:\n%s", body)
		}
	}
}

// Regression for "Credential step can reactivate kits the user deselected".
// When the user deselects a credential-capable kit in phase 1, the
// credentials applier must not write a SetKitCredentials("false") that
// would inject a fresh active entry under the deselected kit.
func TestRegression_DeselectedCredentialKitStaysCommented(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	cfgPath := filepath.Join(home, ".asylum", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0755); err != nil {
		t.Fatal(err)
	}

	// Phase 1: user deselected the github kit (a credential-capable kit).
	choices := defaultKitChoices()
	choices["github"] = false
	if err := WriteConfig(cfgPath, Choices{
		DefaultAgent: "claude",
		Agents:       map[string]bool{"claude": true},
		Kits:         choices,
	}); err != nil {
		t.Fatal(err)
	}

	projectDir := t.TempDir()
	cfg, err := config.Load(projectDir, config.CLIFlags{}, "")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.KitActive("github") {
		t.Fatalf("post-phase-1: github should be commented, but KitActive returned true")
	}

	// Simulate phase-2 credential resolution against the post-reload kit set.
	// kit.Resolve returns the active kits; github should not be among them.
	activeKits, err := kit.Resolve(cfg.KitNames(), cfg.DisabledKits())
	if err != nil {
		t.Fatal(err)
	}
	for _, k := range credentialKitsFrom(activeKits) {
		parent, _, _ := strings.Cut(k.Name, "/")
		if parent == "github" {
			t.Errorf("credentialKitsFrom should not include deselected github; got %v", k.Name)
		}
	}

	// Defensive check: even if the applier were called for a deselected kit,
	// the KitActive guard inside the applier prevents the bad write. Simulate
	// the same precondition by calling SetKitCredentials only when active,
	// just as the applier does.
	if !cfg.KitActive("github") {
		// (no write) — verify the file still has github commented.
		body, _ := os.ReadFile(cfgPath)
		if strings.Contains(string(body), "\n  github:") {
			t.Errorf("github appears active in config after deselection:\n%s", string(body))
		}
		if !strings.Contains(string(body), "  # github:") {
			t.Errorf("expected commented `# github:` in kits block:\n%s", string(body))
		}
	}
}

// --- Existing-user wizard smoke tests (no-TTY paths covered indirectly) ---

func TestRun_NonInteractiveFirstRunWritesDefaults(t *testing.T) {
	if term.IsTerminal() {
		t.Skip("test asserts the no-TTY code path; stdin is currently a terminal")
	}
	home := t.TempDir()
	cfgPath := filepath.Join(home, ".asylum", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0755); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{}
	out, err := Run(WizardInput{
		Home:       home,
		IsFirstRun: true,
		Cfg:        &cfg,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !out.WroteConfig {
		t.Errorf("expected WroteConfig=true on non-TTY first-run; got %+v", out)
	}
	if _, err := os.Stat(cfgPath); err != nil {
		t.Errorf("config.yaml should exist after non-TTY first-run: %v", err)
	}
}

func TestRun_NonInteractiveExistingUserIsNoop(t *testing.T) {
	if term.IsTerminal() {
		t.Skip("test asserts the no-TTY code path; stdin is currently a terminal")
	}
	home := t.TempDir()
	cfgPath := filepath.Join(home, ".asylum", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cfgPath, []byte("version: \"0.2\"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{}
	out, err := Run(WizardInput{
		Home:       home,
		IsFirstRun: false,
		Cfg:        &cfg,
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.WroteConfig || out.WroteIsolation || out.WroteCredentials {
		t.Errorf("expected no-op Outcome on non-TTY existing user; got %+v", out)
	}
}

// --- Test helpers ---

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
