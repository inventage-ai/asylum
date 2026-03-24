package profile

import (
	"fmt"
	"slices"
	"strings"

	"github.com/inventage-ai/asylum/internal/config"
	"github.com/inventage-ai/asylum/internal/onboarding"
)

// Profile groups all language-specific concerns: installation,
// environment setup, caching, onboarding, and config defaults.
type Profile struct {
	Name              string
	Description       string
	DockerSnippet     string
	EntrypointSnippet string
	BannerLines       string               // shell commands for welcome banner version lines
	CacheDirs         map[string]string     // name → container path
	Config            config.Config
	OnboardingTasks   []onboarding.Task
	SubProfiles       map[string]*Profile
}

var registry = map[string]*Profile{}

// Register adds a top-level profile to the registry.
func Register(p *Profile) {
	registry[p.Name] = p
}

// All returns the names of all registered top-level profiles in sorted order.
func All() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

// Resolve takes a list of profile identifiers and returns a flat, deduplicated
// list of profiles in deterministic order (parents before children).
//
// Semantics:
//   - nil input means "all profiles" (backwards compatibility)
//   - empty slice means "no profiles"
//   - "java" activates java + all sub-profiles
//   - "java/maven" activates java + maven only
func Resolve(names *[]string) ([]*Profile, error) {
	if names == nil {
		return resolveAll(), nil
	}
	if len(*names) == 0 {
		return nil, nil
	}

	seen := map[string]bool{}
	var result []*Profile

	add := func(p *Profile) {
		if !seen[p.Name] {
			seen[p.Name] = true
			result = append(result, p)
		}
	}

	for _, id := range *names {
		parent, child, hasChild := strings.Cut(id, "/")
		p, ok := registry[parent]
		if !ok {
			return nil, fmt.Errorf("unknown profile %q", parent)
		}

		add(p)

		if hasChild {
			sub, ok := p.SubProfiles[child]
			if !ok {
				return nil, fmt.Errorf("unknown sub-profile %q in profile %q", child, parent)
			}
			add(sub)
		} else {
			for _, name := range sortedSubKeys(p) {
				add(p.SubProfiles[name])
			}
		}
	}

	return result, nil
}

func resolveAll() []*Profile {
	var result []*Profile
	for _, name := range All() {
		p := registry[name]
		result = append(result, p)
		for _, sub := range sortedSubKeys(p) {
			result = append(result, p.SubProfiles[sub])
		}
	}
	return result
}

func sortedSubKeys(p *Profile) []string {
	keys := make([]string, 0, len(p.SubProfiles))
	for k := range p.SubProfiles {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// AggregateCacheDirs collects CacheDirs from all provided profiles.
func AggregateCacheDirs(profiles []*Profile) map[string]string {
	dirs := map[string]string{}
	for _, p := range profiles {
		for k, v := range p.CacheDirs {
			dirs[k] = v
		}
	}
	return dirs
}

// AggregateOnboardingTasks collects OnboardingTasks from all provided profiles.
func AggregateOnboardingTasks(profiles []*Profile) []onboarding.Task {
	var tasks []onboarding.Task
	for _, p := range profiles {
		tasks = append(tasks, p.OnboardingTasks...)
	}
	return tasks
}

// AssembleDockerSnippets concatenates DockerSnippets from all provided profiles.
func AssembleDockerSnippets(profiles []*Profile) string {
	var b strings.Builder
	for _, p := range profiles {
		if p.DockerSnippet != "" {
			b.WriteString(p.DockerSnippet)
			if !strings.HasSuffix(p.DockerSnippet, "\n") {
				b.WriteByte('\n')
			}
		}
	}
	return b.String()
}

// AssembleBannerLines concatenates BannerLines from all provided profiles.
func AssembleBannerLines(profiles []*Profile) string {
	var b strings.Builder
	for _, p := range profiles {
		if p.BannerLines != "" {
			b.WriteString(p.BannerLines)
			if !strings.HasSuffix(p.BannerLines, "\n") {
				b.WriteByte('\n')
			}
		}
	}
	return b.String()
}

// AssembleEntrypointSnippets concatenates EntrypointSnippets from all provided profiles.
func AssembleEntrypointSnippets(profiles []*Profile) string {
	var b strings.Builder
	for _, p := range profiles {
		if p.EntrypointSnippet != "" {
			b.WriteString(p.EntrypointSnippet)
			if !strings.HasSuffix(p.EntrypointSnippet, "\n") {
				b.WriteByte('\n')
			}
		}
	}
	return b.String()
}
