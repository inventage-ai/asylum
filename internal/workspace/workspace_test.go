package workspace

import (
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"testing"
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
	dir, redirected, err := Resolve("/home/alice/projects/foo", "/home/alice")
	if err != nil {
		t.Fatal(err)
	}
	if redirected {
		t.Error("safe dir should not be redirected")
	}
	if dir != "/home/alice/projects/foo" {
		t.Errorf("safe dir changed to %q", dir)
	}
}

func TestResolveRedirectsHome(t *testing.T) {
	home := t.TempDir()
	dir, redirected, err := Resolve(home, home)
	if err != nil {
		t.Fatal(err)
	}
	if !redirected {
		t.Fatal("home should be redirected")
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
	_, redirected, err := Resolve("/", home)
	if err != nil {
		t.Fatal(err)
	}
	if !redirected {
		t.Error("filesystem root should be redirected")
	}
}

func TestResolveCollisionReroll(t *testing.T) {
	base := t.TempDir()
	date := "2026-06-23"

	// Predict the first name the rng will produce, then pre-create it so
	// resolveAt must re-roll past the collision.
	taken := name(date, rand.New(rand.NewSource(42)))
	if err := os.MkdirAll(filepath.Join(base, taken), 0o755); err != nil {
		t.Fatal(err)
	}

	dir, redirected, err := resolveAt(base, date, rand.New(rand.NewSource(42)))
	if err != nil {
		t.Fatal(err)
	}
	if !redirected {
		t.Fatal("expected redirect")
	}
	if filepath.Base(dir) == taken {
		t.Errorf("collision not avoided: reused %q", taken)
	}
}
