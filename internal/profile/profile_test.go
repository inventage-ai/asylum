package profile

import (
	"testing"
)

func setupTestRegistry() func() {
	old := make(map[string]*Profile, len(registry))
	for k, v := range registry {
		old[k] = v
	}

	// Clear and register test profiles
	for k := range registry {
		delete(registry, k)
	}

	Register(&Profile{
		Name: "alpha",
		SubProfiles: map[string]*Profile{
			"sub1": {Name: "alpha/sub1"},
			"sub2": {Name: "alpha/sub2"},
		},
	})
	Register(&Profile{
		Name:        "beta",
		SubProfiles: map[string]*Profile{},
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

func profileNames(profiles []*Profile) []string {
	names := make([]string, len(profiles))
	for i, p := range profiles {
		names[i] = p.Name
	}
	return names
}

func TestResolve_NilMeansAll(t *testing.T) {
	cleanup := setupTestRegistry()
	defer cleanup()

	profiles, err := Resolve(nil)
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
	profiles, err := Resolve(&empty)
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
	profiles, err := Resolve(&names)
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
	profiles, err := Resolve(&names)
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
	profiles, err := Resolve(&names)
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
	profiles, err := Resolve(&names)
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
	_, err := Resolve(&names)
	if err == nil {
		t.Fatal("expected error for unknown profile")
	}
}

func TestAggregateCacheDirs(t *testing.T) {
	profiles := []*Profile{
		{Name: "a", CacheDirs: map[string]string{"npm": "/home/claude/.npm"}},
		{Name: "b", CacheDirs: map[string]string{"pip": "/home/claude/.cache/pip"}},
		{Name: "c"}, // no cache dirs
	}
	dirs := AggregateCacheDirs(profiles)
	if len(dirs) != 2 {
		t.Fatalf("expected 2 cache dirs, got %d", len(dirs))
	}
	if dirs["npm"] != "/home/claude/.npm" {
		t.Errorf("npm = %q", dirs["npm"])
	}
	if dirs["pip"] != "/home/claude/.cache/pip" {
		t.Errorf("pip = %q", dirs["pip"])
	}
}

func TestAggregateCacheDirs_Empty(t *testing.T) {
	dirs := AggregateCacheDirs(nil)
	if len(dirs) != 0 {
		t.Fatalf("expected 0 cache dirs, got %d", len(dirs))
	}
}

func TestResolve_UnknownSubProfile(t *testing.T) {
	cleanup := setupTestRegistry()
	defer cleanup()

	names := []string{"alpha/unknown"}
	_, err := Resolve(&names)
	if err == nil {
		t.Fatal("expected error for unknown sub-profile")
	}
}
