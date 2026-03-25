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
	ProfileDeps   []string // profile names this agent needs (e.g., ["node"])
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

// ResolveInstalls returns agent installs for the given names.
// nil defaults to ["claude"]; empty slice means none.
// Emits warnings for agents whose ProfileDeps are not in activeProfiles.
func ResolveInstalls(names *[]string, activeProfiles []string) ([]*AgentInstall, error) {
	var selected []string
	if names == nil {
		selected = []string{"claude"}
	} else {
		selected = *names
	}

	if len(selected) == 0 {
		return nil, nil
	}

	profileSet := map[string]bool{}
	for _, p := range activeProfiles {
		profileSet[p] = true
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
		for _, dep := range i.ProfileDeps {
			if !profileSet[dep] {
				log.Warn("agent %q requires the %q profile which is not active", name, dep)
			}
		}
		result = append(result, i)
	}
	return result, nil
}

// AssembleAgentSnippets concatenates DockerSnippets from the given installs.
func AssembleAgentSnippets(installs []*AgentInstall) string {
	var b strings.Builder
	for _, i := range installs {
		if i.DockerSnippet != "" {
			b.WriteString(i.DockerSnippet)
			if !strings.HasSuffix(i.DockerSnippet, "\n") {
				b.WriteByte('\n')
			}
		}
	}
	return b.String()
}

// AssembleAgentBannerLines concatenates BannerLines from the given installs.
func AssembleAgentBannerLines(installs []*AgentInstall) string {
	var b strings.Builder
	for _, i := range installs {
		if i.BannerLine != "" {
			b.WriteString(i.BannerLine)
			if !strings.HasSuffix(i.BannerLine, "\n") {
				b.WriteByte('\n')
			}
		}
	}
	return b.String()
}
