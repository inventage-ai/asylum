package kit

import (
	"fmt"
	"slices"
	"strings"

	"github.com/inventage-ai/asylum/internal/log"
	"github.com/inventage-ai/asylum/internal/onboarding"
)

// Kit groups all concerns for a tool or language: installation,
// environment setup, caching, onboarding, and config defaults.
type Kit struct {
	Name              string
	Description       string
	DockerSnippet     string
	EntrypointSnippet string
	BannerLines       string            // shell commands for welcome banner version lines
	RulesSnippet      string            // markdown fragment for sandbox rules file
	Tools             []string          // commands this kit makes available
	CacheDirs         map[string]string  // name → container path
	OnboardingTasks   []onboarding.Task
	SubKits           map[string]*Kit
	Deps              []string // kit names this kit depends on
	DefaultOn         bool     // active unless explicitly disabled
}

var registry = map[string]*Kit{}

// Register adds a top-level kit to the registry.
func Register(k *Kit) {
	registry[k.Name] = k
}

// All returns the names of all registered top-level kits in sorted order.
func All() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

// Resolve takes a list of kit names and a set of disabled kit names,
// and returns a flat, deduplicated list of kits in deterministic order.
//
// Semantics:
//   - nil names means "all kits" (backwards compatibility)
//   - empty names slice means "no kits" (default-on kits NOT added)
//   - "java" activates java + all sub-kits
//   - "java/maven" activates java + maven only
//   - default-on kits are added when names is non-nil and non-empty
//   - disabled kits are excluded from the result
//   - dependency warnings are emitted for missing deps
func Resolve(names []string, disabled map[string]bool) ([]*Kit, error) {
	if names == nil {
		result := resolveAll()
		return filterDisabled(result, disabled), nil
	}
	if len(names) == 0 {
		return nil, nil
	}

	seen := map[string]bool{}
	var result []*Kit

	add := func(k *Kit) {
		if !seen[k.Name] && !disabled[k.Name] {
			seen[k.Name] = true
			result = append(result, k)
		}
	}

	for _, id := range names {
		parent, child, hasChild := strings.Cut(id, "/")
		k, ok := registry[parent]
		if !ok {
			return nil, fmt.Errorf("unknown kit %q", parent)
		}

		add(k)

		if hasChild {
			sub, ok := k.SubKits[child]
			if !ok {
				return nil, fmt.Errorf("unknown sub-kit %q in kit %q", child, parent)
			}
			add(sub)
		} else {
			for _, name := range sortedSubKeys(k) {
				add(k.SubKits[name])
			}
		}
	}

	// Add default-on kits that weren't explicitly listed or disabled
	for _, name := range All() {
		k := registry[name]
		if k.DefaultOn && !seen[k.Name] && !disabled[k.Name] {
			add(k)
			for _, sub := range sortedSubKeys(k) {
				add(k.SubKits[sub])
			}
		}
	}

	// Validate dependencies
	activeSet := map[string]bool{}
	for _, k := range result {
		// Extract top-level name from "parent/child"
		top, _, _ := strings.Cut(k.Name, "/")
		activeSet[top] = true
	}
	for _, k := range result {
		for _, dep := range k.Deps {
			if !activeSet[dep] {
				log.Warn("kit %q requires the %q kit which is not active", k.Name, dep)
			}
		}
	}

	return result, nil
}

func filterDisabled(kits []*Kit, disabled map[string]bool) []*Kit {
	if len(disabled) == 0 {
		return kits
	}
	var result []*Kit
	for _, k := range kits {
		if !disabled[k.Name] {
			result = append(result, k)
		}
	}
	return result
}

func resolveAll() []*Kit {
	var result []*Kit
	for _, name := range All() {
		k := registry[name]
		result = append(result, k)
		for _, sub := range sortedSubKeys(k) {
			result = append(result, k.SubKits[sub])
		}
	}
	return result
}

func sortedSubKeys(k *Kit) []string {
	keys := make([]string, 0, len(k.SubKits))
	for key := range k.SubKits {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

// AggregateTools collects Tools from all provided kits into a deduplicated,
// sorted list of "tool (kit-name)" strings.
func AggregateTools(kits []*Kit) []string {
	seen := map[string]bool{}
	var result []string
	for _, k := range kits {
		for _, tool := range k.Tools {
			if !seen[tool] {
				seen[tool] = true
				result = append(result, tool+" ("+k.Name+")")
			}
		}
	}
	return result
}

// AggregateCacheDirs collects CacheDirs from all provided kits.
func AggregateCacheDirs(kits []*Kit) map[string]string {
	dirs := map[string]string{}
	for _, k := range kits {
		for name, path := range k.CacheDirs {
			dirs[name] = path
		}
	}
	return dirs
}

// AggregateOnboardingTasks collects OnboardingTasks from all provided kits.
func AggregateOnboardingTasks(kits []*Kit) []onboarding.Task {
	var tasks []onboarding.Task
	for _, k := range kits {
		tasks = append(tasks, k.OnboardingTasks...)
	}
	return tasks
}

// AssembleDockerSnippets concatenates DockerSnippets from all provided kits.
func AssembleDockerSnippets(kits []*Kit) string {
	var b strings.Builder
	for _, k := range kits {
		if k.DockerSnippet != "" {
			b.WriteString(k.DockerSnippet)
			if !strings.HasSuffix(k.DockerSnippet, "\n") {
				b.WriteByte('\n')
			}
		}
	}
	return b.String()
}

// AssembleBannerLines concatenates BannerLines from all provided kits.
func AssembleBannerLines(kits []*Kit) string {
	var b strings.Builder
	for _, k := range kits {
		if k.BannerLines != "" {
			b.WriteString(k.BannerLines)
			if !strings.HasSuffix(k.BannerLines, "\n") {
				b.WriteByte('\n')
			}
		}
	}
	return b.String()
}

// AssembleRulesSnippets concatenates RulesSnippets from all provided kits.
func AssembleRulesSnippets(kits []*Kit) string {
	var b strings.Builder
	for _, k := range kits {
		if k.RulesSnippet != "" {
			b.WriteString(k.RulesSnippet)
			if !strings.HasSuffix(k.RulesSnippet, "\n") {
				b.WriteByte('\n')
			}
		}
	}
	return b.String()
}

// AssembleEntrypointSnippets concatenates EntrypointSnippets from all provided kits.
func AssembleEntrypointSnippets(kits []*Kit) string {
	var b strings.Builder
	for _, k := range kits {
		if k.EntrypointSnippet != "" {
			b.WriteString(k.EntrypointSnippet)
			if !strings.HasSuffix(k.EntrypointSnippet, "\n") {
				b.WriteByte('\n')
			}
		}
	}
	return b.String()
}
