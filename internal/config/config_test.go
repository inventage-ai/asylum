package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMerge(t *testing.T) {
	tests := []struct {
		name   string
		base   Config
		over   Config
		check  func(t *testing.T, c Config)
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
			name: "lists concatenated",
			base: Config{Ports: []string{"3000"}},
			over: Config{Ports: []string{"8080"}},
			check: func(t *testing.T, c Config) {
				if len(c.Ports) != 2 || c.Ports[0] != "3000" || c.Ports[1] != "8080" {
					t.Errorf("ports = %v, want [3000, 8080]", c.Ports)
				}
			},
		},
		{
			name: "versions map scalar last wins",
			base: Config{Versions: map[string]string{"java": "17"}},
			over: Config{Versions: map[string]string{"java": "21"}},
			check: func(t *testing.T, c Config) {
				if c.Versions["java"] != "21" {
					t.Errorf("java = %q, want %q", c.Versions["java"], "21")
				}
			},
		},
		{
			name: "packages sub-lists concatenated",
			base: Config{Packages: map[string][]string{"apt": {"curl"}}},
			over: Config{Packages: map[string][]string{"apt": {"jq"}}},
			check: func(t *testing.T, c Config) {
				apt := c.Packages["apt"]
				if len(apt) != 2 || apt[0] != "curl" || apt[1] != "jq" {
					t.Errorf("apt = %v, want [curl, jq]", apt)
				}
			},
		},
		{
			name: "packages different keys merged",
			base: Config{Packages: map[string][]string{"apt": {"curl"}}},
			over: Config{Packages: map[string][]string{"npm": {"typescript"}}},
			check: func(t *testing.T, c Config) {
				if len(c.Packages) != 2 {
					t.Errorf("packages has %d keys, want 2", len(c.Packages))
				}
			},
		},
		{
			name: "release-channel last wins",
			base: Config{ReleaseChannel: "stable"},
			over: Config{ReleaseChannel: "dev"},
			check: func(t *testing.T, c Config) {
				if c.ReleaseChannel != "dev" {
					t.Errorf("release-channel = %q, want %q", c.ReleaseChannel, "dev")
				}
			},
		},
		{
			name: "release-channel empty overlay keeps base",
			base: Config{ReleaseChannel: "dev"},
			over: Config{},
			check: func(t *testing.T, c Config) {
				if c.ReleaseChannel != "dev" {
					t.Errorf("release-channel = %q, want %q", c.ReleaseChannel, "dev")
				}
			},
		},
		{
			name: "env map last wins",
			base: Config{Env: map[string]string{"KEY": "old"}},
			over: Config{Env: map[string]string{"KEY": "new"}},
			check: func(t *testing.T, c Config) {
				if c.Env["KEY"] != "new" {
					t.Errorf("env KEY = %q, want %q", c.Env["KEY"], "new")
				}
			},
		},
		{
			name: "env maps merged",
			base: Config{Env: map[string]string{"A": "1"}},
			over: Config{Env: map[string]string{"B": "2"}},
			check: func(t *testing.T, c Config) {
				if c.Env["A"] != "1" || c.Env["B"] != "2" {
					t.Errorf("env = %v, want A=1 B=2", c.Env)
				}
			},
		},
		{
			name: "env nil overlay keeps base",
			base: Config{Env: map[string]string{"A": "1"}},
			over: Config{},
			check: func(t *testing.T, c Config) {
				if c.Env["A"] != "1" {
					t.Errorf("env A = %q, want %q", c.Env["A"], "1")
				}
			},
		},
		{
			name: "onboarding map last wins",
			base: Config{Onboarding: map[string]bool{"npm": true}},
			over: Config{Onboarding: map[string]bool{"npm": false}},
			check: func(t *testing.T, c Config) {
				if c.Onboarding["npm"] != false {
					t.Errorf("onboarding npm = %v, want false", c.Onboarding["npm"])
				}
			},
		},
		{
			name: "onboarding nil overlay keeps base",
			base: Config{Onboarding: map[string]bool{"npm": false}},
			over: Config{},
			check: func(t *testing.T, c Config) {
				if c.Onboarding["npm"] != false {
					t.Errorf("onboarding npm = %v, want false", c.Onboarding["npm"])
				}
			},
		},
		{
			name: "nil base maps handled",
			base: Config{},
			over: Config{Versions: map[string]string{"java": "21"}, Packages: map[string][]string{"apt": {"curl"}}},
			check: func(t *testing.T, c Config) {
				if c.Versions["java"] != "21" {
					t.Errorf("java = %q, want %q", c.Versions["java"], "21")
				}
				if len(c.Packages["apt"]) != 1 {
					t.Errorf("apt = %v, want [curl]", c.Packages["apt"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := merge(tt.base, tt.over)
			tt.check(t, result)
		})
	}
}

func TestMergeDoesNotMutateBase(t *testing.T) {
	base := Config{
		Versions: map[string]string{"java": "17"},
		Packages: map[string][]string{"apt": {"curl"}},
	}
	overlay := Config{
		Versions: map[string]string{"java": "21", "node": "20"},
		Packages: map[string][]string{"apt": {"jq"}, "npm": {"typescript"}},
	}

	merge(base, overlay)

	if base.Versions["java"] != "17" {
		t.Errorf("merge mutated base.Versions[java]: got %q, want %q", base.Versions["java"], "17")
	}
	if _, ok := base.Versions["node"]; ok {
		t.Error("merge mutated base.Versions: added node key")
	}
	if len(base.Packages["apt"]) != 1 || base.Packages["apt"][0] != "curl" {
		t.Errorf("merge mutated base.Packages[apt]: got %v, want [curl]", base.Packages["apt"])
	}
	if _, ok := base.Packages["npm"]; ok {
		t.Error("merge mutated base.Packages: added npm key")
	}
}

func TestApplyFlags(t *testing.T) {
	cfg := Config{
		Agent: "claude",
		Ports: []string{"3000"},
	}
	flags := CLIFlags{
		Agent: "codex",
		Ports: []string{"9090"},
		Java:  "17",
	}
	result := applyFlags(cfg, flags)

	if result.Agent != "codex" {
		t.Errorf("agent = %q, want %q", result.Agent, "codex")
	}
	if len(result.Ports) != 2 {
		t.Errorf("ports = %v, want 2 entries", result.Ports)
	}
	if result.Versions["java"] != "17" {
		t.Errorf("java = %q, want %q", result.Versions["java"], "17")
	}
}

func TestApplyFlagsEnv(t *testing.T) {
	cfg := Config{Env: map[string]string{"A": "1"}}
	flags := CLIFlags{Env: map[string]string{"A": "2", "B": "3"}}
	result := applyFlags(cfg, flags)

	if result.Env["A"] != "2" {
		t.Errorf("env A = %q, want %q", result.Env["A"], "2")
	}
	if result.Env["B"] != "3" {
		t.Errorf("env B = %q, want %q", result.Env["B"], "3")
	}
}

func TestParseVolume(t *testing.T) {
	home := "/home/user"

	tests := []struct {
		name string
		raw  string
		want Volume
	}{
		{
			name: "standard syntax",
			raw:  "/host/path:/container/path",
			want: Volume{Host: "/host/path", Container: "/container/path"},
		},
		{
			name: "standard with options",
			raw:  "/host/path:/container/path:ro",
			want: Volume{Host: "/host/path", Container: "/container/path", Options: "ro"},
		},
		{
			name: "shorthand single path",
			raw:  "/data",
			want: Volume{Host: "/data", Container: "/data"},
		},
		{
			name: "shorthand with mount option",
			raw:  "/data:ro",
			want: Volume{Host: "/data", Container: "/data", Options: "ro"},
		},
		{
			name: "tilde expansion standard",
			raw:  "~/data:/data:ro",
			want: Volume{Host: "/home/user/data", Container: "/data", Options: "ro"},
		},
		{
			name: "tilde shorthand",
			raw:  "~/data",
			want: Volume{Host: "/home/user/data", Container: "/home/user/data"},
		},
		{
			name: "tilde shorthand with option",
			raw:  "~/data:rw",
			want: Volume{Host: "/home/user/data", Container: "/home/user/data", Options: "rw"},
		},
		{
			name: "tilde in container path preserved",
			raw:  "~/host:~/container",
			want: Volume{Host: "/home/user/host", Container: "~/container"},
		},
		{
			name: "tilde in container path with options",
			raw:  "~/host:~/container:ro",
			want: Volume{Host: "/home/user/host", Container: "~/container", Options: "ro"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseVolume(tt.raw, home)
			if got != tt.want {
				t.Errorf("ParseVolume(%q) = %+v, want %+v", tt.raw, got, tt.want)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	dir := t.TempDir()

	// Redirect HOME so os.UserHomeDir returns a temp dir with no config
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	// Create project config
	projectConfig := `agent: gemini
ports:
  - "8080"
`
	os.WriteFile(filepath.Join(dir, ".asylum"), []byte(projectConfig), 0644)

	// Create local override
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
	// Ports should include those from project config
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

func TestLoadEnv(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".asylum"), []byte("env:\n  MY_KEY: val1\n  OTHER: base\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".asylum.local"), []byte("env:\n  MY_KEY: val2\n"), 0644)

	cfg, err := Load(dir, CLIFlags{})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Env["MY_KEY"] != "val2" {
		t.Errorf("env MY_KEY = %q, want %q", cfg.Env["MY_KEY"], "val2")
	}
	if cfg.Env["OTHER"] != "base" {
		t.Errorf("env OTHER = %q, want %q", cfg.Env["OTHER"], "base")
	}
}

func TestToolVersionsJava(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	t.Run("provides java version", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, ".tool-versions"), []byte("java 21.0.2\nnodejs 20.11.0\n"), 0644)

		cfg, err := Load(dir, CLIFlags{})
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Versions["java"] != "21.0.2" {
			t.Errorf("java = %q, want %q", cfg.Versions["java"], "21.0.2")
		}
	})

	t.Run("tab-separated java version", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, ".tool-versions"), []byte("java\t21.0.2\nnodejs\t20.11.0\n"), 0644)

		cfg, err := Load(dir, CLIFlags{})
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Versions["java"] != "21.0.2" {
			t.Errorf("java = %q, want %q", cfg.Versions["java"], "21.0.2")
		}
	})

	t.Run("no java line", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, ".tool-versions"), []byte("nodejs 20.11.0\n"), 0644)

		cfg, err := Load(dir, CLIFlags{})
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Versions["java"] != "" {
			t.Errorf("java = %q, want empty", cfg.Versions["java"])
		}
	})

	t.Run("no file", func(t *testing.T) {
		dir := t.TempDir()

		cfg, err := Load(dir, CLIFlags{})
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Versions["java"] != "" {
			t.Errorf("java = %q, want empty", cfg.Versions["java"])
		}
	})

	t.Run("asylum config overrides", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, ".tool-versions"), []byte("java 21.0.2\n"), 0644)
		os.WriteFile(filepath.Join(dir, ".asylum"), []byte("versions:\n  java: \"17\"\n"), 0644)

		cfg, err := Load(dir, CLIFlags{})
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Versions["java"] != "17" {
			t.Errorf("java = %q, want %q", cfg.Versions["java"], "17")
		}
	})

	t.Run("CLI flag overrides", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, ".tool-versions"), []byte("java 21.0.2\n"), 0644)

		cfg, err := Load(dir, CLIFlags{Java: "25"})
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Versions["java"] != "25" {
			t.Errorf("java = %q, want %q", cfg.Versions["java"], "25")
		}
	})
}

func TestLoadSkipsDirectoryAsConfig(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	// Create ~/.asylum as a directory (as it would be in normal use)
	os.MkdirAll(filepath.Join(homeDir, ".asylum"), 0755)

	// Use home as projectDir — .asylum resolves to the ~/.asylum directory
	cfg, err := Load(homeDir, CLIFlags{})
	if err != nil {
		t.Fatalf("Load should skip directories, got error: %v", err)
	}
	if cfg.Agent != "" {
		t.Errorf("agent = %q, want empty", cfg.Agent)
	}
}
