package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMerge(t *testing.T) {
	tests := []struct {
		name  string
		base  Config
		over  Config
		check func(t *testing.T, c Config)
	}{
		{
			name: "scalar last wins",
			base: Config{Agent: "claude"},
			over: Config{Agent: "gemini"},
			check: func(t *testing.T, c Config) {
				if c.Agent != "gemini" {
					t.Errorf("agent = %q, want %q", c.Agent, "gemini")
				}
			},
		},
		{
			name: "scalar empty overlay keeps base",
			base: Config{Agent: "claude"},
			over: Config{},
			check: func(t *testing.T, c Config) {
				if c.Agent != "claude" {
					t.Errorf("agent = %q, want %q", c.Agent, "claude")
				}
			},
		},
		{
			name: "ports concatenated",
			base: Config{Ports: []string{"3000"}},
			over: Config{Ports: []string{"8080"}},
			check: func(t *testing.T, c Config) {
				if len(c.Ports) != 2 {
					t.Errorf("ports = %v, want 2 entries", c.Ports)
				}
			},
		},
		{
			name: "env merged per key",
			base: Config{Env: map[string]string{"A": "1"}},
			over: Config{Env: map[string]string{"A": "2", "B": "3"}},
			check: func(t *testing.T, c Config) {
				if c.Env["A"] != "2" {
					t.Errorf("env A = %q, want 2", c.Env["A"])
				}
				if c.Env["B"] != "3" {
					t.Errorf("env B = %q, want 3", c.Env["B"])
				}
			},
		},
		{
			name: "kits last-wins",
			base: Config{Kits: map[string]*KitConfig{"java": nil}},
			over: Config{Kits: map[string]*KitConfig{"python": nil}},
			check: func(t *testing.T, c Config) {
				if _, ok := c.Kits["java"]; ok {
					t.Error("java should not be present (last-wins)")
				}
				if _, ok := c.Kits["python"]; !ok {
					t.Error("python should be present")
				}
			},
		},
		{
			name: "agents last-wins",
			base: Config{Agents: map[string]*AgentConfig{"claude": nil}},
			over: Config{Agents: map[string]*AgentConfig{"gemini": nil}},
			check: func(t *testing.T, c Config) {
				if _, ok := c.Agents["claude"]; ok {
					t.Error("claude should not be present (last-wins)")
				}
				if _, ok := c.Agents["gemini"]; !ok {
					t.Error("gemini should be present")
				}
			},
		},
		{
			name: "nil kits overlay keeps base",
			base: Config{Kits: map[string]*KitConfig{"java": nil}},
			over: Config{},
			check: func(t *testing.T, c Config) {
				if !c.KitActive("java") {
					t.Error("java should still be active")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Merge(tt.base, tt.over)
			tt.check(t, result)
		})
	}
}

func TestApplyFlags(t *testing.T) {
	cfg := Config{Agent: "claude", Ports: []string{"3000"}}
	flags := CLIFlags{Agent: "codex", Ports: []string{"9090"}, Java: "17"}
	result := applyFlags(cfg, flags)

	if result.Agent != "codex" {
		t.Errorf("agent = %q, want %q", result.Agent, "codex")
	}
	if len(result.Ports) != 2 {
		t.Errorf("ports = %v, want 2 entries", result.Ports)
	}
	if result.JavaVersion() != "17" {
		t.Errorf("java = %q, want %q", result.JavaVersion(), "17")
	}
}

func TestApplyFlagsKits(t *testing.T) {
	k := []string{"java", "python"}
	cfg := Config{}
	flags := CLIFlags{Kits: &k}
	result := applyFlags(cfg, flags)
	if !result.KitActive("java") || !result.KitActive("python") {
		t.Errorf("kits should include java and python, got %v", result.KitNames())
	}
}

func TestApplyFlagsAgents(t *testing.T) {
	a := []string{"claude", "gemini"}
	cfg := Config{}
	flags := CLIFlags{Agents: &a}
	result := applyFlags(cfg, flags)
	if !result.AgentActive("claude") || !result.AgentActive("gemini") {
		t.Errorf("agents should include claude and gemini, got %v", result.AgentNames())
	}
}

func TestKitHelpers(t *testing.T) {
	b := true
	cfg := Config{
		Kits: map[string]*KitConfig{
			"java":  {DefaultVersion: "21"},
			"node":  {ShadowNodeModules: &b},
			"title": {TabTitle: "test"},
		},
	}

	if cfg.JavaVersion() != "21" {
		t.Errorf("JavaVersion() = %q, want 21", cfg.JavaVersion())
	}
	if cfg.TabTitle() != "test" {
		t.Errorf("TabTitle() = %q, want test", cfg.TabTitle())
	}
	if cfg.ShadowNodeModulesOff() {
		t.Error("ShadowNodeModulesOff() should be false when true")
	}
	if !cfg.KitActive("java") {
		t.Error("java should be active")
	}
	if cfg.KitActive("python") {
		t.Error("python should not be active")
	}
}

func TestParseVolume(t *testing.T) {
	home := "/home/user"
	tests := []struct {
		name string
		raw  string
		want Volume
	}{
		{"standard syntax", "/host/path:/container/path", Volume{Host: "/host/path", Container: "/container/path"}},
		{"standard with options", "/host/path:/container/path:ro", Volume{Host: "/host/path", Container: "/container/path", Options: "ro"}},
		{"shorthand single path", "/data", Volume{Host: "/data", Container: "/data"}},
		{"shorthand with mount option", "/data:ro", Volume{Host: "/data", Container: "/data", Options: "ro"}},
		{"tilde expansion standard", "~/data:/data:ro", Volume{Host: "/home/user/data", Container: "/data", Options: "ro"}},
		{"tilde shorthand", "~/data", Volume{Host: "/home/user/data", Container: "/home/user/data"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVolume(tt.raw, home)
			if err != nil {
				t.Fatalf("ParseVolume(%q) unexpected error: %v", tt.raw, err)
			}
			if got != tt.want {
				t.Errorf("ParseVolume(%q) = %+v, want %+v", tt.raw, got, tt.want)
			}
		})
	}

	// Error cases
	t.Run("empty string", func(t *testing.T) {
		_, err := ParseVolume("", home)
		if err == nil {
			t.Error("ParseVolume(\"\") expected error, got nil")
		}
	})
}

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	projectConfig := `agent: gemini
ports:
  - "8080"
`
	os.WriteFile(filepath.Join(dir, ".asylum"), []byte(projectConfig), 0644)

	localConfig := `agent: codex
`
	os.WriteFile(filepath.Join(dir, ".asylum.local"), []byte(localConfig), 0644)

	cfg, err := Load(dir, CLIFlags{})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Agent != "codex" {
		t.Errorf("agent = %q, want %q", cfg.Agent, "codex")
	}
	found := false
	for _, p := range cfg.Ports {
		if p == "8080" {
			found = true
		}
	}
	if !found {
		t.Errorf("ports %v missing 8080", cfg.Ports)
	}
}

func TestToolVersionsJava(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	t.Run("provides java version", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, ".tool-versions"), []byte("java 21.0.2\n"), 0644)
		cfg, err := Load(dir, CLIFlags{})
		if err != nil {
			t.Fatal(err)
		}
		if cfg.JavaVersion() != "21.0.2" {
			t.Errorf("java = %q, want %q", cfg.JavaVersion(), "21.0.2")
		}
	})

	t.Run("CLI flag overrides", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, ".tool-versions"), []byte("java 21.0.2\n"), 0644)
		cfg, err := Load(dir, CLIFlags{Java: "25"})
		if err != nil {
			t.Fatal(err)
		}
		if cfg.JavaVersion() != "25" {
			t.Errorf("java = %q, want %q", cfg.JavaVersion(), "25")
		}
	})
}

func TestLoadSkipsDirectoryAsConfig(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	os.MkdirAll(filepath.Join(homeDir, ".asylum"), 0755)

	cfg, err := Load(homeDir, CLIFlags{})
	if err != nil {
		t.Fatalf("Load should skip directories, got error: %v", err)
	}
	if cfg.Agent != "" {
		t.Errorf("agent = %q, want empty", cfg.Agent)
	}
}
