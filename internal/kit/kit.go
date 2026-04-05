package kit

import (
	"fmt"
	"slices"
	"strings"

	"github.com/inventage-ai/asylum/internal/log"
	"github.com/inventage-ai/asylum/internal/onboarding"

	"gopkg.in/yaml.v3"
)

// CredentialMode determines how a kit discovers credentials.
type CredentialMode int

const (
	CredentialOff      CredentialMode = iota // credentials disabled (default)
	CredentialAuto                           // discover from project files
	CredentialExplicit                       // user-specified identifiers
)

// RunArg represents a single docker run argument with provenance.
type RunArg struct {
	Flag     string // e.g. "-p", "-v", "-e", "--privileged", "--mount", "--cap-add"
	Value    string // e.g. "10000:10000", "/host:/ctr:ro", "FOO=bar", "" for boolean flags
	Source   string // e.g. "core", "ports kit", "user config (ports)"
	Priority int    // higher wins on dedup key collision
}

// String formats a RunArg as it would appear in docker run args.
func (r RunArg) String() string {
	if r.Value == "" {
		return r.Flag
	}
	return r.Flag + " " + r.Value
}

// Priority levels for RunArg sources.
const (
	PriorityCore   = 0
	PriorityKit    = 1
	PriorityConfig = 2
	PriorityCLI    = 3
)

// Override records a RunArg that was replaced by a higher-priority one.
type Override struct {
	Replaced RunArg
	Winner   RunArg
}

// ContainerOpts is passed to a kit's ContainerFunc during container creation.
type ContainerOpts struct {
	ProjectDir    string
	ContainerName string
	HomeDir       string
	Config        interface{ PortCount() int } // avoid circular import with config package
}

// CredentialOpts is passed to a kit's CredentialFunc.
type CredentialOpts struct {
	ProjectDir    string
	HomeDir       string
	ContainerName string
	Isolation     string         // kit-specific isolation level from config
	Mode          CredentialMode
	Explicit      []string // identifiers when Mode is CredentialExplicit
}

// CredentialMount describes a credential to mount into the container.
// Use either HostPath (direct bind mount) or Content (generated file), not both.
type CredentialMount struct {
	HostPath    string // host path to bind-mount directly (file or directory)
	Content     []byte // generated file content (written to staging dir)
	FileName    string // filename for Content within staging subdir (enables directory mounts)
	Destination string // container path (e.g. ~/.m2/settings.xml)
	Writable    bool   // mount read-write instead of read-only (for tools that modify their config)
}

// Tier controls how a kit is activated and presented in the config.
type Tier int

const (
	TierDefault  Tier = iota // active when present in config; added uncommented by default
	TierAlwaysOn             // active even if absent from config
	TierOptIn                // only active if user explicitly enables in config
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
	Tier              Tier              // activation tier (TierDefault, TierAlwaysOn, TierOptIn)
	CredentialFunc    func(CredentialOpts) ([]CredentialMount, error) // optional credential provider
	CredentialLabel   string            // display label for onboarding (e.g. "Java/Maven")
	MountFunc         func(CredentialOpts) ([]CredentialMount, error) // volume mounts without credential UI
	ContainerFunc     func(ContainerOpts) ([]RunArg, error)          // docker run args contributed at container creation
	ConfigureFunc     func(versions []string, defaultVersion string) // called after config is loaded to customize snippets
	Hidden            bool              // exclude from interactive selection UIs (config TUI, kit sync prompt, sandbox rules disabled list)
	NeedsMount        bool              // kit uses mount --bind at runtime (requires SYS_ADMIN)
	DockerPriority    int               // lower = earlier in Dockerfile (stable/expensive first); 0 means default (50)
	ConfigSnippet     string            // YAML snippet for default config (indented at 2 spaces under kits:)
	ConfigNodes       []*yaml.Node      // structured key+value nodes for kits mapping (len 2: key, value)
	ConfigComment     string            // comment text for opt-in/always-on kits shown in config
}

var registry = map[string]*Kit{}

// aliases maps old kit names to their current names so that existing
// configs keep working after a kit rename.
var aliases = map[string]string{
	"browser": "agent-browser",
}

// Register adds a top-level kit to the registry.
func Register(k *Kit) {
	registry[k.Name] = k
}

// Get returns a registered kit by name, or nil if not found.
func Get(name string) *Kit {
	return registry[name]
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
		if actual, ok := aliases[parent]; ok {
			parent = actual
		}
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

	// Add always-on kits that weren't explicitly listed or disabled
	for _, name := range All() {
		k := registry[name]
		if k.Tier == TierAlwaysOn && !seen[k.Name] && !disabled[k.Name] {
			add(k)
			for _, sub := range sortedSubKeys(k) {
				add(k.SubKits[sub])
			}
		}
	}

	// Auto-activate dependencies (iterate by index to pick up transitive deps)
	activeSet := map[string]bool{}
	for _, k := range result {
		top, _, _ := strings.Cut(k.Name, "/")
		activeSet[top] = true
	}
	for i := 0; i < len(result); i++ {
		for _, dep := range result[i].Deps {
			if activeSet[dep] {
				continue
			}
			depKit, ok := registry[dep]
			if !ok || disabled[dep] {
				log.Warn("kit %q requires %q which is not available", result[i].Name, dep)
				continue
			}
			add(depKit)
			for _, sub := range sortedSubKeys(depKit) {
				add(depKit.SubKits[sub])
			}
			activeSet[dep] = true
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

// AggregateContainerArgs calls ContainerFunc on each kit that has one and
// collects all returned RunArgs. Errors are logged as warnings and skipped.
func AggregateContainerArgs(kits []*Kit, opts ContainerOpts) []RunArg {
	var result []RunArg
	for _, k := range kits {
		if k.ContainerFunc == nil {
			continue
		}
		args, err := k.ContainerFunc(opts)
		if err != nil {
			log.Warn("container args for %s: %v", k.Name, err)
			continue
		}
		result = append(result, args...)
	}
	return result
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

// AnyNeedsMount reports whether any kit uses mount --bind at runtime.
func AnyNeedsMount(kits []*Kit) bool {
	for _, k := range kits {
		if k.NeedsMount {
			return true
		}
	}
	return false
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

// ScalarNode creates a yaml scalar node with an optional line comment.
func ScalarNode(value, comment string) *yaml.Node {
	n := &yaml.Node{Kind: yaml.ScalarNode, Value: value, Tag: "!!str"}
	if comment != "" {
		n.LineComment = comment
	}
	return n
}

// MappingNode creates a yaml mapping node from key-value pairs.
func MappingNode(content ...*yaml.Node) *yaml.Node {
	return &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map", Content: content}
}

// SeqNode creates a yaml sequence node from values.
func SeqNode(values ...string) *yaml.Node {
	content := make([]*yaml.Node, len(values))
	for i, v := range values {
		content[i] = &yaml.Node{Kind: yaml.ScalarNode, Value: v, Tag: "!!str"}
	}
	return &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq", Content: content}
}

// BoolNode creates a yaml scalar node with a boolean value.
func BoolNode(v bool) *yaml.Node {
	s := "false"
	if v {
		s = "true"
	}
	return &yaml.Node{Kind: yaml.ScalarNode, Value: s, Tag: "!!bool"}
}

// IntNode creates a yaml scalar node with an integer value.
func IntNode(v int) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprintf("%d", v), Tag: "!!int"}
}

// configNodes builds the standard [key, value] node pair for a kit's config entry.
// If content is nil, the value is an empty mapping (displayed as `name: {}`).
// Comments are placed on the value node for empty mappings (yaml.v3 produces
// invalid YAML when a LineComment sits on a key whose value is block-style empty).
func configNodes(name, comment string, content []*yaml.Node) []*yaml.Node {
	key := ScalarNode(name, "")
	val := MappingNode(content...)
	if len(content) == 0 && comment != "" {
		val.LineComment = comment
	} else if comment != "" {
		key.LineComment = comment
	}
	return []*yaml.Node{key, val}
}

// CredentialCapableKits returns all registered kits (including sub-kits) that
// have a non-nil CredentialFunc, in sorted order by name.
func CredentialCapableKits() []*Kit {
	var result []*Kit
	for _, name := range All() {
		k := registry[name]
		if k.CredentialFunc != nil {
			result = append(result, k)
		}
		for _, sub := range sortedSubKeys(k) {
			sk := k.SubKits[sub]
			if sk.CredentialFunc != nil {
				result = append(result, sk)
			}
		}
	}
	return result
}

// AssembleConfigSnippets returns all registered kits' ConfigSnippets in sorted
// order, with commented-out snippets grouped after active ones.
func AssembleConfigSnippets() string {
	var active, commented strings.Builder
	for _, name := range All() {
		k := registry[name]
		if k.ConfigSnippet == "" {
			continue
		}
		trimmed := strings.TrimSpace(k.ConfigSnippet)
		if strings.HasPrefix(trimmed, "#") {
			commented.WriteString("\n")
			commented.WriteString(k.ConfigSnippet)
		} else {
			active.WriteString("\n")
			active.WriteString(k.ConfigSnippet)
		}
	}
	var b strings.Builder
	b.WriteString(active.String())
	if commented.Len() > 0 {
		b.WriteString(commented.String())
	}
	return b.String()
}

// AssembleRulesSnippets concatenates RulesSnippets from all provided kits,
// separated by blank lines.
func AssembleRulesSnippets(kits []*Kit) string {
	var b strings.Builder
	first := true
	for _, k := range kits {
		if k.RulesSnippet != "" {
			if !first {
				b.WriteByte('\n')
			}
			b.WriteString(k.RulesSnippet)
			if !strings.HasSuffix(k.RulesSnippet, "\n") {
				b.WriteByte('\n')
			}
			first = false
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
