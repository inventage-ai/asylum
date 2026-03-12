package agent

import (
	"os"
	"path/filepath"
	"strings"
)

type Claude struct{}

func (Claude) Name() string             { return "claude" }
func (Claude) Binary() string           { return "claude" }
func (Claude) NativeConfigDir() string  { return "~/.claude" }
func (Claude) ContainerConfigDir() string { return "/home/claude/.claude" }
func (Claude) AsylumConfigDir() string  { return "~/.asylum/agents/claude" }

func (Claude) EnvVars() map[string]string {
	return map[string]string{
		"CLAUDE_CONFIG_DIR": "/home/claude/.claude",
	}
}

func (Claude) HasSession(projectPath string) bool {
	configDir, err := expandHome("~/.asylum/agents/claude")
	if err != nil {
		return false
	}
	projectsDir := filepath.Join(configDir, "projects")
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return false
	}
	return len(entries) > 0
}

func (Claude) Command(resume bool, extraArgs []string) []string {
	parts := []string{"claude", "--dangerously-skip-permissions"}
	if resume {
		parts = append(parts, "--continue")
	}
	parts = append(parts, extraArgs...)
	return wrapZsh(strings.Join(parts, " "))
}

func expandHome(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}
