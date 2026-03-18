package ssh

import (
	"os"
	"path/filepath"
	"testing"
)

// TestInit_KeyAlreadyExists verifies Init returns nil and skips keygen when
// the key file already exists.
func TestInit_KeyAlreadyExists(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	sshDir := filepath.Join(home, ".asylum", "ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatal(err)
	}
	keyPath := filepath.Join(sshDir, "id_ed25519")
	if err := os.WriteFile(keyPath, []byte("existing key"), 0600); err != nil {
		t.Fatal(err)
	}

	if err := Init(); err != nil {
		t.Fatalf("Init returned error when key exists: %v", err)
	}
}

// TestInit_CopiesKnownHosts verifies that an existing ~/.ssh/known_hosts is
// copied into the asylum ssh directory.
func TestInit_CopiesKnownHosts(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	// Create ~/.ssh/known_hosts
	dotSSH := filepath.Join(home, ".ssh")
	if err := os.MkdirAll(dotSSH, 0700); err != nil {
		t.Fatal(err)
	}
	knownHostsContent := []byte("github.com ssh-ed25519 AAAA...")
	if err := os.WriteFile(filepath.Join(dotSSH, "known_hosts"), knownHostsContent, 0600); err != nil {
		t.Fatal(err)
	}

	// Also pre-create the key so Init doesn't invoke ssh-keygen.
	sshDir := filepath.Join(home, ".asylum", "ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sshDir, "id_ed25519"), []byte("existing key"), 0600); err != nil {
		t.Fatal(err)
	}

	if err := Init(); err != nil {
		t.Fatalf("Init returned error: %v", err)
	}

	dst := filepath.Join(sshDir, "known_hosts")
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("known_hosts not copied: %v", err)
	}
	if string(got) != string(knownHostsContent) {
		t.Errorf("known_hosts content = %q, want %q", got, knownHostsContent)
	}
}

// TestInit_NoKnownHosts verifies Init succeeds when ~/.ssh/known_hosts does
// not exist (no copy attempted).
func TestInit_NoKnownHosts(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	// Pre-create the key so Init doesn't invoke ssh-keygen.
	sshDir := filepath.Join(home, ".asylum", "ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sshDir, "id_ed25519"), []byte("existing key"), 0600); err != nil {
		t.Fatal(err)
	}

	if err := Init(); err != nil {
		t.Fatalf("Init returned error when no known_hosts: %v", err)
	}

	dst := filepath.Join(sshDir, "known_hosts")
	if _, err := os.Stat(dst); err == nil {
		t.Error("known_hosts should not have been created when source does not exist")
	}
}
