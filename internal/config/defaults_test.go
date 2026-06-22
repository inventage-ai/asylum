package config

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/inventage-ai/asylum/internal/kit"
)

// The openspec kit is default-on, so it must appear as an active (uncommented)
// kit in the generated default config — otherwise its Docker snippet, init
// script, rules, and seeded config would never be applied.
func TestDefaultConfigActivatesDefaultKits(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".asylum")
	if err := WriteDefaults(path, kit.AssembleConfigSnippets()); err != nil {
		t.Fatalf("write defaults: %v", err)
	}
	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("parse default config: %v", err)
	}
	if !slices.Contains(cfg.KitNames(), "openspec") {
		t.Errorf("expected openspec active in default config, got kits: %v", cfg.KitNames())
	}
}
