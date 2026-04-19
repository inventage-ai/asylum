package kit

import (
	"fmt"
	"strings"
	"testing"
)

func setupTestRegistry() func() {
	old := make(map[string]*Kit, len(registry))
	for k, v := range registry {
		old[k] = v
	}

	// Clear and register test profiles
	for k := range registry {
		delete(registry, k)
	}

	Register(&Kit{
		Name: "alpha",
		SubKits: map[string]*Kit{
			"sub1": {Name: "alpha/sub1"},
			"sub2": {Name: "alpha/sub2"},
		},
	})
	Register(&Kit{
		Name:        "beta",
		SubKits: map[string]*Kit{},
	})

	return func() {
		for k := range registry {
			delete(registry, k)
		}
		for k, v := range old {
			registry[k] = v
		}
	}
}

func profileNames(profiles []*Kit) []string {
	names := make([]string, len(profiles))
	for i, p := range profiles {
		names[i] = p.Name
	}
	return names
}

func TestResolve_NilMeansAll(t *testing.T) {
	cleanup := setupTestRegistry()
	defer cleanup()

	profiles, err := Resolve(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	got := profileNames(profiles)
	want := []string{"alpha", "alpha/sub1", "alpha/sub2", "beta"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestResolve_EmptyMeansNone(t *testing.T) {
	cleanup := setupTestRegistry()
	defer cleanup()

	empty := []string{}
	profiles, err := Resolve(empty, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 0 {
		t.Fatalf("expected no profiles, got %v", profileNames(profiles))
	}
}

func TestResolve_TopLevelActivatesAllChildren(t *testing.T) {
	cleanup := setupTestRegistry()
	defer cleanup()

	names := []string{"alpha"}
	profiles, err := Resolve(names, nil)
	if err != nil {
		t.Fatal(err)
	}
	got := profileNames(profiles)
	want := []string{"alpha", "alpha/sub1", "alpha/sub2"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestResolve_ChildActivatesParentOnly(t *testing.T) {
	cleanup := setupTestRegistry()
	defer cleanup()

	names := []string{"alpha/sub1"}
	profiles, err := Resolve(names, nil)
	if err != nil {
		t.Fatal(err)
	}
	got := profileNames(profiles)
	want := []string{"alpha", "alpha/sub1"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestResolve_Deduplication(t *testing.T) {
	cleanup := setupTestRegistry()
	defer cleanup()

	// Both "alpha" and "alpha/sub1" — alpha should appear only once
	names := []string{"alpha", "alpha/sub1"}
	profiles, err := Resolve(names, nil)
	if err != nil {
		t.Fatal(err)
	}
	got := profileNames(profiles)
	want := []string{"alpha", "alpha/sub1", "alpha/sub2"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestResolve_AllChildrenEquivalent(t *testing.T) {
	cleanup := setupTestRegistry()
	defer cleanup()

	// Selecting all children individually should be same as parent
	names := []string{"alpha/sub1", "alpha/sub2"}
	profiles, err := Resolve(names, nil)
	if err != nil {
		t.Fatal(err)
	}
	got := profileNames(profiles)
	want := []string{"alpha", "alpha/sub1", "alpha/sub2"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestResolve_UnknownProfile(t *testing.T) {
	cleanup := setupTestRegistry()
	defer cleanup()

	names := []string{"unknown"}
	_, err := Resolve(names, nil)
	if err == nil {
		t.Fatal("expected error for unknown profile")
	}
}

func TestAggregateCacheDirs(t *testing.T) {
	profiles := []*Kit{
		{Name: "a", CacheDirs: map[string]string{"npm": "~/.npm"}},
		{Name: "b", CacheDirs: map[string]string{"pip": "~/.cache/pip"}},
		{Name: "c"}, // no cache dirs
	}
	dirs := AggregateCacheDirs(profiles)
	if len(dirs) != 2 {
		t.Fatalf("expected 2 cache dirs, got %d", len(dirs))
	}
	if dirs["npm"] != "~/.npm" {
		t.Errorf("npm = %q", dirs["npm"])
	}
	if dirs["pip"] != "~/.cache/pip" {
		t.Errorf("pip = %q", dirs["pip"])
	}
}

func TestAggregateCacheDirs_Empty(t *testing.T) {
	dirs := AggregateCacheDirs(nil)
	if len(dirs) != 0 {
		t.Fatalf("expected 0 cache dirs, got %d", len(dirs))
	}
}

func TestAssembleRulesSnippets(t *testing.T) {
	kits := []*Kit{
		{Name: "a", RulesSnippet: "## A\nTools from A.\n"},
		{Name: "b", RulesSnippet: ""},
		{Name: "c", RulesSnippet: "## C\nTools from C."},
	}
	got := AssembleRulesSnippets(kits, nil)
	want := "## A\nTools from A.\n\n## C\nTools from C.\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestAssembleRulesSnippets_Empty(t *testing.T) {
	got := AssembleRulesSnippets(nil, nil)
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestSnippetFuncFallback(t *testing.T) {
	t.Run("func overrides static string", func(t *testing.T) {
		kits := []*Kit{{
			Name:          "test",
			DockerSnippet: "STATIC\n",
			DockerSnippetFunc: func(sc *SnippetConfig) string {
				return "DYNAMIC\n"
			},
		}}
		got := AssembleDockerSnippets(kits, nil)
		if got != "DYNAMIC\n" {
			t.Errorf("got %q, want DYNAMIC", got)
		}
	})

	t.Run("static used when no func", func(t *testing.T) {
		kits := []*Kit{{Name: "test", DockerSnippet: "STATIC\n"}}
		got := AssembleDockerSnippets(kits, nil)
		if got != "STATIC\n" {
			t.Errorf("got %q, want STATIC", got)
		}
	})

	t.Run("func receives nil config when no accessor", func(t *testing.T) {
		var received *SnippetConfig
		kits := []*Kit{{
			Name: "test",
			DockerSnippetFunc: func(sc *SnippetConfig) string {
				received = sc
				return "OK\n"
			},
		}}
		AssembleDockerSnippets(kits, nil)
		if received != nil {
			t.Error("expected nil SnippetConfig when kitConfig accessor is nil")
		}
	})

	t.Run("func receives config from accessor", func(t *testing.T) {
		var received *SnippetConfig
		kits := []*Kit{{
			Name: "test",
			DockerSnippetFunc: func(sc *SnippetConfig) string {
				received = sc
				return "OK\n"
			},
		}}
		accessor := func(name string) *SnippetConfig {
			if name == "test" {
				return &SnippetConfig{Versions: []string{"1"}}
			}
			return nil
		}
		AssembleDockerSnippets(kits, accessor)
		if received == nil || len(received.Versions) != 1 {
			t.Error("expected SnippetConfig with versions to be passed to func")
		}
	})

	t.Run("rules func fallback", func(t *testing.T) {
		kits := []*Kit{{
			Name:         "a",
			RulesSnippet: "STATIC\n",
			RulesSnippetFunc: func(sc *SnippetConfig) string {
				return "DYNAMIC\n"
			},
		}}
		got := AssembleRulesSnippets(kits, nil)
		if got != "DYNAMIC\n" {
			t.Errorf("got %q, want DYNAMIC", got)
		}
	})
}

func TestAssembleProjectSnippets(t *testing.T) {
	kits := []*Kit{
		{Name: "a", ProjectSnippetFunc: func(sc *SnippetConfig) string { return "RUN a\n" }},
		{Name: "b"}, // no func
		{Name: "c", ProjectSnippetFunc: func(sc *SnippetConfig) string { return "" }}, // empty return
	}
	got := AssembleProjectSnippets(kits, nil)
	if got != "RUN a\n" {
		t.Errorf("got %q, want %q", got, "RUN a\n")
	}
}

func TestAssembleEnvVars(t *testing.T) {
	kits := []*Kit{
		{Name: "a", EnvFunc: func(sc *SnippetConfig) map[string]string { return map[string]string{"A": "1"} }},
		{Name: "b"}, // no func
		{Name: "c", EnvFunc: func(sc *SnippetConfig) map[string]string { return nil }},
	}
	got := AssembleEnvVars(kits, nil)
	if len(got) != 1 || got["A"] != "1" {
		t.Errorf("got %v, want {A: 1}", got)
	}

	// No kits with env funcs
	got = AssembleEnvVars([]*Kit{{Name: "x"}}, nil)
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestAssembleConfigSnippets(t *testing.T) {
	cleanup := setupTestRegistry()
	defer cleanup()

	registry["alpha"].ConfigSnippet = "  alpha:\n"
	Register(&Kit{Name: "commented", ConfigSnippet: "  # commented:\n"})
	defer delete(registry, "commented")

	got := AssembleConfigSnippets()

	// Active snippets come first
	if !strings.Contains(got, "  alpha:\n") {
		t.Error("expected active snippet for alpha")
	}
	// Commented snippets come after active ones
	if !strings.Contains(got, "  # commented:\n") {
		t.Error("expected commented snippet")
	}
	// Active comes before commented
	alphaIdx := strings.Index(got, "alpha:")
	commentIdx := strings.Index(got, "# commented:")
	if alphaIdx > commentIdx {
		t.Error("active snippets should come before commented snippets")
	}
}

func TestAggregateTools(t *testing.T) {
	kits := []*Kit{
		{Name: "github", Tools: []string{"gh"}},
		{Name: "java"},
		{Name: "java/maven", Tools: []string{"mvn"}},
		{Name: "node/pnpm", Tools: []string{"pnpm"}},
	}
	got := AggregateTools(kits)
	want := []string{"gh (github)", "mvn (java/maven)", "pnpm (node/pnpm)"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestAggregateTools_Empty(t *testing.T) {
	got := AggregateTools(nil)
	if len(got) != 0 {
		t.Errorf("expected empty, got %v", got)
	}
}

func TestResolve_UnknownSubProfile(t *testing.T) {
	cleanup := setupTestRegistry()
	defer cleanup()

	names := []string{"alpha/unknown"}
	_, err := Resolve(names, nil)
	if err == nil {
		t.Fatal("expected error for unknown sub-profile")
	}
}

func TestResolve_DefaultOnIncluded(t *testing.T) {
	cleanup := setupTestRegistry()
	defer cleanup()

	Register(&Kit{Name: "defkit", Tier: TierAlwaysOn})
	defer delete(registry, "defkit")

	// Explicit kits should also include default-on
	names := []string{"alpha"}
	result, err := Resolve(names, nil)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, k := range result {
		if k.Name == "defkit" {
			found = true
		}
	}
	if !found {
		t.Error("default-on kit should be included when explicit kits are listed")
	}
}

func TestResolve_DefaultOnExcludedWhenDisabled(t *testing.T) {
	cleanup := setupTestRegistry()
	defer cleanup()

	Register(&Kit{Name: "defkit", Tier: TierAlwaysOn})
	defer delete(registry, "defkit")

	names := []string{"alpha"}
	disabled := map[string]bool{"defkit": true}
	result, err := Resolve(names, disabled)
	if err != nil {
		t.Fatal(err)
	}
	for _, k := range result {
		if k.Name == "defkit" {
			t.Error("disabled default-on kit should NOT be included")
		}
	}
}

func TestResolve_DefaultOnNotAddedToEmpty(t *testing.T) {
	cleanup := setupTestRegistry()
	defer cleanup()

	Register(&Kit{Name: "defkit", Tier: TierAlwaysOn})
	defer delete(registry, "defkit")

	// Empty slice = no kits, default-on NOT added
	result, err := Resolve([]string{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 0 {
		t.Errorf("empty slice should resolve to no kits, got %d", len(result))
	}
}

func TestResolve_DependencyAutoActivated(t *testing.T) {
	cleanup := setupTestRegistry()
	defer cleanup()

	Register(&Kit{Name: "depkit", Deps: []string{"alpha"}})
	defer delete(registry, "depkit")

	// Only depkit is listed; alpha should be auto-activated as a dependency
	names := []string{"depkit"}
	result, err := Resolve(names, nil)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, k := range result {
		if k.Name == "alpha" {
			found = true
		}
	}
	if !found {
		t.Error("dependency 'alpha' should be auto-activated")
	}
}

func TestResolve_DependencySatisfied(t *testing.T) {
	cleanup := setupTestRegistry()
	defer cleanup()

	Register(&Kit{Name: "depkit", Deps: []string{"alpha"}})
	defer delete(registry, "depkit")

	// alpha is active, so depkit's dependency is satisfied — no error
	names := []string{"alpha", "depkit"}
	_, err := Resolve(names, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestResolve_MissingDependency(t *testing.T) {
	cleanup := setupTestRegistry()
	defer cleanup()

	Register(&Kit{Name: "depkit", Deps: []string{"missing"}})
	defer delete(registry, "depkit")

	// "missing" is not in the registry as a kit, but depkit references it.
	// Resolution succeeds (warn only, don't block).
	names := []string{"alpha", "depkit"}
	result, err := Resolve(names, nil)
	if err != nil {
		t.Fatal(err)
	}
	// depkit should still be in the result
	found := false
	for _, k := range result {
		if k.Name == "depkit" {
			found = true
		}
	}
	if !found {
		t.Error("kit with missing dep should still be included (warn only)")
	}
}

func TestResolve_DisabledKitExcluded(t *testing.T) {
	cleanup := setupTestRegistry()
	defer cleanup()

	names := []string{"alpha", "beta"}
	disabled := map[string]bool{"beta": true}
	result, err := Resolve(names, disabled)
	if err != nil {
		t.Fatal(err)
	}
	for _, k := range result {
		if k.Name == "beta" {
			t.Error("disabled kit should not be in result")
		}
	}
}

type stubConfig struct{ portCount int }

func (s stubConfig) PortCount() int { return s.portCount }

func TestAggregateContainerArgs(t *testing.T) {
	opts := ContainerOpts{
		ProjectDir:    "/proj",
		ContainerName: "asylum-test",
		HomeDir:       "/home/test",
		Config:        stubConfig{portCount: 5},
	}

	t.Run("collects args from kits with ContainerFunc", func(t *testing.T) {
		kits := []*Kit{
			{
				Name: "kitA",
				ContainerFunc: func(ContainerOpts) ([]RunArg, error) {
					return []RunArg{{Flag: "-e", Value: "A=1", Source: "kitA", Priority: PriorityKit}}, nil
				},
			},
			{
				Name: "kitB",
				ContainerFunc: func(ContainerOpts) ([]RunArg, error) {
					return []RunArg{{Flag: "-e", Value: "B=2", Source: "kitB", Priority: PriorityKit}}, nil
				},
			},
		}
		args := AggregateContainerArgs(kits, opts)
		if len(args) != 2 {
			t.Fatalf("expected 2 args, got %d", len(args))
		}
		if args[0].Value != "A=1" || args[1].Value != "B=2" {
			t.Errorf("unexpected args: %v", args)
		}
	})

	t.Run("skips kits without ContainerFunc", func(t *testing.T) {
		kits := []*Kit{
			{Name: "noFunc"},
			{
				Name: "hasFunc",
				ContainerFunc: func(ContainerOpts) ([]RunArg, error) {
					return []RunArg{{Flag: "-e", Value: "X=1", Source: "hasFunc", Priority: PriorityKit}}, nil
				},
			},
		}
		args := AggregateContainerArgs(kits, opts)
		if len(args) != 1 {
			t.Fatalf("expected 1 arg, got %d", len(args))
		}
	})

	t.Run("logs warning and skips on error", func(t *testing.T) {
		kits := []*Kit{
			{
				Name: "failing",
				ContainerFunc: func(ContainerOpts) ([]RunArg, error) {
					return nil, fmt.Errorf("allocation failed")
				},
			},
			{
				Name: "working",
				ContainerFunc: func(ContainerOpts) ([]RunArg, error) {
					return []RunArg{{Flag: "-p", Value: "8080:8080", Source: "working", Priority: PriorityKit}}, nil
				},
			},
		}
		args := AggregateContainerArgs(kits, opts)
		if len(args) != 1 {
			t.Fatalf("expected 1 arg (failing kit skipped), got %d", len(args))
		}
		if args[0].Source != "working" {
			t.Errorf("expected working kit arg, got source %q", args[0].Source)
		}
	})
}

func TestJavaSnippetGeneration(t *testing.T) {
	java := Get("java")
	if java == nil {
		t.Fatal("java kit not registered")
	}

	t.Run("default versions when nil config", func(t *testing.T) {
		s := java.DockerSnippetFunc(nil)
		if !strings.Contains(s, "java@17") || !strings.Contains(s, "java@21") || !strings.Contains(s, "java@25") {
			t.Errorf("expected default versions 17/21/25, got: %s", s)
		}
		if !strings.Contains(s, "java@21\n") {
			t.Errorf("expected default version 21, got: %s", s)
		}
	})

	t.Run("custom versions from config", func(t *testing.T) {
		sc := &SnippetConfig{Versions: []string{"21"}, DefaultVersion: "21"}
		s := java.DockerSnippetFunc(sc)
		if !strings.Contains(s, "java@21") {
			t.Errorf("expected java@21, got: %s", s)
		}
		if strings.Contains(s, "java@17") || strings.Contains(s, "java@25") {
			t.Errorf("should not contain non-configured versions, got: %s", s)
		}
	})

	t.Run("rules reflect configured versions", func(t *testing.T) {
		sc := &SnippetConfig{Versions: []string{"17", "25"}, DefaultVersion: "25"}
		s := java.RulesSnippetFunc(sc)
		if !strings.Contains(s, "17, 25") || !strings.Contains(s, "default is 25") {
			t.Errorf("rules should reflect configured versions, got: %s", s)
		}
	})

	t.Run("env func returns default version", func(t *testing.T) {
		sc := &SnippetConfig{DefaultVersion: "25"}
		env := java.EnvFunc(sc)
		if env["ASYLUM_JAVA_VERSION"] != "25" {
			t.Errorf("expected ASYLUM_JAVA_VERSION=25, got %v", env)
		}
	})

	t.Run("project snippet empty when default in versions", func(t *testing.T) {
		sc := &SnippetConfig{Versions: []string{"17", "21", "25"}, DefaultVersion: "21"}
		s := java.ProjectSnippetFunc(sc)
		if s != "" {
			t.Errorf("expected empty project snippet, got: %s", s)
		}
	})

	t.Run("project snippet installs non-preinstalled version", func(t *testing.T) {
		sc := &SnippetConfig{Versions: []string{"17", "21", "25"}, DefaultVersion: "11"}
		s := java.ProjectSnippetFunc(sc)
		if !strings.Contains(s, "java@11") {
			t.Errorf("expected mise install java@11, got: %s", s)
		}
	})
}
