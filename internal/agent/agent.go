package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/inventage-ai/asylum/internal/config"
)

// CmdOpts carries context from the container layer into agent command
// generation. Fields are optional; agents ignore what they don't use.
type CmdOpts struct {
	KitSkillsDir string // shared container path holding kit-provided Claude skills; empty means no skill kits active
}

type Agent interface {
	Name() string
	Binary() string
	NativeConfigDir() string
	ContainerConfigDir() string
	AsylumConfigDir() string
	EnvVars() map[string]string
	HasSession(configDir, projectPath string) bool
	Command(resume bool, extraArgs []string, opts CmdOpts) []string
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

// ResolveConfigDir returns the host-side directory that backs the agent's
// config inside the container. Which directory is used depends on the
// isolation mode: "shared" → native config dir, "project" → per-project
// dir, default ("isolated") → asylum agents dir.
func ResolveConfigDir(a Agent, isolation, containerName string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	switch isolation {
	case "shared":
		return config.ExpandTilde(a.NativeConfigDir(), home), nil
	case "project":
		return filepath.Join(home, ".asylum", "projects", containerName, a.Name()+"-config"), nil
	default:
		return config.ExpandTilde(a.AsylumConfigDir(), home), nil
	}
}
