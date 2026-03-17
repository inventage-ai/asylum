package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Agent    string              `yaml:"agent"`
	Ports    []string            `yaml:"ports"`
	Volumes  []string            `yaml:"volumes"`
	Versions map[string]string   `yaml:"versions"`
	Packages map[string][]string `yaml:"packages"`
}

type CLIFlags struct {
	Agent   string
	Ports   []string
	Volumes []string
	Java    string
	New     bool
	Rebuild bool
	Cleanup bool
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

	result.Ports = make([]string, 0, len(base.Ports)+len(overlay.Ports))
	result.Ports = append(result.Ports, base.Ports...)
	result.Ports = append(result.Ports, overlay.Ports...)

	result.Volumes = make([]string, 0, len(base.Volumes)+len(overlay.Volumes))
	result.Volumes = append(result.Volumes, base.Volumes...)
	result.Volumes = append(result.Volumes, overlay.Volumes...)

	if overlay.Versions != nil {
		if result.Versions == nil {
			result.Versions = make(map[string]string)
		}
		for k, v := range overlay.Versions {
			result.Versions[k] = v
		}
	}

	if overlay.Packages != nil {
		if result.Packages == nil {
			result.Packages = make(map[string][]string)
		}
		for k, v := range overlay.Packages {
			result.Packages[k] = append(result.Packages[k], v...)
		}
	}

	return result
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
