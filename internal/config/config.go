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

type Config struct {
	Agent          string              `yaml:"agent"`
	ReleaseChannel string              `yaml:"release-channel"`
	Ports          []string            `yaml:"ports"`
	Volumes        []string            `yaml:"volumes"`
	Env            map[string]string   `yaml:"env"`
	Versions       map[string]string   `yaml:"versions"`
	Packages       map[string][]string `yaml:"packages"`
}

type CLIFlags struct {
	Agent   string
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
		layer, err := loadFile(path)
		if err != nil {
			return Config{}, fmt.Errorf("%s: %w", path, err)
		}
		cfg = merge(cfg, layer)
	}

	if java := readToolVersionsJava(projectDir); java != "" && cfg.Versions["java"] == "" {
		cfg.Versions = setVersion(cfg.Versions, "java", java)
	}

	cfg = applyFlags(cfg, flags)
	return cfg, nil
}

func loadFile(path string) (Config, error) {
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

func merge(base, overlay Config) Config {
	result := base

	if overlay.Agent != "" {
		result.Agent = overlay.Agent
	}
	if overlay.ReleaseChannel != "" {
		result.ReleaseChannel = overlay.ReleaseChannel
	}

	result.Ports = slices.Concat(base.Ports, overlay.Ports)
	result.Volumes = slices.Concat(base.Volumes, overlay.Volumes)

	if overlay.Env != nil {
		merged := make(map[string]string, len(base.Env)+len(overlay.Env))
		maps.Copy(merged, base.Env)
		maps.Copy(merged, overlay.Env)
		result.Env = merged
	}

	if overlay.Versions != nil {
		merged := make(map[string]string, len(base.Versions)+len(overlay.Versions))
		maps.Copy(merged, base.Versions)
		maps.Copy(merged, overlay.Versions)
		result.Versions = merged
	}

	if overlay.Packages != nil {
		merged := make(map[string][]string, len(base.Packages)+len(overlay.Packages))
		for k, v := range base.Packages {
			merged[k] = append([]string(nil), v...)
		}
		for k, v := range overlay.Packages {
			merged[k] = append(merged[k], v...)
		}
		result.Packages = merged
	}

	return result
}

func applyFlags(cfg Config, flags CLIFlags) Config {
	if flags.Agent != "" {
		cfg.Agent = flags.Agent
	}
	if flags.Java != "" {
		cfg.Versions = setVersion(cfg.Versions, "java", flags.Java)
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

func setVersion(m map[string]string, key, value string) map[string]string {
	if m == nil {
		m = make(map[string]string)
	}
	m[key] = value
	return m
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

func ParseVolume(raw string, homeDir string) Volume {
	parts := strings.Split(raw, ":")

	switch len(parts) {
	case 1:
		// "/data" → same path both sides
		host := ExpandTilde(parts[0], homeDir)
		return Volume{Host: host, Container: host}
	case 2:
		if mountOptions[parts[1]] {
			// "/data:ro" → shorthand with option
			host := ExpandTilde(parts[0], homeDir)
			return Volume{Host: host, Container: host, Options: parts[1]}
		}
		// "/host:/container"
		return Volume{Host: ExpandTilde(parts[0], homeDir), Container: parts[1]}
	default:
		// "/host:/container:opts" or more
		return Volume{Host: ExpandTilde(parts[0], homeDir), Container: parts[1], Options: strings.Join(parts[2:], ":")}
	}
}
