package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func ptrBool(b bool) *bool { return &b }

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
			name: "kits per-key merge preserves base",
			base: Config{Kits: map[string]*KitConfig{"java": nil, "openspec": nil}},
			over: Config{Kits: map[string]*KitConfig{"python": nil}},
			check: func(t *testing.T, c Config) {
				for _, name := range []string{"java", "openspec", "python"} {
					if !c.KitActive(name) {
						t.Errorf("%s should be active", name)
					}
				}
			},
		},
		{
			name: "kits overlay overrides single kit options",
			base: Config{Kits: map[string]*KitConfig{
				"java": {DefaultVersion: "17"},
				"node": nil,
			}},
			over: Config{Kits: map[string]*KitConfig{
				"java": {DefaultVersion: "21"},
			}},
			check: func(t *testing.T, c Config) {
				if c.KitOption("java").DefaultVersion != "21" {
					t.Errorf("java version = %q, want 21", c.KitOption("java").DefaultVersion)
				}
				if !c.KitActive("node") {
					t.Error("node should still be active")
				}
			},
		},
		{
			name: "kits nil overlay KitConfig preserves base",
			base: Config{Kits: map[string]*KitConfig{"node": {Packages: []string{"tsx"}}}},
			over: Config{Kits: map[string]*KitConfig{"node": nil}},
			check: func(t *testing.T, c Config) {
				kc := c.KitOption("node")
				if kc == nil || len(kc.Packages) != 1 || kc.Packages[0] != "tsx" {
					t.Errorf("node packages = %v, want [tsx]", kc)
				}
			},
		},
		{
			name: "agents per-key merge preserves base",
			base: Config{Agents: map[string]*AgentConfig{"claude": nil}},
			over: Config{Agents: map[string]*AgentConfig{"gemini": nil}},
			check: func(t *testing.T, c Config) {
				if !c.AgentActive("claude") {
					t.Error("claude should still be active")
				}
				if !c.AgentActive("gemini") {
					t.Error("gemini should be active")
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
	if result.KitOption("java").DefaultVersion != "17" {
		t.Errorf("java = %q, want %q", result.KitOption("java").DefaultVersion, "17")
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

	if cfg.KitOption("java").DefaultVersion != "21" {
		t.Errorf("JavaVersion() = %q, want 21", cfg.KitOption("java").DefaultVersion)
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

func TestPortCount(t *testing.T) {
	// Default
	cfg := Config{}
	if got := cfg.PortCount(); got != 5 {
		t.Errorf("default PortCount() = %d, want 5", got)
	}

	// Explicit
	cfg = Config{Kits: map[string]*KitConfig{"ports": {Count: 10}}}
	if got := cfg.PortCount(); got != 10 {
		t.Errorf("explicit PortCount() = %d, want 10", got)
	}

	// Zero means default
	cfg = Config{Kits: map[string]*KitConfig{"ports": {Count: 0}}}
	if got := cfg.PortCount(); got != 5 {
		t.Errorf("zero PortCount() = %d, want 5", got)
	}
}

func TestAgentIsolation(t *testing.T) {
	t.Run("returns value when set", func(t *testing.T) {
		cfg := Config{Agents: map[string]*AgentConfig{"claude": {Config: "shared"}}}
		if got := cfg.AgentIsolation("claude"); got != "shared" {
			t.Errorf("got %q, want shared", got)
		}
	})
	t.Run("returns empty when not set", func(t *testing.T) {
		cfg := Config{Agents: map[string]*AgentConfig{"claude": {}}}
		if got := cfg.AgentIsolation("claude"); got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})
	t.Run("returns empty when agent missing", func(t *testing.T) {
		cfg := Config{Agents: map[string]*AgentConfig{}}
		if got := cfg.AgentIsolation("claude"); got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})
	t.Run("returns empty when agents nil", func(t *testing.T) {
		cfg := Config{}
		if got := cfg.AgentIsolation("claude"); got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})
}

func TestSetAgentIsolation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	os.WriteFile(path, []byte("version: \"0.2\"\nagent: claude\nagents:\n  claude:\n"), 0644)

	if err := SetAgentIsolation(path, "claude", "shared"); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AgentIsolation("claude") != "shared" {
		t.Errorf("isolation = %q, want shared", cfg.AgentIsolation("claude"))
	}
}

func TestSetAgentIsolation_CreatesAgent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	os.WriteFile(path, []byte("version: \"0.2\"\nagent: claude\n"), 0644)

	if err := SetAgentIsolation(path, "claude", "project"); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AgentIsolation("claude") != "project" {
		t.Errorf("isolation = %q, want project", cfg.AgentIsolation("claude"))
	}
}

func TestSetKitCredentials(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	initial := "version: \"2\"\nkits:\n  java:\n    versions:\n      - 17\n      - 21\n"
	os.WriteFile(cfgPath, []byte(initial), 0644)

	if err := SetKitCredentials(cfgPath, "java", "auto"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(cfgPath)
	content := string(data)
	if !strings.Contains(content, "credentials: auto") {
		t.Errorf("expected 'credentials: auto' in config, got:\n%s", content)
	}
	if !strings.Contains(content, "versions:") {
		t.Errorf("existing content should be preserved, got:\n%s", content)
	}
}

func TestSetKitCredentials_NewKit(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	initial := "version: \"2\"\nkits:\n  node:\n"
	os.WriteFile(cfgPath, []byte(initial), 0644)

	if err := SetKitCredentials(cfgPath, "java", "auto"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(cfgPath)
	content := string(data)
	if !strings.Contains(content, "java:") {
		t.Errorf("expected java kit added, got:\n%s", content)
	}
	if !strings.Contains(content, "credentials: auto") {
		t.Errorf("expected credentials: auto, got:\n%s", content)
	}
}

func TestSetKitCredentials_NoKitsSection(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	initial := "version: \"2\"\n"
	os.WriteFile(cfgPath, []byte(initial), 0644)

	if err := SetKitCredentials(cfgPath, "java", "auto"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(cfgPath)
	content := string(data)
	if !strings.Contains(content, "kits:") {
		t.Errorf("expected kits section added, got:\n%s", content)
	}
	if !strings.Contains(content, "credentials: auto") {
		t.Errorf("expected credentials: auto, got:\n%s", content)
	}
}

func TestSetAgentIsolation_PreservesBlankLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	initial := "version: \"0.2\"\n\n# Comment\nagent: claude\n\nagents:\n  claude:\n  # gemini:\n"
	os.WriteFile(path, []byte(initial), 0644)

	if err := SetAgentIsolation(path, "claude", "isolated"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)
	if !strings.Contains(content, "\n\n# Comment") {
		t.Errorf("blank line before comment should be preserved, got:\n%s", content)
	}
	if !strings.Contains(content, "  # gemini:") {
		t.Errorf("comment should remain at original indentation, got:\n%s", content)
	}
}

func TestExpandTilde(t *testing.T) {
	home := "/Users/simon"
	tests := []struct {
		path, want string
	}{
		{"~/data", "/Users/simon/data"},
		{"~", "/Users/simon"},
		{"/absolute/path", "/absolute/path"},
		{"/home/claude/.m2", "/Users/simon/.m2"},
		{"/home/claude/.cache/pip", "/Users/simon/.cache/pip"},
		{"/home/claude", "/Users/simon"},
		{"relative", "relative"},
	}
	for _, tt := range tests {
		got := ExpandTilde(tt.path, home)
		if got != tt.want {
			t.Errorf("ExpandTilde(%q, %q) = %q, want %q", tt.path, home, got, tt.want)
		}
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
		{"tilde expansion in container", "~/data:~/data:ro", Volume{Host: "/home/user/data", Container: "/home/user/data", Options: "ro"}},
		{"tilde expansion both paths", "~/src:~/dest", Volume{Host: "/home/user/src", Container: "/home/user/dest"}},
		{"three parts with single option", "/host:/container:ro", Volume{Host: "/host", Container: "/container", Options: "ro"}},
		{"four parts with two options", "/host:/container:ro:z", Volume{Host: "/host", Container: "/container", Options: "ro:z"}},
		{"three parts with selinux label", "/host:/container:z", Volume{Host: "/host", Container: "/container", Options: "z"}},
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
	errTests := []struct {
		name string
		raw  string
	}{
		{"empty string", ""},
		{"invalid option in third part", "/host:/container:bogus"},
		{"invalid option in fourth part", "/host:/container:ro:bogus"},
		{"empty host in two-part", ":/container"},
		{"empty host with mount option", ":ro"},
		{"empty container in two-part", "/host:"},
		{"empty host in multi-part", ":/container:ro"},
		{"empty container in multi-part", "/host::ro"},
	}
	for _, tt := range errTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseVolume(tt.raw, home)
			if err == nil {
				t.Errorf("ParseVolume(%q) expected error, got nil", tt.raw)
			}
		})
	}
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

	cfg, err := Load(dir, CLIFlags{}, "")
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
		cfg, err := Load(dir, CLIFlags{}, "")
		if err != nil {
			t.Fatal(err)
		}
		if cfg.KitOption("java").DefaultVersion != "21.0.2" {
			t.Errorf("java = %q, want %q", cfg.KitOption("java").DefaultVersion, "21.0.2")
		}
	})

	t.Run("CLI flag overrides", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, ".tool-versions"), []byte("java 21.0.2\n"), 0644)
		cfg, err := Load(dir, CLIFlags{Java: "25"}, "")
		if err != nil {
			t.Fatal(err)
		}
		if cfg.KitOption("java").DefaultVersion != "25" {
			t.Errorf("java = %q, want %q", cfg.KitOption("java").DefaultVersion, "25")
		}
	})

	t.Run("project config overrides tool-versions", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, ".tool-versions"), []byte("java 21.0.2\n"), 0644)
		os.WriteFile(filepath.Join(dir, ".asylum"), []byte("kits:\n  java:\n    default-version: \"17\"\n"), 0644)
		cfg, err := Load(dir, CLIFlags{}, "")
		if err != nil {
			t.Fatal(err)
		}
		if cfg.KitOption("java").DefaultVersion != "17" {
			t.Errorf("java = %q, want %q (project config should override .tool-versions)", cfg.KitOption("java").DefaultVersion, "17")
		}
	})

	t.Run("tool-versions overrides global config", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, ".tool-versions"), []byte("java 21.0.2\n"), 0644)
		// Set global config with java 11
		os.MkdirAll(filepath.Join(homeDir, ".asylum"), 0755)
		os.WriteFile(filepath.Join(homeDir, ".asylum", "config.yaml"), []byte("version: \"0.2\"\nkits:\n  java:\n    default-version: \"11\"\n"), 0644)
		defer os.Remove(filepath.Join(homeDir, ".asylum", "config.yaml"))
		cfg, err := Load(dir, CLIFlags{}, "")
		if err != nil {
			t.Fatal(err)
		}
		if cfg.KitOption("java").DefaultVersion != "21.0.2" {
			t.Errorf("java = %q, want %q (.tool-versions should override global config)", cfg.KitOption("java").DefaultVersion, "21.0.2")
		}
	})
}

func TestMergeKitConfig(t *testing.T) {
	boolPtr := func(v bool) *bool { return &v }

	tests := []struct {
		name  string
		base  *KitConfig
		over  *KitConfig
		check func(t *testing.T, kc *KitConfig)
	}{
		{
			name: "nil base returns overlay",
			base: nil,
			over: &KitConfig{DefaultVersion: "21"},
			check: func(t *testing.T, kc *KitConfig) {
				if kc.DefaultVersion != "21" {
					t.Errorf("version = %q, want 21", kc.DefaultVersion)
				}
			},
		},
		{
			name: "nil overlay returns base",
			base: &KitConfig{DefaultVersion: "17"},
			over: nil,
			check: func(t *testing.T, kc *KitConfig) {
				if kc.DefaultVersion != "17" {
					t.Errorf("version = %q, want 17", kc.DefaultVersion)
				}
			},
		},
		{
			name: "both nil returns nil",
			base: nil,
			over: nil,
			check: func(t *testing.T, kc *KitConfig) {
				if kc != nil {
					t.Error("expected nil")
				}
			},
		},
		{
			name: "scalar override",
			base: &KitConfig{DefaultVersion: "17", TabTitle: "base"},
			over: &KitConfig{DefaultVersion: "21"},
			check: func(t *testing.T, kc *KitConfig) {
				if kc.DefaultVersion != "21" {
					t.Errorf("version = %q, want 21", kc.DefaultVersion)
				}
				if kc.TabTitle != "base" {
					t.Errorf("tab title = %q, want base", kc.TabTitle)
				}
			},
		},
		{
			name: "disabled flag override",
			base: &KitConfig{},
			over: &KitConfig{Disabled: boolPtr(true)},
			check: func(t *testing.T, kc *KitConfig) {
				if kc.Disabled == nil || !*kc.Disabled {
					t.Error("disabled should be true")
				}
			},
		},
		{
			name: "disabled false overrides disabled true",
			base: &KitConfig{Disabled: boolPtr(true)},
			over: &KitConfig{Disabled: boolPtr(false)},
			check: func(t *testing.T, kc *KitConfig) {
				if kc.Disabled == nil || *kc.Disabled {
					t.Error("disabled should be false (project override)")
				}
			},
		},
		{
			name: "packages concat",
			base: &KitConfig{Packages: []string{"tsx"}},
			over: &KitConfig{Packages: []string{"vitest"}},
			check: func(t *testing.T, kc *KitConfig) {
				if len(kc.Packages) != 2 || kc.Packages[0] != "tsx" || kc.Packages[1] != "vitest" {
					t.Errorf("packages = %v, want [tsx vitest]", kc.Packages)
				}
			},
		},
		{
			name: "build concat",
			base: &KitConfig{Build: []string{"apt-get install foo"}},
			over: &KitConfig{Build: []string{"curl bar"}},
			check: func(t *testing.T, kc *KitConfig) {
				if len(kc.Build) != 2 || kc.Build[0] != "apt-get install foo" || kc.Build[1] != "curl bar" {
					t.Errorf("build = %v, want [apt-get install foo, curl bar]", kc.Build)
				}
			},
		},
		{
			name: "versions replace",
			base: &KitConfig{Versions: []string{"17", "21"}},
			over: &KitConfig{Versions: []string{"25"}},
			check: func(t *testing.T, kc *KitConfig) {
				if len(kc.Versions) != 1 || kc.Versions[0] != "25" {
					t.Errorf("versions = %v, want [25]", kc.Versions)
				}
			},
		},
		{
			name: "nil overlay versions preserves base",
			base: &KitConfig{Versions: []string{"17", "21"}},
			over: &KitConfig{},
			check: func(t *testing.T, kc *KitConfig) {
				if len(kc.Versions) != 2 {
					t.Errorf("versions = %v, want [17 21]", kc.Versions)
				}
			},
		},
		{
			name: "count nonzero replaces",
			base: &KitConfig{Count: 5},
			over: &KitConfig{Count: 10},
			check: func(t *testing.T, kc *KitConfig) {
				if kc.Count != 10 {
					t.Errorf("count = %d, want 10", kc.Count)
				}
			},
		},
		{
			name: "count zero preserves base",
			base: &KitConfig{Count: 5},
			over: &KitConfig{},
			check: func(t *testing.T, kc *KitConfig) {
				if kc.Count != 5 {
					t.Errorf("count = %d, want 5", kc.Count)
				}
			},
		},
		{
			name: "credentials overlay replaces base",
			base: &KitConfig{Credentials: &Credentials{Auto: true}},
			over: &KitConfig{Credentials: &Credentials{Explicit: []string{"nexus"}}},
			check: func(t *testing.T, kc *KitConfig) {
				if kc.Credentials == nil || kc.Credentials.Auto || len(kc.Credentials.Explicit) != 1 || kc.Credentials.Explicit[0] != "nexus" {
					t.Errorf("credentials = %+v, want explicit [nexus]", kc.Credentials)
				}
			},
		},
		{
			name: "absent credentials in overlay preserves base",
			base: &KitConfig{Credentials: &Credentials{Auto: true}},
			over: &KitConfig{},
			check: func(t *testing.T, kc *KitConfig) {
				if kc.Credentials == nil || !kc.Credentials.Auto {
					t.Errorf("credentials = %+v, want auto", kc.Credentials)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeKitConfig(tt.base, tt.over)
			tt.check(t, result)
		})
	}
}

func TestLoadSkipsDirectoryAsConfig(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	os.MkdirAll(filepath.Join(homeDir, ".asylum"), 0755)

	cfg, err := Load(homeDir, CLIFlags{}, "")
	if err != nil {
		t.Fatalf("Load should skip directories, got error: %v", err)
	}
	if cfg.Agent != "" {
		t.Errorf("agent = %q, want empty", cfg.Agent)
	}
}

func TestKitPackagesFromProjectConfig(t *testing.T) {
	// Regression test: kit packages from project config must survive merge
	// and be accessible via KitPackages (issue #16).
	global := Config{
		Kits: map[string]*KitConfig{
			"node": {ShadowNodeModules: ptrBool(true)},
		},
	}
	project := Config{
		Kits: map[string]*KitConfig{
			"node": {Packages: []string{"typescript-language-server"}},
		},
	}
	merged := Merge(global, project)

	pkgs := merged.KitPackages("node")
	if len(pkgs) != 1 || pkgs[0] != "typescript-language-server" {
		t.Errorf("KitPackages(node) = %v, want [typescript-language-server]", pkgs)
	}

	// Also verify shadow-node-modules survived the merge
	kc := merged.KitOption("node")
	if kc == nil || kc.ShadowNodeModules == nil || !*kc.ShadowNodeModules {
		t.Error("shadow-node-modules should be preserved from global config")
	}
}

func TestKitPackagesFromProjectConfig_NilBase(t *testing.T) {
	// When global config has node: ~ (nil KitConfig), project packages must still work
	global := Config{
		Kits: map[string]*KitConfig{
			"node": nil,
		},
	}
	project := Config{
		Kits: map[string]*KitConfig{
			"node": {Packages: []string{"turbo"}},
		},
	}
	merged := Merge(global, project)

	pkgs := merged.KitPackages("node")
	if len(pkgs) != 1 || pkgs[0] != "turbo" {
		t.Errorf("KitPackages(node) = %v, want [turbo]", pkgs)
	}
}

func TestConfigHash(t *testing.T) {
	t.Run("deterministic", func(t *testing.T) {
		cfg := Config{
			Volumes: []string{"/b", "/a"},
			Env:     map[string]string{"Z": "1", "A": "2"},
			Ports:   []string{"8080", "3000"},
		}
		h1 := ConfigHash(cfg)
		h2 := ConfigHash(cfg)
		if h1 != h2 {
			t.Errorf("same config produced different hashes: %s vs %s", h1, h2)
		}
	})

	t.Run("different configs differ", func(t *testing.T) {
		a := Config{Volumes: []string{"/a"}}
		b := Config{Volumes: []string{"/b"}}
		if ConfigHash(a) == ConfigHash(b) {
			t.Error("different volumes should produce different hashes")
		}

		c := Config{Env: map[string]string{"K": "v1"}}
		d := Config{Env: map[string]string{"K": "v2"}}
		if ConfigHash(c) == ConfigHash(d) {
			t.Error("different env should produce different hashes")
		}

		e := Config{Ports: []string{"3000"}}
		f := Config{Ports: []string{"8080"}}
		if ConfigHash(e) == ConfigHash(f) {
			t.Error("different ports should produce different hashes")
		}
	})

	t.Run("empty config stable", func(t *testing.T) {
		h := ConfigHash(Config{})
		if h == "" {
			t.Error("empty config should produce a non-empty hash")
		}
		if h != ConfigHash(Config{}) {
			t.Error("empty config hash should be stable")
		}
	})

	t.Run("order independent", func(t *testing.T) {
		a := Config{
			Volumes: []string{"/z", "/a", "/m"},
			Ports:   []string{"9000", "3000", "5000"},
		}
		b := Config{
			Volumes: []string{"/a", "/m", "/z"},
			Ports:   []string{"3000", "5000", "9000"},
		}
		if ConfigHash(a) != ConfigHash(b) {
			t.Error("same values in different order should produce same hash")
		}
	})

	t.Run("credential change detected", func(t *testing.T) {
		auto := Config{Kits: map[string]*KitConfig{
			"java": {Credentials: &Credentials{Auto: true}},
		}}
		explicit := Config{Kits: map[string]*KitConfig{
			"java": {Credentials: &Credentials{Explicit: []string{"nexus"}}},
		}}
		if ConfigHash(auto) == ConfigHash(explicit) {
			t.Error("different credential configs should produce different hashes")
		}
	})

	t.Run("nil kit config differs from absent kit", func(t *testing.T) {
		withKit := Config{Kits: map[string]*KitConfig{
			"java": nil,
		}}
		without := Config{}
		if ConfigHash(withKit) == ConfigHash(without) {
			t.Error("declared kit (even with nil config) should differ from absent kit")
		}
	})

	t.Run("non-runtime fields excluded", func(t *testing.T) {
		base := Config{}
		varied := Config{Agent: "claude", Version: "2", ReleaseChannel: "dev", Agents: map[string]*AgentConfig{"claude": {Config: "shared"}}}
		if ConfigHash(base) != ConfigHash(varied) {
			t.Error("non-runtime fields (Agent, Version, ReleaseChannel, Agents) should not affect hash")
		}
	})

	t.Run("new kit config fields included automatically", func(t *testing.T) {
		without := Config{Kits: map[string]*KitConfig{"java": {}}}
		with := Config{Kits: map[string]*KitConfig{"java": {Isolation: "project"}}}
		if ConfigHash(without) == ConfigHash(with) {
			t.Error("any non-zero config field should affect hash without explicit code changes")
		}
	})
}
