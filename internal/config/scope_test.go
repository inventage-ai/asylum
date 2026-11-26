package config

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/inventage-ai/asylum/internal/kit"
	"gopkg.in/yaml.v3"
)

// Every option needs a deliberate decision about where it may be set.
func TestScopeTags(t *testing.T) {
	valid := []kit.Scope{kit.ScopeAny, kit.ScopePersonal, kit.ScopeGlobal}
	for _, typ := range []reflect.Type{reflect.TypeFor[Config](), reflect.TypeFor[KitConfig](), reflect.TypeFor[AgentConfig]()} {
		for f := range typ.Fields() {
			name, _, _ := strings.Cut(f.Tag.Get("yaml"), ",")
			if name == "" || name == "-" {
				continue
			}
			if s := kit.Scope(f.Tag.Get("scope")); !slices.Contains(valid, s) {
				t.Errorf("%s.%s: scope tag %q, want one of %v", typ.Name(), f.Name, s, valid)
			}
		}
	}
}

func filterYAML(t *testing.T, l layer, src string) (Config, []Rejection) {
	t.Helper()
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(src), &doc); err != nil {
		t.Fatal(err)
	}
	rejected, err := filterScopes(&doc, l, ".asylum")
	if err != nil {
		t.Fatalf("filterScopes: %v", err)
	}
	var cfg Config
	if err := doc.Decode(&cfg); err != nil {
		t.Fatal(err)
	}
	return cfg, rejected
}

func rejectedPaths(rs []Rejection) []string {
	var paths []string
	for _, r := range rs {
		paths = append(paths, r.Path)
	}
	return paths
}

func TestFilterScopes(t *testing.T) {
	tests := []struct {
		name  string
		layer layer
		src   string
		want  []string // rejected paths
		check func(*testing.T, Config)
	}{
		{
			name:  "any-scope option in project",
			layer: layerProject,
			src:   "kits:\n  apt:\n    packages: [golang]\nenv:\n  DEBUG: \"1\"\nports: [\"8080\"]\n",
			check: func(t *testing.T, c Config) {
				if got := c.KitPackages("apt"); !slices.Equal(got, []string{"golang"}) {
					t.Errorf("apt packages = %v", got)
				}
			},
		},
		{
			name:  "personal options in project",
			layer: layerProject,
			src:   "volumes: [\"~/data:/data\"]\nenv-from-host: {A: B}\ndefault-resume: true\nenv: {DEBUG: \"1\"}\n",
			want:  []string{"volumes", "env-from-host", "default-resume"},
			check: func(t *testing.T, c Config) {
				if c.Volumes != nil || c.EnvFromHost != nil || c.DefaultResume != nil {
					t.Errorf("personal options kept: %+v", c)
				}
				if c.Env["DEBUG"] != "1" {
					t.Errorf("env dropped: %v", c.Env)
				}
			},
		},
		{
			name:  "personal options in local",
			layer: layerLocal,
			src:   "volumes: [\"~/data:/data\"]\nkits:\n  docker: {}\n  ssh: {isolation: shared}\n",
			check: func(t *testing.T, c Config) {
				if len(c.Volumes) != 1 || !c.KitActive("docker") || c.KitOption("ssh").Isolation != "shared" {
					t.Errorf("personal options dropped in local layer: %+v", c)
				}
			},
		},
		{
			name:  "global options in local",
			layer: layerLocal,
			src:   "release-channel: dev\nversion-check-interval: 1m\nagent: codex\n",
			want:  []string{"release-channel", "version-check-interval"},
			check: func(t *testing.T, c Config) {
				if c.ReleaseChannel != "" || c.Agent != "codex" {
					t.Errorf("got %+v", c)
				}
			},
		},
		{
			name:  "kit options in project",
			layer: layerProject,
			src:   "kits:\n  ssh:\n    isolation: shared\n  java:\n    credentials: auto\n    default-version: \"21\"\n  browser-open:\n    schemes: [vscode]\n  shell:\n    allow-agent-terminal-title: true\n",
			want:  []string{"kits.ssh.isolation", "kits.java.credentials", "kits.browser-open.schemes", "kits.shell.allow-agent-terminal-title"},
			check: func(t *testing.T, c Config) {
				if c.KitOption("ssh").Isolation != "" || c.KitOption("java").Credentials != nil {
					t.Errorf("kit options kept")
				}
				if c.KitOption("java").DefaultVersion != "21" {
					t.Errorf("default-version dropped")
				}
			},
		},
		{
			name:  "agent options in project",
			layer: layerProject,
			src:   "agent: codex\nagents:\n  claude:\n    config: shared\n    seed: full\n    companions: [codex]\n",
			want:  []string{"agents.claude.config", "agents.claude.seed", "agents.claude.companions"},
			check: func(t *testing.T, c Config) {
				if c.AgentIsolation("claude") != "" || c.Agent != "codex" {
					t.Errorf("got %+v", c)
				}
			},
		},
		{
			name:  "personal kits enabled in project",
			layer: layerProject,
			src:   "kits:\n  docker: {}\n  dropshare:\n  iterm: {}\n  cx:\n    packages: [go]\n  rtk: {}\n  python: {}\n",
			want:  []string{"kits.docker", "kits.dropshare", "kits.iterm", "kits.cx", "kits.rtk"},
			check: func(t *testing.T, c Config) {
				if c.KitActive("docker") || c.KitActive("cx") || !c.KitActive("python") {
					t.Errorf("kits = %v", c.KitNames())
				}
			},
		},
		{
			name:  "personal kit disabled in project",
			layer: layerProject,
			src:   "kits:\n  docker:\n    disabled: true\n",
			check: func(t *testing.T, c Config) {
				if !c.DisabledKits()["docker"] {
					t.Errorf("disable dropped")
				}
			},
		},
		{
			name:  "personal kit disabled false in project",
			layer: layerProject,
			src:   "kits:\n  docker:\n    disabled: false\n",
			want:  []string{"kits.docker"},
		},
		{
			name:  "global layer accepts everything",
			layer: layerGlobal,
			src:   "release-channel: dev\nvolumes: [\"~/x:/x\"]\nkits:\n  docker: {}\n",
		},
		{
			name:  "unknown keys pass through",
			layer: layerProject,
			src:   "something: 1\nkits:\n  unknown-kit:\n    whatever: 2\n",
		},
		{
			name:  "empty document",
			layer: layerProject,
			src:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, rejected := filterYAML(t, tt.layer, tt.src)
			if got := rejectedPaths(rejected); !slices.Equal(got, tt.want) {
				t.Errorf("rejected = %v, want %v", got, tt.want)
			}
			if tt.check != nil {
				tt.check(t, cfg)
			}
		})
	}
}

func TestKitOptionScopeOverride(t *testing.T) {
	kit.Register(&kit.Kit{Name: "scopetest", OptionScopes: map[string]kit.Scope{"servers": kit.ScopeGlobal, "isolation": kit.ScopeAny}})
	_, rejected := filterYAML(t, layerLocal, "kits:\n  scopetest:\n    servers: {a: {}}\n")
	if got := rejectedPaths(rejected); !slices.Equal(got, []string{"kits.scopetest.servers"}) {
		t.Errorf("rejected = %v", got)
	}
	_, rejected = filterYAML(t, layerProject, "kits:\n  scopetest:\n    isolation: shared\n")
	if len(rejected) != 0 {
		t.Errorf("kit override to any ignored: %v", rejectedPaths(rejected))
	}
}

func TestFilterScopesRejection(t *testing.T) {
	_, rejected := filterYAML(t, layerProject, "agent: codex\nkits:\n  ssh:\n    isolation: shared\nvolumes:\n  - ~/.m2/settings.xml:/home/claude/.m2/settings.xml\n")
	if len(rejected) != 2 {
		t.Fatalf("rejected = %v", rejectedPaths(rejected))
	}
	ssh, vol := rejected[0], rejected[1]
	if ssh.Line != 4 || ssh.File != ".asylum" || ssh.Snippet != "kits:\n  ssh:\n    isolation: shared" {
		t.Errorf("ssh rejection = %+v", ssh)
	}
	if vol.Line != 5 || vol.Snippet != "volumes:\n  - ~/.m2/settings.xml:/home/claude/.m2/settings.xml" {
		t.Errorf("volume rejection = %+v", vol)
	}
	if vol.AcceptedIn() != ".asylum.local or ~/.asylum/config.yaml" {
		t.Errorf("AcceptedIn = %q", vol.AcceptedIn())
	}
	if (Rejection{Scope: kit.ScopeGlobal}).AcceptedIn() != "~/.asylum/config.yaml" {
		t.Error("global AcceptedIn")
	}
}

// Aliases and merge keys could pull an entry in from a key the filter does not
// inspect, so they are refused where entries are checked.
func TestFilterScopesAliases(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		wantErr bool
	}{
		{"top-level merge key", "x: &v {volumes: [\"~/.ssh:/s\"]}\n<<: *v\n", true},
		{"aliased kits", "x: &k {docker: {}}\nkits: *k\n", true},
		{"aliased kit entry", "x: &k {packages: [a]}\nkits:\n  python: *k\n", true},
		{"merge key in kit", "x: &k {isolation: shared}\nkits:\n  ssh:\n    <<: *k\n", true},
		{"merge key in agent", "x: &a {config: shared}\nagents:\n  claude:\n    <<: *a\n", true},
		{"aliased leaf value", "x: &p [a, b]\nkits:\n  apt:\n    packages: *p\n", false},
		{"alias as top-level key", "x: &k volumes\n*k : [\"~/.ssh:/s\"]\n", true},
		{"alias as kit name", "x: &d docker\nkits:\n  *d : {}\n", true},
		{"binary key", "!!binary dm9sdW1lcw== : [\"~/.ssh:/s\"]\n", true},
		{"binary option key", "kits:\n  ssh:\n    !!binary aXNvbGF0aW9u : shared\n", true},
		{"alias key posing as disabled", "x: &disabled packages\nkits:\n  docker: {*disabled : true}\n", true},
		{"quoted and tagged string keys", "\"agent\": codex\n!!str env: {A: b}\n", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var doc yaml.Node
			if err := yaml.Unmarshal([]byte(tt.src), &doc); err != nil {
				t.Fatal(err)
			}
			_, err := filterScopes(&doc, layerProject, ".asylum")
			if (err != nil) != tt.wantErr {
				t.Errorf("err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadScopes(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(homeDir, ".asylum"), 0755)
	os.WriteFile(filepath.Join(homeDir, ".asylum", "config.yaml"), []byte("version: \"0.2\"\nkits:\n  ssh:\n    isolation: isolated\n  browser-open:\n    schemes: [dropshare5]\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".asylum"), []byte("kits:\n  ssh:\n    isolation: shared\n  browser-open:\n    schemes: [vscode]\n"), 0644)

	cfg, err := Load(dir, CLIFlags{}, "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := cfg.KitOption("ssh").Isolation; got != "isolated" {
		t.Errorf("ssh isolation = %q, want isolated", got)
	}
	if got := cfg.KitOption("browser-open").Schemes; !slices.Equal(got, []string{"dropshare5"}) {
		t.Errorf("schemes = %v, project layer must be ignored", got)
	}
	if got := rejectedPaths(cfg.Rejected); !slices.Equal(got, []string{"kits.ssh.isolation", "kits.browser-open.schemes"}) {
		t.Errorf("rejected = %v", got)
	}
	if cfg.Rejected[0].File != ".asylum" {
		t.Errorf("file = %q, want .asylum", cfg.Rejected[0].File)
	}

	os.Remove(filepath.Join(dir, ".asylum"))
	os.WriteFile(filepath.Join(dir, ".asylum.local"), []byte("kits:\n  browser-open:\n    schemes: [vscode]\n"), 0644)
	cfg, err = Load(dir, CLIFlags{}, "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := cfg.KitOption("browser-open").Schemes; !slices.Equal(got, []string{"dropshare5", "vscode"}) {
		t.Errorf("schemes = %v, global and local layers must accumulate", got)
	}
	if cfg.Rejected != nil {
		t.Errorf("rejected = %v", rejectedPaths(cfg.Rejected))
	}
}

func TestLoadCLIFlagsNotScoped(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg, err := Load(t.TempDir(), CLIFlags{Volumes: []string{"~/data:/data"}}, "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !slices.Equal(cfg.Volumes, []string{"~/data:/data"}) || cfg.Rejected != nil {
		t.Errorf("volumes = %v, rejected = %v", cfg.Volumes, cfg.Rejected)
	}
}
