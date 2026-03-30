package kit

import (
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
	got := AssembleRulesSnippets(kits)
	want := "## A\nTools from A.\n\n## C\nTools from C.\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestAssembleRulesSnippets_Empty(t *testing.T) {
	got := AssembleRulesSnippets(nil)
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
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
