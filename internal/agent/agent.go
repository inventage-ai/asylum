package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/inventage-ai/asylum/internal/term"
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
		names := make([]string, 0, len(agents))
		for k := range agents {
			names = append(names, k)
		}
		slices.Sort(names)
		return nil, fmt.Errorf("unknown agent: %q (valid: %s)", name, strings.Join(names, ", "))
	}
	return a, nil
}

func wrapZsh(cmd string) []string {
	return []string{"zsh", "-c", "source ~/.zshrc && exec " + cmd}
}

// shellQuote delegates to term.ShellQuote for backward compatibility within this package.
var shellQuote = term.ShellQuote

// quoteArgs delegates to term.ShellQuoteArgs for backward compatibility within this package.
var quoteArgs = term.ShellQuoteArgs

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
