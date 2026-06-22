package firstrun

import (
	"path/filepath"
	"strings"

	"github.com/inventage-ai/asylum/internal/config"
	"github.com/inventage-ai/asylum/internal/kit"
	"github.com/inventage-ai/asylum/internal/log"
	"github.com/inventage-ai/asylum/internal/term"
	"github.com/inventage-ai/asylum/internal/tui"
)

// WizardInput holds the inputs the wizard inspects to decide which steps to
// present and which defaults to pre-select. After the image-shaping phase
// writes its config, Reload is invoked so the runtime-phase steps and
// appliers operate on the fresh state rather than the pre-wizard view.
type WizardInput struct {
	Home       string
	IsFirstRun bool
	Cfg        *config.Config
	AllKits    []*kit.Kit
	// Reload re-loads the merged config and re-resolves the active kit set
	// from disk. firstrun calls it after writing the image-shaping config
	// so subsequent steps don't gate on stale pre-wizard state. May be nil
	// for non-first-run callers (no image-shaping phase runs).
	Reload func() (*config.Config, []*kit.Kit, error)
}

// Outcome reports what the wizard wrote so the caller can decide whether
// the merged config needs to be re-loaded once more for downstream image
// generation.
type Outcome struct {
	WroteConfig      bool // first-run config was written from scratch
	WroteIsolation   bool
	WroteCredentials bool
}

// Run presents the first-run wizard (or its trimmed existing-user variant)
// and writes results to ~/.asylum/config.yaml. The caller is expected to
// re-load the config when Outcome.WroteConfig is true; the wizard itself
// has already reloaded internally between its two phases.
func Run(in WizardInput) (Outcome, error) {
	var out Outcome

	if !term.IsTerminal() {
		if in.IsFirstRun {
			if err := writeDefaults(in.Home); err != nil {
				return out, err
			}
			out.WroteConfig = true
		}
		return out, nil
	}

	if in.IsFirstRun {
		if err := runImageShapingPhase(in, &out); err != nil {
			return out, err
		}
		// Reload between phases so the runtime-phase step builders gate
		// on the user's selections rather than the empty pre-wizard view.
		if out.WroteConfig && in.Reload != nil {
			cfg, kits, err := in.Reload()
			if err != nil {
				return out, err
			}
			*in.Cfg = *cfg
			in.AllKits = kits
		}
	}

	if err := runRuntimePhase(in, &out); err != nil {
		return out, err
	}
	return out, nil
}

// runImageShapingPhase prompts for agents, default agent, and kits, then
// writes the resulting config from scratch. Only fires on first-run.
func runImageShapingPhase(in WizardInput, out *Outcome) error {
	agentNames := agentPickerNames()
	if len(agentNames) == 0 {
		return nil
	}

	agentOpts := make([]tui.Option, len(agentNames))
	var agentDefault []int
	for i, n := range agentNames {
		agentOpts[i] = tui.Option{Label: n}
		if n == "claude" {
			agentDefault = append(agentDefault, i)
		}
	}

	defaultAgentOpts := make([]tui.Option, len(agentNames))
	claudeIdx := 0
	for i, n := range agentNames {
		defaultAgentOpts[i] = tui.Option{Label: n}
		if n == "claude" {
			claudeIdx = i
		}
	}

	steps := []tui.WizardStep{
		{
			Title:       "Agents",
			Description: "Which coding agents should be installed in the sandbox image? Multiple are allowed; pick one as the default in the next step if needed.",
			Kind:        tui.StepMultiSelect,
			Options:     agentOpts,
			DefaultSel:  agentDefault,
		},
		{
			Title:       "Default agent",
			Description: "Which of the selected agents should be invoked when you run `asylum` without `-a`?",
			Kind:        tui.StepSelect,
			Options:     defaultAgentOpts,
			DefaultIdx:  claudeIdx,
		},
	}

	var kitOptions []*kit.Kit
	for _, name := range kit.All() {
		k := kit.Get(name)
		if k == nil || !isSelectable(k) {
			continue
		}
		kitOptions = append(kitOptions, k)
	}
	if len(kitOptions) > 0 {
		kitOpts := make([]tui.Option, len(kitOptions))
		var kitDefault []int
		for i, k := range kitOptions {
			kitOpts[i] = tui.Option{Label: k.Name, Description: k.Description}
			if k.Tier == kit.TierDefault {
				kitDefault = append(kitDefault, i)
			}
		}
		steps = append(steps, tui.WizardStep{
			Title:       "Kits",
			Description: "Which language toolchains and tools should the sandbox include? Defaults match a typical setup; pick more to add them, deselect to omit.",
			Kind:        tui.StepMultiSelect,
			Options:     kitOpts,
			DefaultSel:  kitDefault,
		})
	}

	log.Info("Welcome to asylum — let's set up your sandbox.")
	results, err := tui.Wizard(steps)
	if err != nil {
		return err
	}

	pickedAgents := pickedFromMulti(results[0], agentNames, []string{"claude"})
	chosenDefault := resolveDefaultAgent(results, pickedAgents, agentNames)
	chosenKits := map[string]bool{}
	if len(kitOptions) > 0 {
		for _, idx := range results[2].MultiIdx {
			chosenKits[kitOptions[idx].Name] = true
		}
	}

	agentMap := map[string]bool{}
	for _, name := range pickedAgents {
		agentMap[name] = true
	}
	kitMap := map[string]bool{}
	for _, k := range kitOptions {
		kitMap[k.Name] = chosenKits[k.Name]
	}

	cfgPath := filepath.Join(in.Home, ".asylum", "config.yaml")
	if err := WriteConfig(cfgPath, Choices{
		DefaultAgent: chosenDefault,
		Agents:       agentMap,
		Kits:         kitMap,
	}); err != nil {
		return err
	}
	out.WroteConfig = true
	return nil
}

// runRuntimePhase prompts for isolation (if Claude is active and unconfigured)
// and credentials (if any active credential-capable kit is unconfigured).
// Both steps gate on the post-reload state, so first-run selections that
// dropped Claude or a credential-capable kit do not surface stale prompts.
func runRuntimePhase(in WizardInput, out *Outcome) error {
	if in.Cfg == nil {
		return nil
	}

	type applier func(tui.StepResult)
	var steps []tui.WizardStep
	var appliers []applier

	// Isolation only matters for Claude. If the active agent isn't Claude,
	// skip — and never create a stray `agents.claude.config` entry that
	// would activate a deselected agent.
	if activeAgentIsClaude(in.Cfg) && in.Cfg.AgentIsolation("claude") == "" {
		steps = append(steps, tui.WizardStep{
			Title:       "Config Isolation",
			Description: "How should Claude's config (~/.claude) be managed inside the sandbox?",
			Kind:        tui.StepSelect,
			Options: []tui.Option{
				{Label: "Shared with host (recommended)", Description: "Use your host ~/.claude directly. Changes sync both ways."},
				{Label: "Isolated", Description: "Separate from host, shared across projects."},
				{Label: "Project-isolated", Description: "Separate config per project. No state shared between projects."},
			},
			DefaultIdx: 0,
		})
		appliers = append(appliers, func(r tui.StepResult) {
			levels := []string{"shared", "isolated", "project"}
			level := levels[r.SelectIdx]
			cfgPath := filepath.Join(in.Home, ".asylum", "config.yaml")
			if err := config.SetAgentIsolation(cfgPath, "claude", level); err != nil {
				log.Error("save isolation config: %v", err)
				return
			}
			if in.Cfg.Agents == nil {
				in.Cfg.Agents = map[string]*config.AgentConfig{}
			}
			if in.Cfg.Agents["claude"] == nil {
				in.Cfg.Agents["claude"] = &config.AgentConfig{}
			}
			in.Cfg.Agents["claude"].Config = level
			out.WroteIsolation = true
		})
	}

	// Credential prompts are gated on the post-reload active kit set so a
	// kit the user deselected in phase 1 cannot reappear as an active YAML
	// entry via a `credentials: false` write.
	credKits := credentialKitsFrom(in.AllKits)
	hasUnconfigured := false
	for _, k := range credKits {
		parent, _, _ := strings.Cut(k.Name, "/")
		if in.Cfg.KitCredentialMode(parent) == "" {
			hasUnconfigured = true
			break
		}
	}
	if hasUnconfigured {
		options := make([]tui.Option, len(credKits))
		var preSelected []int
		for i, k := range credKits {
			label := k.CredentialLabel
			if label == "" {
				label = k.Name
			}
			options[i] = tui.Option{Label: label}
			parent, _, _ := strings.Cut(k.Name, "/")
			if in.Cfg.KitCredentialMode(parent) != "" {
				preSelected = append(preSelected, i)
			}
		}
		steps = append(steps, tui.WizardStep{
			Title:       "Credentials",
			Description: "Allow the sandbox to access host credentials for private registries and repositories.",
			Kind:        tui.StepMultiSelect,
			Options:     options,
			DefaultSel:  preSelected,
		})
		appliers = append(appliers, func(r tui.StepResult) {
			cfgPath := filepath.Join(in.Home, ".asylum", "config.yaml")
			selected := map[int]bool{}
			for _, idx := range r.MultiIdx {
				selected[idx] = true
			}
			for i, k := range credKits {
				parent, _, _ := strings.Cut(k.Name, "/")
				if selected[i] {
					if err := config.SetKitCredentials(cfgPath, parent, "auto"); err != nil {
						log.Error("save credential config: %v", err)
						continue
					}
					if in.Cfg.Kits == nil {
						in.Cfg.Kits = map[string]*config.KitConfig{}
					}
					if in.Cfg.Kits[parent] == nil {
						in.Cfg.Kits[parent] = &config.KitConfig{}
					}
					in.Cfg.Kits[parent].Credentials = &config.Credentials{Auto: true}
					out.WroteCredentials = true
				} else if in.Cfg.KitCredentialMode(parent) == "" {
					// Only record an explicit "off" for kits that are already
					// active in the resolved config. Writing under a kit that
					// only exists as a comment would inject a fresh active
					// entry and resurrect a deselected kit.
					if !in.Cfg.KitActive(parent) {
						continue
					}
					if err := config.SetKitCredentials(cfgPath, parent, "false"); err != nil {
						log.Error("save credential config: %v", err)
						continue
					}
					if in.Cfg.Kits == nil {
						in.Cfg.Kits = map[string]*config.KitConfig{}
					}
					if in.Cfg.Kits[parent] == nil {
						in.Cfg.Kits[parent] = &config.KitConfig{}
					}
					in.Cfg.Kits[parent].Credentials = &config.Credentials{}
					out.WroteCredentials = true
				}
			}
		})
	}

	if len(steps) == 0 {
		return nil
	}
	results, err := tui.Wizard(steps)
	if err != nil {
		return err
	}
	for i, r := range results {
		if r.Completed {
			appliers[i](r)
		}
	}
	return nil
}

// activeAgentIsClaude reports whether the resolved config selects Claude as
// its default agent (either explicitly or via the historical fallback).
func activeAgentIsClaude(cfg *config.Config) bool {
	if cfg == nil {
		return false
	}
	name := cfg.Agent
	if name == "" {
		name = "claude"
	}
	if name != "claude" {
		return false
	}
	// When the user explicitly deselected Claude in the wizard, the
	// reloaded config has Agents populated and Claude must appear there
	// for it to be considered active.
	if cfg.Agents == nil {
		return true
	}
	_, ok := cfg.Agents["claude"]
	return ok
}

// credentialKitsFrom filters a kit slice down to ones that contribute a
// credential prompt.
func credentialKitsFrom(kits []*kit.Kit) []*kit.Kit {
	var out []*kit.Kit
	for _, k := range kits {
		if k.CredentialFunc != nil {
			out = append(out, k)
		}
	}
	return out
}

// pickedFromMulti maps a multi-select StepResult back to the picked names,
// applying a fallback when the user accepted with an empty selection.
func pickedFromMulti(r tui.StepResult, names []string, emptyFallback []string) []string {
	if len(r.MultiIdx) == 0 {
		return append([]string(nil), emptyFallback...)
	}
	out := make([]string, 0, len(r.MultiIdx))
	for _, idx := range r.MultiIdx {
		out = append(out, names[idx])
	}
	return out
}

// resolveDefaultAgent picks the top-level `agent:` value from the wizard's
// default-agent step, falling back to the single picked agent when only one
// was selected (default-agent prompt was effectively a no-op).
func resolveDefaultAgent(results []tui.StepResult, picked []string, agentNames []string) string {
	if len(picked) == 1 {
		return picked[0]
	}
	if len(results) < 2 || results[1].SelectIdx >= len(agentNames) {
		return picked[0]
	}
	candidate := agentNames[results[1].SelectIdx]
	for _, p := range picked {
		if p == candidate {
			return candidate
		}
	}
	return picked[0]
}

// writeDefaults writes today's silent-default configuration: claude only,
// TierDefault kits active, TierAvailable kits commented. Used for the
// no-TTY code path.
func writeDefaults(home string) error {
	cfgPath := filepath.Join(home, ".asylum", "config.yaml")
	return WriteConfig(cfgPath, Choices{
		DefaultAgent: "claude",
		Agents:       map[string]bool{"claude": true},
		Kits:         defaultKitChoices(),
	})
}

// defaultKitChoices returns the "press enter through everything" kit map:
// TierDefault top-level kits active, everything else inactive.
func defaultKitChoices() map[string]bool {
	out := map[string]bool{}
	for _, name := range kit.All() {
		k := kit.Get(name)
		if k == nil || !isSelectable(k) {
			continue
		}
		out[name] = k.Tier == kit.TierDefault
	}
	return out
}
