package agent

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/inventage-ai/asylum/internal/log"
)

type Codex struct{}

func (Codex) Name() string             { return "codex" }
func (Codex) Binary() string           { return "codex" }
func (Codex) NativeConfigDir() string  { return "~/.codex" }
func (Codex) ContainerConfigDir() string { return "/home/claude/.codex" }
func (Codex) AsylumConfigDir() string  { return "~/.asylum/agents/codex" }

func (Codex) EnvVars() map[string]string {
	return map[string]string{
		"CODEX_HOME": "/home/claude/.codex",
	}
}

// markerDir returns the directory used to store per-project session markers.
// Codex stores sessions in a global date-organized tree with no per-project
// metadata, so we use a separate marker to avoid resuming the wrong project.
func (Codex) markerDir(projectPath string) (string, error) {
	configDir, err := expandHome("~/.asylum/agents/codex")
	if err != nil {
		return "", err
	}
	encoded := strings.ReplaceAll(projectPath, "/", "-")
	return filepath.Join(configDir, "projects", encoded), nil
}

func (c Codex) HasSession(projectPath string) bool {
	dir, err := c.markerDir(projectPath)
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(dir, ".has_session"))
	return err == nil
}

func (c Codex) WriteMarker(projectPath string) error {
	dir, err := c.markerDir(projectPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, ".has_session"), nil, 0644)
}

func (Codex) Command(resume bool, extraArgs []string) []string {
	if resume {
		if len(extraArgs) == 0 {
			return wrapZsh("codex resume --last --yolo")
		}
		log.Warn("codex: resume skipped because extra args were provided")
	}
	parts := []string{"codex", "--yolo"}
	parts = append(parts, quoteArgs(extraArgs)...)
	return wrapZsh(strings.Join(parts, " "))
}
