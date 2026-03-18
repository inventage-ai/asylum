package config

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Agent          string              `yaml:"agent"`
	ReleaseChannel string              `yaml:"release-channel"`
	Ports          []string            `yaml:"ports"`
	Volumes        []string            `yaml:"volumes"`
	Versions       map[string]string   `yaml:"versions"`
	Packages       map[string][]string `yaml:"packages"`
}

type CLIFlags struct {
	Agent   string
	Ports   []string
	Volumes []string
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

	cfg = applyFlags(cfg, flags)
	return cfg, nil
}

func loadFile(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, nil
		}
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

	result.Ports = concatSlices(base.Ports, overlay.Ports)
	result.Volumes = concatSlices(base.Volumes, overlay.Volumes)

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

func concatSlices(a, b []string) []string {
	out := make([]string, 0, len(a)+len(b))
	out = append(out, a...)
	out = append(out, b...)
	return out
}

func applyFlags(cfg Config, flags CLIFlags) Config {
	if flags.Agent != "" {
		cfg.Agent = flags.Agent
	}
	if flags.Java != "" {
		if cfg.Versions == nil {
			cfg.Versions = make(map[string]string)
		}
		cfg.Versions["java"] = flags.Java
	}
	cfg.Ports = append(cfg.Ports, flags.Ports...)
	cfg.Volumes = append(cfg.Volumes, flags.Volumes...)
	return cfg
}

var mountOptions = map[string]bool{
	"ro": true, "rw": true, "z": true, "Z": true,
	"shared": true, "slave": true, "private": true,
	"rshared": true, "rslave": true, "rprivate": true,
	"nocopy": true, "consistent": true, "cached": true, "delegated": true,
}

func ParseVolume(raw string, homeDir string) Volume {
	expandTilde := func(p string) string {
		if strings.HasPrefix(p, "~/") {
			return filepath.Join(homeDir, p[2:])
		}
		if p == "~" {
			return homeDir
		}
		return p
	}

	parts := strings.Split(raw, ":")

	switch len(parts) {
	case 1:
		// "/data" → same path both sides
		host := expandTilde(parts[0])
		return Volume{Host: host, Container: host}
	case 2:
		if mountOptions[parts[1]] {
			// "/data:ro" → shorthand with option
			host := expandTilde(parts[0])
			return Volume{Host: host, Container: host, Options: parts[1]}
		}
		// "/host:/container"
		return Volume{Host: expandTilde(parts[0]), Container: parts[1]}
	default:
		// "/host:/container:opts" or more
		return Volume{Host: expandTilde(parts[0]), Container: parts[1], Options: strings.Join(parts[2:], ":")}
	}
}
