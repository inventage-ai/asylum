package agent

import (
	"fmt"
	"slices"
	"strings"

	"github.com/inventage-ai/asylum/internal/log"
)

// AgentInstall holds build-time installation metadata for an agent CLI.
type AgentInstall struct {
	Name          string   // matches Agent.Name()
	DockerSnippet string   // Dockerfile RUN instructions
	KitDeps   []string // kit names this agent needs (e.g., ["node"])
	BannerLine    string   // shell command for welcome banner
}

var installs = map[string]*AgentInstall{}

// RegisterInstall adds an agent install definition to the registry.
func RegisterInstall(i *AgentInstall) {
	installs[i.Name] = i
}

// AllInstallNames returns sorted names of all registered agent installs.
func AllInstallNames() []string {
	names := make([]string, 0, len(installs))
	for name := range installs {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

// ResolveInstalls returns agent installs for the given agent map.
// nil map defaults to claude-only; empty map means none.
// Also accepts a string slice (names) for CLI flag compatibility.
// Emits warnings for agents whose KitDeps are not in activeKits.
func ResolveInstalls(agents map[string]bool, activeKits []string) ([]*AgentInstall, error) {
	var selected []string
	if agents == nil {
		selected = []string{"claude"}
	} else {
		selected = make([]string, 0, len(agents))
		for name := range agents {
			selected = append(selected, name)
		}
		slices.Sort(selected)
	}

	if len(selected) == 0 {
		return nil, nil
	}

	kitSet := map[string]bool{}
	for _, p := range activeKits {
		kitSet[p] = true
	}

	var result []*AgentInstall
	seen := map[string]bool{}
	for _, name := range selected {
		if seen[name] {
			continue
		}
		seen[name] = true
		i, ok := installs[name]
		if !ok {
			return nil, fmt.Errorf("unknown agent %q", name)
		}
		for _, dep := range i.KitDeps {
			if !kitSet[dep] {
				log.Warn("agent %q requires the %q kit which is not active", name, dep)
			}
		}
		result = append(result, i)
	}
	return result, nil
}

// AssembleAgentSnippets concatenates DockerSnippets from the given installs.
func AssembleAgentSnippets(installs []*AgentInstall) string {
	return joinFields(installs, func(i *AgentInstall) string { return i.DockerSnippet })
}

// AssembleAgentBannerLines concatenates BannerLines from the given installs.
func AssembleAgentBannerLines(installs []*AgentInstall) string {
	return joinFields(installs, func(i *AgentInstall) string { return i.BannerLine })
}

// joinFields concatenates non-empty field values, ensuring each ends with a newline.
func joinFields(installs []*AgentInstall, field func(*AgentInstall) string) string {
	var b strings.Builder
	for _, i := range installs {
		s := field(i)
		if s != "" {
			b.WriteString(s)
			if !strings.HasSuffix(s, "\n") {
				b.WriteByte('\n')
			}
		}
	}
	return b.String()
}
