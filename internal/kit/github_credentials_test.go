package kit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGithubCredentialFunc_DirExists(t *testing.T) {
	home := t.TempDir()
	ghDir := filepath.Join(home, ".config", "gh")
	os.MkdirAll(ghDir, 0755)
	os.WriteFile(filepath.Join(ghDir, "hosts.yml"), []byte("github.com:\n    oauth_token: test\n"), 0600)

	mounts, err := githubCredentialFunc(CredentialOpts{HomeDir: home})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mounts) != 1 {
		t.Fatalf("expected 1 mount, got %d", len(mounts))
	}
	if mounts[0].HostPath != ghDir {
		t.Errorf("HostPath = %q, want %q", mounts[0].HostPath, ghDir)
	}
	if mounts[0].Destination != "~/.config/gh" {
		t.Errorf("Destination = %q, want %q", mounts[0].Destination, "~/.config/gh")
	}
}

func TestGithubCredentialFunc_DirMissing(t *testing.T) {
	home := t.TempDir()

	mounts, err := githubCredentialFunc(CredentialOpts{HomeDir: home})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mounts) != 0 {
		t.Fatalf("expected 0 mounts, got %d", len(mounts))
	}
}
