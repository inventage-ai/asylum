package ports

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func setup(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	os.MkdirAll(filepath.Join(home, ".asylum"), 0755)
}

func seedRegistry(t *testing.T, ranges []Range) {
	t.Helper()
	path, err := registryPath()
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.MarshalIndent(registry{Ranges: ranges}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
}

func readRegistryFile(t *testing.T) registry {
	t.Helper()
	path, err := registryPath()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var reg registry
	if err := json.Unmarshal(data, &reg); err != nil {
		t.Fatal(err)
	}
	return reg
}

func TestAllocate_FirstProject(t *testing.T) {
	setup(t)

	r, err := Allocate("/proj/a", "asylum-aaa", 5)
	if err != nil {
		t.Fatal(err)
	}
	if r.Start != BasePort {
		t.Errorf("start = %d, want %d", r.Start, BasePort)
	}
	if r.Count != 5 {
		t.Errorf("count = %d, want 5", r.Count)
	}
	if r.Project != "/proj/a" {
		t.Errorf("project = %q", r.Project)
	}
	if r.Container != "asylum-aaa" {
		t.Errorf("container = %q", r.Container)
	}
}

func TestAllocate_Subsequent(t *testing.T) {
	setup(t)

	Allocate("/proj/a", "asylum-aaa", 5)
	r, err := Allocate("/proj/b", "asylum-bbb", 3)
	if err != nil {
		t.Fatal(err)
	}
	if r.Start != BasePort+5 {
		t.Errorf("start = %d, want %d", r.Start, BasePort+5)
	}
	if r.Count != 3 {
		t.Errorf("count = %d, want 3", r.Count)
	}
}

func TestAllocate_Reuse(t *testing.T) {
	setup(t)

	r1, _ := Allocate("/proj/a", "asylum-aaa", 5)
	r2, err := Allocate("/proj/a", "asylum-aaa", 5)
	if err != nil {
		t.Fatal(err)
	}
	if r1 != r2 {
		t.Errorf("expected same range, got %+v and %+v", r1, r2)
	}
}

func TestAllocate_Extend(t *testing.T) {
	setup(t)

	Allocate("/proj/a", "asylum-aaa", 5)

	// Extend from 5 to 8
	r, err := Allocate("/proj/a", "asylum-aaa", 8)
	if err != nil {
		t.Fatal(err)
	}
	if r.Count != 8 {
		t.Errorf("count = %d, want 8", r.Count)
	}
	if r.Start != BasePort {
		t.Errorf("start = %d, want %d", r.Start, BasePort)
	}
}

func TestAllocate_ExtendBlocked(t *testing.T) {
	setup(t)

	Allocate("/proj/a", "asylum-aaa", 5) // BasePort..BasePort+4
	Allocate("/proj/b", "asylum-bbb", 5) // BasePort+5..BasePort+9

	// Try to extend a to 8 — blocked by b
	r, err := Allocate("/proj/a", "asylum-aaa", 8)
	if err != nil {
		t.Fatal(err)
	}
	if r.Count != 5 {
		t.Errorf("count = %d, want 5 (extension should be blocked)", r.Count)
	}
}

func TestAllocate_ReassignsStaleLegacyRange(t *testing.T) {
	setup(t)

	// Seed the registry with a legacy-range entry (Start >= 10000).
	seedRegistry(t, []Range{
		{Project: "/proj/a", Container: "asylum-old", Start: 10000, Count: 5},
	})

	r, err := Allocate("/proj/a", "asylum-new", 5)
	if err != nil {
		t.Fatal(err)
	}
	if r.Start != BasePort {
		t.Errorf("start = %d, want %d (should be reassigned from BasePort)", r.Start, BasePort)
	}
	if r.Container != "asylum-new" {
		t.Errorf("container = %q, want asylum-new (should use current container name)", r.Container)
	}

	// Old entry must be gone; registry should hold exactly the new one.
	reg := readRegistryFile(t)
	if len(reg.Ranges) != 1 {
		t.Fatalf("registry has %d entries, want 1: %+v", len(reg.Ranges), reg.Ranges)
	}
	if reg.Ranges[0].Start != BasePort || reg.Ranges[0].Container != "asylum-new" {
		t.Errorf("registry entry = %+v, want Start=%d Container=asylum-new", reg.Ranges[0], BasePort)
	}
}

func TestNextStart_IgnoresLegacyEntries(t *testing.T) {
	setup(t)

	// One stale legacy entry for another project, plus one fresh entry at BasePort.
	seedRegistry(t, []Range{
		{Project: "/proj/legacy", Container: "asylum-legacy", Start: 10000, Count: 5},
		{Project: "/proj/a", Container: "asylum-aaa", Start: BasePort, Count: 5},
	})

	r, err := Allocate("/proj/b", "asylum-bbb", 5)
	if err != nil {
		t.Fatal(err)
	}
	if r.Start != BasePort+5 {
		t.Errorf("start = %d, want %d (should skip legacy entry)", r.Start, BasePort+5)
	}
}

func TestAllocate_SubLegacyReturnedUnchanged(t *testing.T) {
	setup(t)

	r1, _ := Allocate("/proj/a", "asylum-aaa", 5)
	if r1.Start != BasePort {
		t.Fatalf("seed start = %d, want %d", r1.Start, BasePort)
	}

	r2, err := Allocate("/proj/a", "asylum-aaa", 5)
	if err != nil {
		t.Fatal(err)
	}
	if r2 != r1 {
		t.Errorf("expected unchanged range, got %+v (was %+v)", r2, r1)
	}
}

func TestRenameContainer(t *testing.T) {
	setup(t)

	Allocate("/proj/a", "asylum-aaa", 5)
	Allocate("/proj/b", "asylum-bbb", 5)

	if err := RenameContainer("asylum-aaa", "asylum-aaa-myproject"); err != nil {
		t.Fatal(err)
	}

	// Re-allocate should find the renamed entry
	r, err := Allocate("/proj/a", "asylum-aaa-myproject", 5)
	if err != nil {
		t.Fatal(err)
	}
	if r.Container != "asylum-aaa-myproject" {
		t.Errorf("container = %q, want asylum-aaa-myproject", r.Container)
	}
	if r.Start != BasePort {
		t.Errorf("start = %d, want %d", r.Start, BasePort)
	}

	// Renaming non-existent container is a no-op
	if err := RenameContainer("asylum-zzz", "asylum-zzz-other"); err != nil {
		t.Fatal(err)
	}
}

func TestRange_Ports(t *testing.T) {
	r := Range{Start: 7001, Count: 3}
	got := r.Ports()
	want := []int{7001, 7002, 7003}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}
