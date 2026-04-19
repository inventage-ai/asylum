package config

import (
	"crypto/sha256"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"

	"github.com/inventage-ai/asylum/internal/kit"
	"gopkg.in/yaml.v3"
)

// Credentials supports three YAML forms: "auto" (string), false/absent (off),
// or a list of kit-specific identifiers.
type Credentials struct {
	Auto     bool     // true when "auto"
	Explicit []string // non-nil when a list is provided
}

func (c *Credentials) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		if value.Tag == "!!bool" {
			// credentials: false → off (zero value)
			return nil
		}
		if value.Value == "auto" {
			c.Auto = true
			return nil
		}
		return fmt.Errorf("credentials: unsupported value %q (use auto, false, or a list)", value.Value)
	case yaml.SequenceNode:
		var ids []string
		if err := value.Decode(&ids); err != nil {
			return err
		}
		c.Explicit = ids
		return nil
	default:
		return fmt.Errorf("credentials: expected string, bool, or list")
	}
}

// KitConfig holds per-kit configuration options.
// Fields tagged merge:"concat" are accumulated (base + overlay) during config
// merge. All other fields use last-wins semantics (overlay replaces base when
// non-zero). Adding a new field requires no merge code changes.
type KitConfig struct {
	Disabled            *bool        `yaml:"disabled,omitempty"`
	Versions            []string     `yaml:"versions,omitempty"`
	DefaultVersion      string       `yaml:"default-version,omitempty"`
	Packages            []string     `yaml:"packages,omitempty" merge:"concat"`
	ShadowNodeModules   *bool        `yaml:"shadow-node-modules,omitempty"`
	Onboarding          *bool        `yaml:"onboarding,omitempty"`
	TabTitle            string       `yaml:"tab-title,omitempty"`
	AllowAgentTermTitle *bool        `yaml:"allow-agent-terminal-title,omitempty"`
	Build               []string     `yaml:"build,omitempty" merge:"concat"`
	Count               int          `yaml:"count,omitempty"`
	Credentials         *Credentials `yaml:"credentials,omitempty"`
	Isolation           string       `yaml:"isolation,omitempty"`
}

// AgentConfig holds per-agent configuration.
type AgentConfig struct {
	Config string `yaml:"config,omitempty"` // shared, isolated, project
}

// AgentIsolation returns the config isolation level for the named agent.
// Returns empty string if not set (triggers first-run prompt).
func (c Config) AgentIsolation(agentName string) string {
	if c.Agents == nil {
		return ""
	}
	ac, ok := c.Agents[agentName]
	if !ok || ac == nil {
		return ""
	}
	return ac.Config
}

type Config struct {
	Version        string                  `yaml:"version,omitempty"`
	Agent          string                  `yaml:"agent,omitempty"`
	ReleaseChannel string                  `yaml:"release-channel,omitempty"`
	Kits           map[string]*KitConfig   `yaml:"kits,omitempty"`
	Agents         map[string]*AgentConfig `yaml:"agents,omitempty"`
	Ports          []string                `yaml:"ports,omitempty"`
	Volumes        []string                `yaml:"volumes,omitempty"`
	Env            map[string]string       `yaml:"env,omitempty"`
}

// KitActive returns true if the named kit is present and not disabled.
func (c Config) KitActive(name string) bool {
	if c.Kits == nil {
		return false
	}
	kc, ok := c.Kits[name]
	if !ok {
		return false
	}
	if kc != nil && kc.Disabled != nil && *kc.Disabled {
		return false
	}
	return true
}

// DisabledKits returns a set of kit names that are explicitly disabled.
func (c Config) DisabledKits() map[string]bool {
	if c.Kits == nil {
		return nil
	}
	disabled := map[string]bool{}
	for name, kc := range c.Kits {
		if kc != nil && kc.Disabled != nil && *kc.Disabled {
			disabled[name] = true
		}
	}
	if len(disabled) == 0 {
		return nil
	}
	return disabled
}

// KitOption returns the KitConfig for the named kit, or nil if not present.
func (c Config) KitOption(name string) *KitConfig {
	if c.Kits == nil {
		return nil
	}
	return c.Kits[name]
}

// KitSnippetConfig returns a kit.SnippetConfig for the named kit, or nil.
func (c Config) KitSnippetConfig(name string) *kit.SnippetConfig {
	kc := c.KitOption(name)
	if kc == nil {
		return nil
	}
	return &kit.SnippetConfig{
		Versions:       kc.Versions,
		DefaultVersion: kc.DefaultVersion,
	}
}

// SSHIsolation returns the SSH kit isolation level.
// Returns "isolated" when not configured.
func (c Config) SSHIsolation() string {
	if kc := c.KitOption("ssh"); kc != nil && kc.Isolation != "" {
		return kc.Isolation
	}
	return "isolated"
}

// AgentActive returns true if the named agent is present in the config.
func (c Config) AgentActive(name string) bool {
	if c.Agents == nil {
		return false
	}
	_, ok := c.Agents[name]
	return ok
}

// KitNames returns sorted kit names from the map, or nil if Kits is nil.
// An empty but non-nil Kits map returns an empty (non-nil) slice, which
// kit.Resolve treats as "no kits" — distinct from nil which means "all kits".
func (c Config) KitNames() []string {
	if c.Kits == nil {
		return nil
	}
	if len(c.Kits) == 0 {
		return []string{}
	}
	return slices.Sorted(maps.Keys(c.Kits))
}

// AgentNames returns sorted agent names from the map, or nil if Agents is nil.
func (c Config) AgentNames() []string {
	if c.Agents == nil {
		return nil
	}
	return slices.Sorted(maps.Keys(c.Agents))
}

// TabTitle returns the tab title from the title kit, or empty string.
func (c Config) TabTitle() string {
	if kc := c.KitOption("title"); kc != nil {
		return kc.TabTitle
	}
	return ""
}

// ShadowNodeModulesOff returns true if shadow-node-modules is explicitly disabled.
func (c Config) ShadowNodeModulesOff() bool {
	if kc := c.KitOption("node"); kc != nil && kc.ShadowNodeModules != nil {
		return !*kc.ShadowNodeModules
	}
	return false
}

// AllowAgentTermTitle returns true if agent terminal title is allowed.
func (c Config) AllowAgentTermTitle() bool {
	if kc := c.KitOption("title"); kc != nil && kc.AllowAgentTermTitle != nil {
		return *kc.AllowAgentTermTitle
	}
	return false
}

// OnboardingEnabled returns whether onboarding is enabled for the given kit.
func (c Config) OnboardingEnabled(kitName string) bool {
	if kc := c.KitOption(kitName); kc != nil && kc.Onboarding != nil {
		return *kc.Onboarding
	}
	return false
}

// KitCredentialMode returns the credential mode for the named kit.
// Returns "" if never configured, "none" if explicitly declined,
// "auto" for automatic credentials, or "explicit" for a user-provided list.
func (c Config) KitCredentialMode(kitName string) string {
	kc := c.KitOption(kitName)
	if kc == nil || kc.Credentials == nil {
		return ""
	}
	if kc.Credentials.Auto {
		return "auto"
	}
	if kc.Credentials.Explicit != nil {
		return "explicit"
	}
	return "none"
}

// KitCredentialExplicit returns the explicit credential identifiers for the named kit.
func (c Config) KitCredentialExplicit(kitName string) []string {
	kc := c.KitOption(kitName)
	if kc == nil || kc.Credentials == nil {
		return nil
	}
	return kc.Credentials.Explicit
}

// Packages returns all packages for a given kit, or nil.
func (c Config) KitPackages(kitName string) []string {
	if kc := c.KitOption(kitName); kc != nil {
		return kc.Packages
	}
	return nil
}

// PortCount returns the number of ports to allocate from the ports kit, default 5.
func (c Config) PortCount() int {
	if kc := c.KitOption("ports"); kc != nil && kc.Count > 0 {
		return kc.Count
	}
	return 5
}

type CLIFlags struct {
	Agent   string
	Kits    *[]string
	Agents  *[]string
	Ports   []string
	Volumes []string
	Env     map[string]string
	Java    string
}

type Volume struct {
	Host      string
	Container string
	Options   string
}

func Load(projectDir string, flags CLIFlags, kitSnippets string) (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, fmt.Errorf("home dir: %w", err)
	}

	var cfg Config

	// Load global config
	globalPath := filepath.Join(home, ".asylum", "config.yaml")
	if NeedsMigration(globalPath) {
		if err := MigrateV1ToV2(globalPath, kitSnippets); err != nil {
			return Config{}, fmt.Errorf("migrate %s: %w", globalPath, err)
		}
	}
	if layer, err := LoadFile(globalPath); err != nil {
		return Config{}, fmt.Errorf("%s: %w", globalPath, err)
	} else {
		cfg = Merge(cfg, layer)
	}

	// Apply .tool-versions after global config but before project config,
	// so project-local .asylum can override it.
	if java := readToolVersionsJava(projectDir); java != "" {
		setJavaVersion(&cfg, java)
	}

	// Load project configs (override .tool-versions)
	for _, path := range []string{
		filepath.Join(projectDir, ".asylum"),
		filepath.Join(projectDir, ".asylum.local"),
	} {
		if NeedsMigration(path) {
			if err := MigrateV1ToV2(path, kitSnippets); err != nil {
				return Config{}, fmt.Errorf("migrate %s: %w", path, err)
			}
		}
		layer, err := LoadFile(path)
		if err != nil {
			return Config{}, fmt.Errorf("%s: %w", path, err)
		}
		cfg = Merge(cfg, layer)
	}

	cfg = applyFlags(cfg, flags)
	return cfg, nil
}

// LoadFile reads a single config file and returns its parsed Config.
// Returns a zero Config if the file doesn't exist or is a directory.
func LoadFile(path string) (Config, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, nil
		}
		return Config{}, err
	}
	if info.IsDir() {
		return Config{}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Merge applies overlay values on top of base, following per-field merge rules.
func Merge(base, overlay Config) Config {
	result := base

	if overlay.Version != "" {
		result.Version = overlay.Version
	}
	if overlay.Agent != "" {
		result.Agent = overlay.Agent
	}
	if overlay.ReleaseChannel != "" {
		result.ReleaseChannel = overlay.ReleaseChannel
	}

	// Kits: per-key deep merge
	if overlay.Kits != nil {
		if result.Kits == nil {
			result.Kits = make(map[string]*KitConfig, len(overlay.Kits))
		}
		for name, overKC := range overlay.Kits {
			result.Kits[name] = mergeKitConfig(result.Kits[name], overKC)
		}
	}
	// Agents: per-key merge
	if overlay.Agents != nil {
		if result.Agents == nil {
			result.Agents = make(map[string]*AgentConfig, len(overlay.Agents))
		}
		for name, ac := range overlay.Agents {
			if _, ok := result.Agents[name]; !ok {
				result.Agents[name] = ac
			}
		}
	}

	result.Ports = slices.Concat(base.Ports, overlay.Ports)
	result.Volumes = slices.Concat(base.Volumes, overlay.Volumes)

	if overlay.Env != nil {
		merged := make(map[string]string, len(base.Env)+len(overlay.Env))
		maps.Copy(merged, base.Env)
		maps.Copy(merged, overlay.Env)
		result.Env = merged
	}

	return result
}

// mergeKitConfig deep-merges two KitConfig values using struct tags.
// Fields tagged merge:"concat" are accumulated (base + overlay).
// All other fields use last-wins: overlay replaces base when non-zero.
func mergeKitConfig(base, overlay *KitConfig) *KitConfig {
	if base == nil {
		return overlay
	}
	if overlay == nil {
		return base
	}
	result := *base
	rv := reflect.ValueOf(&result).Elem()
	ov := reflect.ValueOf(overlay).Elem()
	for i := range rv.NumField() {
		of := ov.Field(i)
		tag := rv.Type().Field(i).Tag.Get("merge")
		if tag == "concat" {
			rv.Field(i).Set(reflect.AppendSlice(rv.Field(i), of))
		} else if !of.IsZero() {
			rv.Field(i).Set(of)
		}
	}
	return &result
}

func applyFlags(cfg Config, flags CLIFlags) Config {
	if flags.Agent != "" {
		cfg.Agent = flags.Agent
	}
	if flags.Kits != nil {
		m := make(map[string]*KitConfig, len(*flags.Kits))
		for _, name := range *flags.Kits {
			m[name] = nil
		}
		cfg.Kits = m
	}
	if flags.Agents != nil {
		m := make(map[string]*AgentConfig, len(*flags.Agents))
		for _, name := range *flags.Agents {
			m[name] = nil
		}
		cfg.Agents = m
	}
	if flags.Java != "" {
		setJavaVersion(&cfg, flags.Java)
	}
	if flags.Env != nil {
		if cfg.Env == nil {
			cfg.Env = make(map[string]string, len(flags.Env))
		}
		maps.Copy(cfg.Env, flags.Env)
	}
	cfg.Ports = slices.Concat(cfg.Ports, flags.Ports)
	cfg.Volumes = slices.Concat(cfg.Volumes, flags.Volumes)
	return cfg
}

// setJavaVersion ensures the java kit entry exists and sets its DefaultVersion.
func setJavaVersion(cfg *Config, version string) {
	if cfg.Kits == nil {
		cfg.Kits = map[string]*KitConfig{}
	}
	if cfg.Kits["java"] == nil {
		cfg.Kits["java"] = &KitConfig{}
	}
	cfg.Kits["java"].DefaultVersion = version
}

func readToolVersionsJava(projectDir string) string {
	data, err := os.ReadFile(filepath.Join(projectDir, ".tool-versions"))
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "java" {
			return fields[1]
		}
	}
	return ""
}

var mountOptions = map[string]bool{
	"ro": true, "rw": true, "z": true, "Z": true,
	"shared": true, "slave": true, "private": true,
	"rshared": true, "rslave": true, "rprivate": true,
	"nocopy": true, "consistent": true, "cached": true, "delegated": true,
}

// legacyHome is the old hardcoded container home directory.
// Paths starting with this prefix are transparently rewritten
// to the current home directory so that existing configs and
// user-specified volume mounts keep working after the switch
// to host-aligned home directories.
const legacyHome = "/home/claude/"

// ExpandTilde replaces a leading ~/ or the legacy /home/claude/ prefix
// with homeDir. A bare "~" returns homeDir.
func ExpandTilde(path, homeDir string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(homeDir, path[2:])
	}
	if path == "~" {
		return homeDir
	}
	if strings.HasPrefix(path, legacyHome) {
		return filepath.Join(homeDir, path[len(legacyHome):])
	}
	if path == "/home/claude" {
		return homeDir
	}
	return path
}

func ParseVolume(raw string, homeDir string) (Volume, error) {
	if strings.TrimSpace(raw) == "" {
		return Volume{}, fmt.Errorf("empty volume specification")
	}

	parts := strings.Split(raw, ":")

	switch len(parts) {
	case 1:
		host := ExpandTilde(parts[0], homeDir)
		return Volume{Host: host, Container: host}, nil
	case 2:
		if mountOptions[parts[1]] {
			if parts[0] == "" {
				return Volume{}, fmt.Errorf("empty path in volume %q", raw)
			}
			host := ExpandTilde(parts[0], homeDir)
			return Volume{Host: host, Container: host, Options: parts[1]}, nil
		}
		if parts[0] == "" || parts[1] == "" {
			return Volume{}, fmt.Errorf("empty path in volume %q", raw)
		}
		return Volume{Host: ExpandTilde(parts[0], homeDir), Container: ExpandTilde(parts[1], homeDir)}, nil
	default:
		if parts[0] == "" || parts[1] == "" {
			return Volume{}, fmt.Errorf("empty path in volume %q", raw)
		}
		for _, opt := range parts[2:] {
			if !mountOptions[opt] {
				return Volume{}, fmt.Errorf("invalid mount option %q in volume %q", opt, raw)
			}
		}
		return Volume{Host: ExpandTilde(parts[0], homeDir), Container: ExpandTilde(parts[1], homeDir), Options: strings.Join(parts[2:], ":")}, nil
	}
}

// ConfigHash computes a deterministic hash of the config for detecting drift
// against a running container. It serializes the full config to YAML (which
// sorts map keys deterministically) after clearing non-runtime fields and
// normalizing order-insensitive lists. New fields are included automatically.
func ConfigHash(cfg Config) string {
	// Clear fields that don't affect the running container.
	cfg.Version = ""
	cfg.Agent = ""
	cfg.ReleaseChannel = ""
	cfg.Agents = nil

	// Sort order-insensitive lists so the hash is stable regardless of YAML order.
	slices.Sort(cfg.Volumes)
	slices.Sort(cfg.Ports)

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", sha256.Sum256(data))
}
