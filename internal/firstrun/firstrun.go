package firstrun

import (
	"os"
	"path/filepath"
)

// Run detects a first-run condition. Credential prompting has moved to
// the unified onboarding wizard in main.go. This function remains as a
// shell for any future first-run-only tasks.
// Uses ~/.asylum/agents/ as the signal that asylum has been used before.
func Run(homeDir string) error {
	agentsDir := filepath.Join(homeDir, ".asylum", "agents")
	if _, err := os.Stat(agentsDir); err == nil {
		return nil // existing user
	}

	// Future first-run-only tasks go here.
	return nil
}
