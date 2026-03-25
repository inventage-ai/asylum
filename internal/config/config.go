package config

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// KitConfig holds per-kit configuration options.
type KitConfig struct {
	Versions            []string `yaml:"versions,omitempty"`
	DefaultVersion      string   `yaml:"default-version,omitempty"`
	Packages            []string `yaml:"packages,omitempty"`
	ShadowNodeModules   *bool    `yaml:"shadow-node-modules,omitempty"`
	Onboarding          *bool    `yaml:"onboarding,omitempty"`
	TabTitle            string   `yaml:"tab-title,omitempty"`
	AllowAgentTermTitle *bool    `yaml:"allow-agent-terminal-title,omitempty"`
	Build               []string `yaml:"build,omitempty"`
	Start               []string `yaml:"start,omitempty"`
}

// AgentConfig holds per-agent configuration (empty for now, placeholder for future use).
type AgentConfig struct{}

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

// KitActive returns true if the named kit is present in the config.
func (c Config) KitActive(name string) bool {
	if c.Kits == nil {
		return false
	}
	_, ok := c.Kits[name]
	return ok
}

// KitOption returns the KitConfig for the named kit, or nil if not present.
func (c Config) KitOption(name string) *KitConfig {
	if c.Kits == nil {
		return nil
	}
	return c.Kits[name]
}

// KitBool returns the value of a *bool kit option, defaulting to the given value if nil.
func KitBool(b *bool, defaultVal bool) bool {
	if b == nil {
		return defaultVal
	}
	return *b
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
func (c Config) KitNames() []string {
	if c.Kits == nil {
		return nil
	}
	names := make([]string, 0, len(c.Kits))
	for k := range c.Kits {
		names = append(names, k)
	}
	slices.Sort(names)
	return names
}

// AgentNames returns sorted agent names from the map, or nil if Agents is nil.
func (c Config) AgentNames() []string {
	if c.Agents == nil {
		return nil
	}
	names := make([]string, 0, len(c.Agents))
	for k := range c.Agents {
		names = append(names, k)
	}
	slices.Sort(names)
	return names
}

// JavaVersion returns the effective java version from the java kit's DefaultVersion.
func (c Config) JavaVersion() string {
	if kc := c.KitOption("java"); kc != nil {
		return kc.DefaultVersion
	}
	return ""
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

// Packages returns all packages for a given kit, or nil.
func (c Config) KitPackages(kitName string) []string {
	if kc := c.KitOption(kitName); kc != nil {
		return kc.Packages
	}
	return nil
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

func Load(projectDir string, flags CLIFlags) (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, fmt.Errorf("home dir: %w", err)
	}

	var cfg Config
	for _, path := range []string{
		filepath.Join(home, ".asylum", "config.yaml"),
		filepath.Join(projectDir, ".asylum"),
		filepath.Join(projectDir, ".asylum.local"),
	} {
		if NeedsMigration(path) {
			if err := MigrateV1ToV2(path); err != nil {
				return Config{}, fmt.Errorf("migrate %s: %w", path, err)
			}
		}
		layer, err := LoadFile(path)
		if err != nil {
			return Config{}, fmt.Errorf("%s: %w", path, err)
		}
		cfg = Merge(cfg, layer)
	}

	// Read Java version from .tool-versions if not already set
	if java := readToolVersionsJava(projectDir); java != "" && cfg.JavaVersion() == "" {
		if cfg.Kits == nil {
			cfg.Kits = map[string]*KitConfig{}
		}
		if cfg.Kits["java"] == nil {
			cfg.Kits["java"] = &KitConfig{}
		}
		if cfg.Kits["java"].DefaultVersion == "" {
			cfg.Kits["java"].DefaultVersion = java
		}
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

	// Kits: last-wins (overlay replaces entirely)
	if overlay.Kits != nil {
		result.Kits = overlay.Kits
	}
	// Agents: last-wins
	if overlay.Agents != nil {
		result.Agents = overlay.Agents
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
		if cfg.Kits == nil {
			cfg.Kits = map[string]*KitConfig{}
		}
		if cfg.Kits["java"] == nil {
			cfg.Kits["java"] = &KitConfig{}
		}
		cfg.Kits["java"].DefaultVersion = flags.Java
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

// ExpandTilde replaces a leading ~/ with homeDir. A bare "~" returns homeDir.
func ExpandTilde(path, homeDir string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(homeDir, path[2:])
	}
	if path == "~" {
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
			host := ExpandTilde(parts[0], homeDir)
			return Volume{Host: host, Container: host, Options: parts[1]}, nil
		}
		if parts[1] == "" {
			return Volume{}, fmt.Errorf("empty container path in volume %q", raw)
		}
		return Volume{Host: ExpandTilde(parts[0], homeDir), Container: parts[1]}, nil
	default:
		if parts[0] == "" || parts[1] == "" {
			return Volume{}, fmt.Errorf("empty path in volume %q", raw)
		}
		for _, opt := range parts[2:] {
			if !mountOptions[opt] {
				return Volume{}, fmt.Errorf("invalid mount option %q in volume %q", opt, raw)
			}
		}
		return Volume{Host: ExpandTilde(parts[0], homeDir), Container: parts[1], Options: strings.Join(parts[2:], ":")}, nil
	}
}
