package ports

import (
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

	Allocate("/proj/a", "asylum-aaa", 5) // 10000-10004
	Allocate("/proj/b", "asylum-bbb", 5) // 10005-10009

	// Try to extend a to 8 — blocked by b
	r, err := Allocate("/proj/a", "asylum-aaa", 8)
	if err != nil {
		t.Fatal(err)
	}
	if r.Count != 5 {
		t.Errorf("count = %d, want 5 (extension should be blocked)", r.Count)
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
	r := Range{Start: 10000, Count: 3}
	got := r.Ports()
	want := []int{10000, 10001, 10002}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}
