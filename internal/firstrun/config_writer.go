package firstrun

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/inventage-ai/asylum/internal/agent"
	"github.com/inventage-ai/asylum/internal/kit"
)

// Choices captures the first-run wizard's outcome for use by the config
// writer. Each map is keyed by registered name; true means active, false
// means commented-out in the generated file.
type Choices struct {
	DefaultAgent string
	Agents       map[string]bool
	Kits         map[string]bool
}

// WriteConfig writes a complete `~/.asylum/config.yaml` reflecting the
// wizard's choices. Active agents/kits appear uncommented; deselected ones
// appear as comments so users can re-enable them by hand. Non-selectable
// kits (TierAlwaysOn or Hidden) keep their default presentation.
func WriteConfig(path string, c Choices) error {
	return os.WriteFile(path, []byte(BuildConfig(c)), 0644)
}

// BuildConfig returns the full config text for first-run output. Exported
// for testability.
func BuildConfig(c Choices) string {
	var b strings.Builder
	b.WriteString(buildHeader(c))
	b.WriteString(buildAgentsBlock(c))
	b.WriteString(buildKitsHeader())
	b.WriteString(buildKitsBlock(c))
	b.WriteString(configFooter)
	return b.String()
}

func buildHeader(c Choices) string {
	names := agentPickerNames()
	return fmt.Sprintf(`version: "0.2"

# Release channel for self-update (stable, dev)
release-channel: stable

# Agent to start by default (one of: %s)
agent: %s

`, strings.Join(names, ", "), c.DefaultAgent)
}

func buildAgentsBlock(c Choices) string {
	var b strings.Builder
	b.WriteString(`# Agent CLIs to install in the container image.
# Remove or comment out agents you don't use to speed up image builds.
agents:
`)
	// Active agents first (sorted), then commented (sorted).
	all := agentPickerNames()
	var active, commented []string
	for _, name := range all {
		if c.Agents[name] {
			active = append(active, name)
		} else {
			commented = append(commented, name)
		}
	}
	for _, name := range active {
		b.WriteString("  " + name + ":\n")
	}
	for _, name := range commented {
		b.WriteString("  # " + name + ":\n")
	}
	return b.String()
}

func buildKitsHeader() string {
	return `
# Kits configure language toolchains and tools installed in the container.
# A kit is active when its key is present (even with no options).
# Comment out or remove a kit to disable it entirely.
kits:`
}

func buildKitsBlock(c Choices) string {
	var active, commented strings.Builder
	for _, name := range kit.All() {
		k := kit.Get(name)
		if k == nil || k.ConfigSnippet == "" {
			continue
		}
		snippet, wantActive := pickKitSnippet(k, c)
		if wantActive {
			active.WriteString("\n")
			active.WriteString(snippet)
		} else {
			commented.WriteString("\n")
			commented.WriteString(snippet)
		}
	}
	return active.String() + commented.String()
}

// pickKitSnippet returns the snippet text and whether it belongs in the
// active group. Snippets are transformed only when the user's selection
// differs from the kit's authored default — keeping format intact in the
// common "no change" path.
func pickKitSnippet(k *kit.Kit, c Choices) (string, bool) {
	authoredActive := kit.SnippetIsActive(k.ConfigSnippet)
	if !isSelectable(k) {
		return k.ConfigSnippet, authoredActive
	}
	selected := c.Kits[k.Name]
	if selected == authoredActive {
		return k.ConfigSnippet, selected
	}
	if selected {
		return kit.UncommentSnippet(k.ConfigSnippet), true
	}
	return kit.CommentSnippet(k.ConfigSnippet), false
}

// isSelectable reports whether the kit can be toggled by the wizard's kit
// multi-select. Always-on and hidden kits stay out of the picker.
func isSelectable(k *kit.Kit) bool {
	return k.Tier != kit.TierAlwaysOn && !k.Hidden
}

// agentPickerNames returns all registered agent install names except the
// `echo` test stub, sorted.
func agentPickerNames() []string {
	all := agent.AllInstallNames()
	out := make([]string, 0, len(all))
	for _, n := range all {
		if n == "echo" {
			continue
		}
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

const configFooter = `
# Port forwarding (host:container or just port for same on both sides)
# ports:
#   - "3000"
#   - "8080:80"

# Additional volume mounts
# Supports: /path, /host:/container, /host:/container:ro, ~/path
# volumes:
#   - ~/shared-data:/data
#   - ~/.aws

# Environment variables
# env:
#   GITHUB_TOKEN: ghp_xxx
#   NODE_ENV: development
`
