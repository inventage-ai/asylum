package config

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/inventage-ai/asylum/internal/kit"
	"github.com/inventage-ai/asylum/internal/log"
)

// SyncNewKits detects kits not yet in state.json, prompts for activation
// (if interactive and promptFn is provided), updates the config file, and
// saves the new state. Returns true if any new kits were processed.
//
// promptFn receives the list of new promptable kits (TierDefault and
// TierOptIn) and returns the names the user chose to activate. The caller
// can inspect each kit's Tier to decide pre-selection. When promptFn is
// nil, all TierDefault kits are activated automatically and TierOptIn
// kits are added as comments.
func SyncNewKits(asylumDir string, interactive bool, promptFn func([]*kit.Kit) []string) (bool, error) {
	state, err := LoadState(asylumDir)
	if err != nil {
		return false, fmt.Errorf("load state: %w", err)
	}

	configPath := filepath.Join(asylumDir, "config.yaml")

	// If the global config doesn't exist yet (first run or upgrade from a
	// pre-config version) or needs v1→v2 migration, the user already has
	// their kits configured (or will get them via WriteDefaults).
	// Mark all kits as seen so they aren't prompted.
	if _, err := os.Stat(configPath); os.IsNotExist(err) || NeedsMigration(configPath) {
		state.KnownKits = kit.All()
		if err := SaveState(asylumDir, state); err != nil {
			return false, fmt.Errorf("save state: %w", err)
		}
		return false, nil
	}

	newKits := NewKits(kit.All(), state)
	if len(newKits) == 0 {
		return false, nil
	}

	// Classify new kits by tier.
	var promptable []*kit.Kit
	for _, name := range newKits {
		k := kit.Get(name)
		if k == nil {
			continue
		}

		switch {
		case k.Tier == kit.TierAlwaysOn:
			log.Info("new kit: %s (always active)", name)
		case k.Hidden:
			// Hidden kits are silently added as comments — not prompted.
			if k.ConfigComment != "" {
				if err := SyncKitCommentToConfig(configPath, k.ConfigComment); err != nil {
					log.Error("sync kit %s: %v", name, err)
				}
			}
		default:
			promptable = append(promptable, k)
		}
	}

	// Prompt for kit selection, or apply defaults when non-interactive.
	if len(promptable) > 0 {
		var activated []string
		if interactive && promptFn != nil {
			activated = promptFn(promptable)
		} else {
			for _, k := range promptable {
				if k.Tier == kit.TierDefault {
					activated = append(activated, k.Name)
				}
			}
		}

		for _, k := range promptable {
			if slices.Contains(activated, k.Name) {
				if k.ConfigSnippet != "" {
					if err := SyncKitToConfig(configPath, k.Name, k.ConfigSnippet); err != nil {
						log.Error("sync kit %s: %v", k.Name, err)
					}
				}
			} else {
				if k.ConfigComment != "" {
					if err := SyncKitCommentToConfig(configPath, k.ConfigComment); err != nil {
						log.Error("sync kit %s: %v", k.Name, err)
					}
				}
				// silently added as comment in config
			}
		}
	}

	// Update state with all currently registered kits
	state.KnownKits = kit.All()
	if err := SaveState(asylumDir, state); err != nil {
		return true, fmt.Errorf("save state: %w", err)
	}

	return true, nil
}
