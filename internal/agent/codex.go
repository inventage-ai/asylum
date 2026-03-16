package agent

import (
	"os"
	"path/filepath"
	"strings"
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

func (Codex) HasSession(projectPath string) bool {
	configDir, err := expandHome("~/.asylum/agents/codex")
	if err != nil {
		return false
	}
	sessionsDir := filepath.Join(configDir, "sessions")
	found := false
	filepath.WalkDir(sessionsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || found {
			return filepath.SkipDir
		}
		if !d.IsDir() && strings.HasPrefix(d.Name(), "rollout-") && strings.HasSuffix(d.Name(), ".jsonl") {
			found = true
			return filepath.SkipDir
		}
		return nil
	})
	return found
}

func (Codex) Command(resume bool, extraArgs []string) []string {
	if resume && len(extraArgs) == 0 {
		return wrapZsh("codex resume --last --yolo")
	}
	parts := []string{"codex", "--yolo"}
	parts = append(parts, extraArgs...)
	return wrapZsh(strings.Join(parts, " "))
}
