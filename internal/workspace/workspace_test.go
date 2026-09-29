package workspace

import (
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"
)

func TestUnsafe(t *testing.T) {
	home := "/home/alice"
	tests := []struct {
		name string
		dir  string
		want bool
	}{
		{"exact home", "/home/alice", true},
		{"home with trailing slash", "/home/alice/", true},
		{"filesystem root", "/", true},
		{"home subdir", "/home/alice/projects/foo", false},
		{"unrelated dir", "/tmp/work", false},
		{"sibling of home", "/home/bob", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := unsafe(tt.dir, home); got != tt.want {
				t.Errorf("unsafe(%q, %q) = %v, want %v", tt.dir, home, got, tt.want)
			}
		})
	}
}

func TestNameFormat(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	got := name("2026-06-23", r)
	re := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}(-[a-z]+){3}$`)
	if !re.MatchString(got) {
		t.Errorf("name = %q, does not match expected format", got)
	}
}

func TestResolveSafe(t *testing.T) {
	dir, outcome, err := Resolve("/home/alice/projects/foo", "/home/alice", false)
	if err != nil {
		t.Fatal(err)
	}
	if outcome != Unchanged {
		t.Errorf("outcome = %v, want Unchanged", outcome)
	}
	if dir != "/home/alice/projects/foo" {
		t.Errorf("safe dir changed to %q", dir)
	}
}

func TestResolveRedirectsHome(t *testing.T) {
	home := t.TempDir()
	dir, outcome, err := Resolve(home, home, false)
	if err != nil {
		t.Fatal(err)
	}
	if outcome != Fresh {
		t.Fatalf("outcome = %v, want Fresh", outcome)
	}
	if got := filepath.Dir(dir); got != filepath.Join(home, "asylum-workspace") {
		t.Errorf("workspace parent = %q, want under asylum-workspace", got)
	}
	if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
		t.Errorf("workspace dir not created: %v", err)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Error("workspace dir should be empty")
	}
}

func TestResolveRedirectsRoot(t *testing.T) {
	home := t.TempDir()
	_, outcome, err := Resolve("/", home, false)
	if err != nil {
		t.Fatal(err)
	}
	if outcome != Fresh {
		t.Errorf("outcome = %v, want Fresh", outcome)
	}
}

func TestResolveCollisionReroll(t *testing.T) {
	base := t.TempDir()
	date := "2026-06-23"

	// Predict the first name the rng will produce, then pre-create it so
	// create must re-roll past the collision.
	taken := name(date, rand.New(rand.NewSource(42)))
	if err := os.MkdirAll(filepath.Join(base, taken), 0o755); err != nil {
		t.Fatal(err)
	}

	dir, err := create(base, date, rand.New(rand.NewSource(42)))
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(dir) == taken {
		t.Errorf("collision not avoided: reused %q", taken)
	}
}

func TestResolveReuse(t *testing.T) {
	const older, newer = "2026-09-01-red-fox-jumps", "2026-09-02-blue-owl-sings"
	tests := []struct {
		name    string
		entries map[string]time.Duration // directory → age
		want    string                   // expected base name; "" means a fresh workspace
	}{
		{"newest by mtime wins", map[string]time.Duration{older: 2 * time.Hour, newer: time.Hour}, newer},
		{"mtime beats name date", map[string]time.Duration{older: time.Hour, newer: 2 * time.Hour}, older},
		{"non-matching name ignored", map[string]time.Duration{older: 2 * time.Hour, "notes": 0}, older},
		{"no workspace falls back to fresh", map[string]time.Duration{"notes": 0}, ""},
		{"missing base falls back to fresh", nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			base := filepath.Join(home, "asylum-workspace")
			for n, age := range tt.entries {
				dir := filepath.Join(base, n)
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatal(err)
				}
				mtime := time.Now().Add(-age)
				if err := os.Chtimes(dir, mtime, mtime); err != nil {
					t.Fatal(err)
				}
			}

			dir, outcome, err := Resolve(home, home, true)
			if err != nil {
				t.Fatal(err)
			}
			if tt.want == "" {
				if outcome != Fresh || !namePattern.MatchString(filepath.Base(dir)) {
					t.Errorf("got %q (outcome %v), want a fresh workspace", dir, outcome)
				}
				return
			}
			if outcome != Reused || dir != filepath.Join(base, tt.want) {
				t.Errorf("got %q (outcome %v), want reused %q", dir, outcome, tt.want)
			}
		})
	}
}

func TestResolveReuseSafeDirUnchanged(t *testing.T) {
	dir, outcome, err := Resolve("/home/alice/projects/foo", "/home/alice", true)
	if err != nil {
		t.Fatal(err)
	}
	if outcome != Unchanged || dir != "/home/alice/projects/foo" {
		t.Errorf("got %q (outcome %v), want unchanged", dir, outcome)
	}
}
