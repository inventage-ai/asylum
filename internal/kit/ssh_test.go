package kit

import (
	"os"
	"path/filepath"
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
