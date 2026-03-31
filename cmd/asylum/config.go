package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/inventage-ai/asylum/internal/config"
	"github.com/inventage-ai/asylum/internal/kit"
	"github.com/inventage-ai/asylum/internal/log"
	"github.com/inventage-ai/asylum/internal/term"
	"github.com/inventage-ai/asylum/internal/tui"
)

func runConfig() {
	if !term.IsTerminal() {
		die("asylum config requires a terminal")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		die("home dir: %v", err)
	}

	cfgPath := filepath.Join(home, ".asylum", "config.yaml")

	kitSnippets := kit.AssembleConfigSnippets()
	cfg, err := config.Load(".", config.CLIFlags{}, kitSnippets)
	if err != nil {
		die("load config: %v", err)
	}

	// Build kits tab: all non-always-on registered kits
	activeKits := map[string]bool{}
	for _, name := range cfg.KitNames() {
		activeKits[name] = cfg.KitActive(name)
	}

	var kitOptions []tui.Option
	var kitDefaultSel []int
	var kitNames []string // parallel to kitOptions
	for _, name := range kit.All() {
		k := kit.Get(name)
		if k.Tier == kit.TierAlwaysOn {
			continue
		}
		kitOptions = append(kitOptions, tui.Option{Label: k.Name, Description: k.Description})
		kitNames = append(kitNames, k.Name)
		if activeKits[k.Name] {
			kitDefaultSel = append(kitDefaultSel, len(kitOptions)-1)
		}
	}

	// Build credentials tab
	credKits := kit.CredentialCapableKits()
	var credOptions []tui.Option
	var credDefaultSel []int
	for i, k := range credKits {
		label := k.CredentialLabel
		if label == "" {
			label = k.Name
		}
		credOptions = append(credOptions, tui.Option{Label: label, Description: k.Name})
		parent, _, _ := strings.Cut(k.Name, "/")
		if cfg.KitCredentialMode(parent) == "auto" {
			credDefaultSel = append(credDefaultSel, i)
		}
	}

	// Build isolation tab
	isolationLevels := []string{"shared", "isolated", "project"}
	isolationOptions := []tui.Option{
		{Label: "Shared with host", Description: "Use your host ~/.claude directly. Changes sync both ways."},
		{Label: "Isolated (recommended)", Description: "Separate from host, shared across projects."},
		{Label: "Project-isolated", Description: "Separate config per project. No state shared between projects."},
	}
	currentIsolation := cfg.AgentIsolation("claude")
	isolationDefault := 1 // default to "isolated"
	for i, level := range isolationLevels {
		if level == currentIsolation {
			isolationDefault = i
			break
		}
	}

	tabs := []tui.Tab{
		{Title: "Kits", Description: "Toggle kits on or off.", Kind: tui.StepMultiSelect, Options: kitOptions, DefaultSel: kitDefaultSel},
		{Title: "Credentials", Description: "Allow the sandbox to access host credentials (read-only).", Kind: tui.StepMultiSelect, Options: credOptions, DefaultSel: credDefaultSel},
		{Title: "Isolation", Description: "How should Claude's config (~/.claude) be managed inside the sandbox?", Kind: tui.StepSelect, Options: isolationOptions, DefaultIdx: isolationDefault},
	}

	results, err := tui.RunTabs(tabs)
	if err != nil {
		return // cancelled
	}

	// Apply kit changes
	newActive := map[string]bool{}
	for _, idx := range results[0].MultiIdx {
		newActive[kitNames[idx]] = true
	}
	for _, name := range kitNames {
		wasActive := activeKits[name]
		nowActive := newActive[name]
		if nowActive && !wasActive {
			if config.KitExistsInFile(cfgPath, name) {
				// Previously enabled, now has disabled: true — remove the flag.
				if err := config.RemoveKitDisabled(cfgPath, name); err != nil {
					log.Error("activate kit %s: %v", name, err)
				}
			} else {
				// Never enabled — remove comment block, insert clean entry.
				config.RemoveKitComment(cfgPath, name)
				if err := config.SyncKitToConfig(cfgPath, name, "  "+name+":"); err != nil {
					log.Error("activate kit %s: %v", name, err)
				}
			}
		} else if !nowActive && wasActive {
			if err := config.SetKitDisabled(cfgPath, name); err != nil {
				log.Error("deactivate kit %s: %v", name, err)
			}
		}
	}

	// Apply credential changes
	credSelected := map[int]bool{}
	for _, idx := range results[1].MultiIdx {
		credSelected[idx] = true
	}
	for i, k := range credKits {
		parent, _, _ := strings.Cut(k.Name, "/")
		if credSelected[i] {
			if err := config.SetKitCredentials(cfgPath, parent, "auto"); err != nil {
				log.Error("set credentials %s: %v", parent, err)
			}
		} else {
			if err := config.SetKitCredentials(cfgPath, parent, "false"); err != nil {
				log.Error("set credentials %s: %v", parent, err)
			}
		}
	}

	// Apply isolation change
	level := isolationLevels[results[2].SelectIdx]
	if err := config.SetAgentIsolation(cfgPath, "claude", level); err != nil {
		log.Error("set isolation: %v", err)
	}

	log.Success("Config updated")
}
