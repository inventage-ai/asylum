package firstrun

import (
	"os"
	"path/filepath"

	"github.com/inventage-ai/asylum/internal/kit"
	"github.com/inventage-ai/asylum/internal/log"
	"github.com/inventage-ai/asylum/internal/tui"
)

// Run detects a first-run condition and prompts the user to enable
// kit credential support via a TUI multiselect. Selected kits get
// `credentials: auto` written to ~/.asylum/config.yaml.
// Uses ~/.asylum/agents/ as the signal that asylum has been used before.
func Run(homeDir string) error {
	agentsDir := filepath.Join(homeDir, ".asylum", "agents")
	if _, err := os.Stat(agentsDir); err == nil {
		return nil // existing user
	}

	kits := kit.CredentialCapableKits()
	if len(kits) == 0 {
		return nil
	}

	selected := promptCredentials(kits)
	if len(selected) == 0 {
		return nil
	}

	cfgPath := filepath.Join(homeDir, ".asylum", "config.yaml")
	for _, k := range selected {
		// Use the parent kit name for config (e.g. "java" not "java/maven")
		kitName, _, _ := cutKitName(k.Name)
		if err := SetKitCredentials(cfgPath, kitName, "auto"); err != nil {
			return err
		}
	}
	log.Success("credential support enabled in %s", cfgPath)
	return nil
}

func promptCredentials(kits []*kit.Kit) []*kit.Kit {
	options := make([]tui.Option, len(kits))
	for i, k := range kits {
		label := k.CredentialLabel
		if label == "" {
			label = k.Name
		}
		options[i] = tui.Option{
			Label:       label,
			Description: "Filters credentials by project needs",
		}
	}

	// Default all selected
	defaults := make([]int, len(kits))
	for i := range defaults {
		defaults[i] = i
	}

	indices, err := tui.MultiSelect("Enable credential support for:", options, defaults)
	if err != nil {
		return nil // cancelled
	}

	var selected []*kit.Kit
	for _, i := range indices {
		selected = append(selected, kits[i])
	}
	return selected
}

func cutKitName(name string) (parent, child string, hasChild bool) {
	for i, c := range name {
		if c == '/' {
			return name[:i], name[i+1:], true
		}
	}
	return name, "", false
}
