package agent

import (
	"fmt"
	"slices"
	"strings"

	"github.com/inventage-ai/asylum/internal/log"
)

// AgentInstall holds build-time installation metadata for an agent CLI.
// Agents are not part of the priority-ordered, state-tracked source set:
// they are always assembled as a block after all kit snippets (see
// image.assembleDockerfile), so agent version bumps never invalidate kit layers.
type AgentInstall struct {
	Name          string   // matches Agent.Name()
	DockerSnippet string   // Dockerfile RUN instructions
	KitDeps       []string // kit names this agent needs (e.g., ["node"])
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

// AssembleVersionedAgentSnippets concatenates version-pinned DockerSnippets from
// the given installs, in registration order. A nil version map yields the
// unversioned snippets. This block is emitted after all kit snippets.
func AssembleVersionedAgentSnippets(installs []*AgentInstall, vm map[string]string) string {
	return joinFields(installs, func(i *AgentInstall) string { return i.VersionedSnippet(vm) })
}

// AssembleAgentBannerLines concatenates BannerLines from the given installs.
func AssembleAgentBannerLines(installs []*AgentInstall) string {
	return joinFields(installs, func(i *AgentInstall) string { return i.BannerLine })
}

// VersionedSnippet returns a Dockerfile snippet with a version ARG and
// modified RUN instruction. Returns the original snippet if no version
// is available for the agent.
func (i *AgentInstall) VersionedSnippet(vm map[string]string) string {
	ver, ok := vm[i.Name]
	if !ok {
		return i.DockerSnippet
	}

	switch i.Name {
	case "gemini", "codex", "pi":
		// npm agents: append @${VERSION} to package name within RUN
		arg := versionArgName(i.Name)
		snippet := replaceInRun(i.DockerSnippet, fmt.Sprintf(`npm install -g @%s/%s`, agentPrefix(i.Name), pkgName(i.Name)), fmt.Sprintf(`npm install -g @%s/%s@${%s}`, agentPrefix(i.Name), pkgName(i.Name), arg))
		return fmt.Sprintf("ARG %s=%s\n%s\n", arg, ver, snippet)
	case "claude":
		// Claude install.sh accepts a version argument via bash -s --
		snippet := strings.Replace(i.DockerSnippet, "| bash", "| bash -s -- ${CLAUDE_VERSION}", 1)
		return fmt.Sprintf("ARG CLAUDE_VERSION=%s\n%s\n", ver, snippet)
	case "copilot":
		// Copilot install script accepts VERSION env var
		snippet := strings.Replace(i.DockerSnippet, "RUN curl", "RUN VERSION=${COPILOT_VERSION} curl", 1)
		return fmt.Sprintf("ARG COPILOT_VERSION=%s\n%s\n", ver, snippet)
	case "opencode":
		// Opencode install script accepts --version flag via bash -s --
		snippet := strings.Replace(i.DockerSnippet, "| bash", "| bash -s -- --version ${OPENCODE_VERSION}", 1)
		return fmt.Sprintf("ARG OPENCODE_VERSION=%s\n%s\n", ver, snippet)
	default:
		// Echo and unknown agents: no version injection
		return i.DockerSnippet
	}
}

// agentPrefix returns the npm scope prefix for an agent (empty for root scope).
func agentPrefix(name string) string {
	switch name {
	case "gemini":
		return "google"
	case "codex":
		return "openai"
	case "pi":
		return "earendil-works"
	default:
		return ""
	}
}

// pkgName returns the package name part (after the scope) for npm agents.
func pkgName(name string) string {
	switch name {
	case "gemini":
		return "gemini-cli"
	case "codex":
		return "codex"
	case "pi":
		return "pi-coding-agent"
	default:
		return name
	}
}

// replaceInRun finds the install command within the Docker snippet and
// replaces it with the versioned variant. It preserves surrounding whitespace
// and multi-line structure (comments, continuation lines).
func replaceInRun(snippet, old, new string) string {
	lines := strings.Split(snippet, "\n")
	for i, line := range lines {
		if strings.Contains(line, old) {
			lines[i] = strings.Replace(line, old, new, 1)
			break
		}
	}
	return strings.Join(lines, "\n")
}

// versionArgName returns the uppercase version ARG name for an npm agent.
func versionArgName(name string) string {
	switch name {
	case "gemini":
		return "GEMINI_VERSION"
	case "codex":
		return "CODEX_VERSION"
	case "pi":
		return "PI_VERSION"
	default:
		return strings.ToUpper(name) + "_VERSION"
	}
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
