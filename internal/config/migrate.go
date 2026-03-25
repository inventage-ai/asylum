package config

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const currentVersion = "0.2"

// NeedsMigration checks if a config file needs migration to v2 format.
// Global config: detected by missing or old version field.
// Project config: detected by presence of "features" key.
func NeedsMigration(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	var raw map[string]any
	if yaml.Unmarshal(data, &raw) != nil {
		return false
	}

	// Global config: has version field that's old or missing
	if v, ok := raw["version"]; ok {
		if s, ok := v.(string); ok && s >= currentVersion {
			return false
		}
		return true
	}

	// Project config: has "features" key (v1 indicator)
	if _, ok := raw["features"]; ok {
		return true
	}

	// Check for other v1-only keys
	for _, key := range []string{"profiles", "packages", "versions", "onboarding", "tab-title"} {
		if _, ok := raw[key]; ok {
			return true
		}
	}

	return false
}

// MigrateV1ToV2 reads a config file, transforms v1 fields to v2 structure,
// and writes back. Creates a .backup before modifying.
func MigrateV1ToV2(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Create backup, but don't overwrite an existing one (preserves original v1 content)
	backupPath := path + ".backup"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		if err := os.WriteFile(backupPath, data, 0644); err != nil {
			return err
		}
	}

	kits := map[string]any{}
	if existing, ok := raw["kits"].(map[string]any); ok {
		kits = existing
	}

	ensureKit := func(name string) map[string]any {
		if k, ok := kits[name].(map[string]any); ok {
			return k
		}
		m := map[string]any{}
		kits[name] = m
		return m
	}

	// Docker was always enabled in v1, so ensure docker kit is present
	if _, exists := kits["docker"]; !exists {
		kits["docker"] = map[string]any{}
	}

	// profiles: [java, node] → kits: {java: {}, node: {}}
	if profiles, ok := raw["profiles"]; ok {
		if list, ok := profiles.([]any); ok {
			for _, p := range list {
				if name, ok := p.(string); ok {
					if _, exists := kits[name]; !exists {
						kits[name] = map[string]any{}
					}
				}
			}
		}
		delete(raw, "profiles")
	}

	// versions: {java: "21"} → kits.java.default-version: "21"
	if versions, ok := raw["versions"].(map[string]any); ok {
		if java, ok := versions["java"]; ok {
			kit := ensureKit("java")
			kit["default-version"] = java
		}
		delete(raw, "versions")
	}

	// packages: {apt: [...], npm: [...], pip: [...], run: [...]}
	if packages, ok := raw["packages"].(map[string]any); ok {
		if apt, ok := packages["apt"]; ok {
			kit := ensureKit("apt")
			kit["packages"] = apt
		}
		if npm, ok := packages["npm"]; ok {
			kit := ensureKit("node")
			kit["packages"] = npm
		}
		if pip, ok := packages["pip"]; ok {
			kit := ensureKit("python")
			kit["packages"] = pip
		}
		if run, ok := packages["run"]; ok {
			kit := ensureKit("shell")
			kit["build"] = run
		}
		delete(raw, "packages")
	}

	// features: {shadow-node-modules: true, onboarding: false, allow-agent-terminal-title: false}
	if features, ok := raw["features"].(map[string]any); ok {
		if v, ok := features["shadow-node-modules"]; ok {
			kit := ensureKit("node")
			kit["shadow-node-modules"] = v
		}
		if v, ok := features["allow-agent-terminal-title"]; ok {
			kit := ensureKit("title")
			kit["allow-agent-terminal-title"] = v
		}
		// features.onboarding is dropped (global disable no longer exists)
		delete(raw, "features")
	}

	// onboarding: {npm: false} → kits.node.onboarding: false
	if onboarding, ok := raw["onboarding"].(map[string]any); ok {
		if npm, ok := onboarding["npm"]; ok {
			kit := ensureKit("node")
			kit["onboarding"] = npm
		}
		delete(raw, "onboarding")
	}

	// tab-title: "..." → kits.title.tab-title: "..."
	if title, ok := raw["tab-title"]; ok {
		kit := ensureKit("title")
		kit["tab-title"] = title
		delete(raw, "tab-title")
	}

	// agents: [claude, gemini] → agents: {claude: {}, gemini: {}}
	if agents, ok := raw["agents"]; ok {
		if list, ok := agents.([]any); ok {
			agentMap := map[string]any{}
			for _, a := range list {
				if name, ok := a.(string); ok {
					agentMap[name] = map[string]any{}
				}
			}
			raw["agents"] = agentMap
		}
	}

	if len(kits) > 0 {
		raw["kits"] = kits
	}

	// Set version for global configs (has version key or is ~/.asylum/config.yaml)
	if _, hadVersion := raw["version"]; hadVersion || strings.HasSuffix(path, "config.yaml") {
		raw["version"] = currentVersion
	}

	out, err := yaml.Marshal(raw)
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0644)
}
