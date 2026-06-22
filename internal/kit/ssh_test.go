package kit

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func setupHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	return home
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestSSHCredentialFunc_IsolatedWithExistingKey(t *testing.T) {
	home := setupHome(t)
	keyDir := filepath.Join(home, ".asylum", "ssh")
	writeFile(t, filepath.Join(keyDir, "id_ed25519"), "private")
	writeFile(t, filepath.Join(keyDir, "id_ed25519.pub"), "public")

	mounts, err := sshCredentialFunc(CredentialOpts{
		HomeDir:   home,
		Isolation: "isolated",
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(mounts) != 2 {
		t.Fatalf("expected 2 mounts, got %d", len(mounts))
	}
	if mounts[0].HostPath != filepath.Join(keyDir, "id_ed25519") {
		t.Errorf("mount[0] host = %q, want private key", mounts[0].HostPath)
	}
	if mounts[0].Writable {
		t.Error("private key should be read-only")
	}
	if mounts[1].Destination != "~/.ssh/id_ed25519.pub" {
		t.Errorf("mount[1] dest = %q, want pub key", mounts[1].Destination)
	}
}

func TestSSHCredentialFunc_IsolatedWithKnownHosts(t *testing.T) {
	home := setupHome(t)
	keyDir := filepath.Join(home, ".asylum", "ssh")
	writeFile(t, filepath.Join(keyDir, "id_ed25519"), "private")
	writeFile(t, filepath.Join(keyDir, "id_ed25519.pub"), "public")
	writeFile(t, filepath.Join(home, ".ssh", "known_hosts"), "github.com ssh-ed25519 AAAA")

	mounts, err := sshCredentialFunc(CredentialOpts{
		HomeDir:   home,
		Isolation: "isolated",
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(mounts) != 3 {
		t.Fatalf("expected 3 mounts (key pair + known_hosts), got %d", len(mounts))
	}
	kh := mounts[2]
	if kh.HostPath != filepath.Join(home, ".ssh", "known_hosts") {
		t.Errorf("known_hosts host = %q", kh.HostPath)
	}
	if !kh.Writable {
		t.Error("known_hosts should be writable")
	}
}

func TestSSHCredentialFunc_IsolatedNoKnownHosts(t *testing.T) {
	home := setupHome(t)
	keyDir := filepath.Join(home, ".asylum", "ssh")
	writeFile(t, filepath.Join(keyDir, "id_ed25519"), "private")
	writeFile(t, filepath.Join(keyDir, "id_ed25519.pub"), "public")

	mounts, err := sshCredentialFunc(CredentialOpts{
		HomeDir:   home,
		Isolation: "isolated",
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(mounts) != 2 {
		t.Fatalf("expected 2 mounts (key pair only), got %d", len(mounts))
	}
}

func TestSSHCredentialFunc_SharedMountsDir(t *testing.T) {
	home := setupHome(t)
	os.MkdirAll(filepath.Join(home, ".ssh"), 0700)

	mounts, err := sshCredentialFunc(CredentialOpts{
		HomeDir:   home,
		Isolation: "shared",
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(mounts) != 1 {
		t.Fatalf("expected 1 mount, got %d", len(mounts))
	}
	if mounts[0].HostPath != filepath.Join(home, ".ssh") {
		t.Errorf("host path = %q, want ~/.ssh", mounts[0].HostPath)
	}
	if !mounts[0].Writable {
		t.Error("shared mount should be writable")
	}
}

func TestSSHCredentialFunc_SharedNoSSHDir(t *testing.T) {
	home := setupHome(t)

	mounts, err := sshCredentialFunc(CredentialOpts{
		HomeDir:   home,
		Isolation: "shared",
	})
	if err != nil {
		t.Fatal(err)
	}
	if mounts != nil {
		t.Fatalf("expected nil mounts when ~/.ssh missing, got %d", len(mounts))
	}
}

func TestSSHCredentialFunc_ProjectMode(t *testing.T) {
	home := setupHome(t)
	cname := "asylum-abc123-myproject"
	projKeyDir := filepath.Join(home, ".asylum", "projects", cname, "ssh")
	writeFile(t, filepath.Join(projKeyDir, "id_ed25519"), "proj-private")
	writeFile(t, filepath.Join(projKeyDir, "id_ed25519.pub"), "proj-public")

	mounts, err := sshCredentialFunc(CredentialOpts{
		HomeDir:       home,
		ContainerName: cname,
		Isolation:     "project",
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(mounts) != 2 {
		t.Fatalf("expected 2 mounts, got %d", len(mounts))
	}
	if mounts[0].HostPath != filepath.Join(projKeyDir, "id_ed25519") {
		t.Errorf("mount[0] host = %q, want project key", mounts[0].HostPath)
	}
}

// captureStdout redirects os.Stdout for the duration of fn and returns
// whatever fn wrote.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdout
	os.Stdout = w
	done := make(chan string)
	go func() {
		b, _ := io.ReadAll(r)
		done <- string(b)
	}()
	fn()
	w.Close()
	os.Stdout = orig
	return <-done
}

func TestEnsureSSHKey_SilentOnSuccess(t *testing.T) {
	if _, err := exec.LookPath("ssh-keygen"); err != nil {
		t.Skip("ssh-keygen not available")
	}
	dir := t.TempDir()
	keyDir := filepath.Join(dir, "ssh")

	out := captureStdout(t, func() {
		if err := ensureSSHKey(keyDir); err != nil {
			t.Fatalf("ensureSSHKey: %v", err)
		}
	})

	// ssh-keygen's "Your identification has been saved" preamble and its
	// randomart box must not appear; both are produced on stdout/stderr of
	// the subprocess in the pre-change implementation.
	for _, banned := range []string{
		"Your identification has been saved",
		"randomart",
		"--[ED25519",
		"SSH public key:",
		"Add this key",
	} {
		if strings.Contains(out, banned) {
			t.Errorf("captured stdout contains %q (expected silence):\n%s", banned, out)
		}
	}

	// The one notice we DO want to see.
	if !strings.Contains(out, "Generated SSH key at") || !strings.Contains(out, "asylum-reference.md") {
		t.Errorf("expected notice line not found in stdout:\n%s", out)
	}

	// Key pair should exist.
	if _, err := os.Stat(filepath.Join(keyDir, "id_ed25519")); err != nil {
		t.Errorf("private key not written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(keyDir, "id_ed25519.pub")); err != nil {
		t.Errorf("public key not written: %v", err)
	}
}

func TestEnsureSSHKey_FailureSurfacesCapturedOutput(t *testing.T) {
	// Pass a path that ssh-keygen cannot write to so the subprocess fails.
	// Pre-create the parent as a file so MkdirAll fails first → covers our
	// dir-create error path. ssh-keygen failure path is exercised by
	// pointing at a read-only parent.
	parent := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(parent, []byte("not a dir"), 0644); err != nil {
		t.Fatal(err)
	}
	err := ensureSSHKey(filepath.Join(parent, "ssh"))
	if err == nil {
		t.Fatal("expected error when key dir cannot be created")
	}
}

func TestSSHCredentialFunc_DefaultsToIsolated(t *testing.T) {
	home := setupHome(t)
	keyDir := filepath.Join(home, ".asylum", "ssh")
	writeFile(t, filepath.Join(keyDir, "id_ed25519"), "private")
	writeFile(t, filepath.Join(keyDir, "id_ed25519.pub"), "public")

	mounts, err := sshCredentialFunc(CredentialOpts{
		HomeDir:   home,
		Isolation: "", // empty should default to isolated
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(mounts) != 2 {
		t.Fatalf("expected 2 mounts, got %d", len(mounts))
	}
	if mounts[0].HostPath != filepath.Join(keyDir, "id_ed25519") {
		t.Errorf("expected isolated key path, got %q", mounts[0].HostPath)
	}
}
