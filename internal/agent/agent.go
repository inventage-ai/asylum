package agent

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/inventage-ai/asylum/internal/config"
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

var agents = map[string]Agent{}

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

// resolveConfigDir expands the agent's AsylumConfigDir (which uses ~ prefix)
// to an absolute path.
func resolveConfigDir(a Agent) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return config.ExpandTilde(a.AsylumConfigDir(), home), nil
}
