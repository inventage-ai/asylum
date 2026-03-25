package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Agent interface {
	Name() string
	Binary() string
	NativeConfigDir() string
	ContainerConfigDir() string
	AsylumConfigDir() string
	EnvVars() map[string]string
	HasSession(projectPath string) bool
	Command(resume bool, extraArgs []string) []string
}

var agents = map[string]Agent{
	"claude": Claude{},
	"gemini": Gemini{},
	"codex":  Codex{},
}

func Get(name string) (Agent, error) {
	a, ok := agents[name]
	if !ok {
		return nil, fmt.Errorf("unknown agent: %q (valid: claude, codex, echo, gemini, opencode)", name)
	}
	return a, nil
}

func wrapZsh(cmd string) []string {
	return []string{"zsh", "-c", "source ~/.zshrc && exec " + cmd}
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

func quoteArgs(args []string) []string {
	quoted := make([]string, len(args))
	for i, a := range args {
		quoted[i] = shellQuote(a)
	}
	return quoted
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

// resolveConfigDir expands the agent's AsylumConfigDir (which uses ~ prefix)
// to an absolute path.
func resolveConfigDir(a Agent) (string, error) {
	return expandHome(a.AsylumConfigDir())
}
