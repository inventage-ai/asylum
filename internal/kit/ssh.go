package kit

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/inventage-ai/asylum/internal/log"
	"gopkg.in/yaml.v3"
)

func init() {
	Register(&Kit{
		Name:            "ssh",
		Description:     "SSH key management",
		Tier:            TierAlwaysOn,
		MountFunc: sshCredentialFunc,
		ConfigComment:   "ssh:                  # SSH key management\n  isolation: isolated # shared, isolated, project",
		ConfigNodes: configNodes("ssh", "SSH key management", []*yaml.Node{
			ScalarNode("isolation", "shared, isolated, project"),
			ScalarNode("isolated", ""),
		}),
	})
}

func sshCredentialFunc(opts CredentialOpts) ([]CredentialMount, error) {
	isolation := opts.Isolation
	if isolation == "" {
		isolation = "isolated"
	}

	if isolation == "shared" {
		sshDir := filepath.Join(opts.HomeDir, ".ssh")
		if !dirExists(sshDir) {
			return nil, nil
		}
		return []CredentialMount{
			{HostPath: sshDir, Destination: "~/.ssh", Writable: true},
		}, nil
	}

	// isolated or project mode
	var keyDir string
	switch isolation {
	case "project":
		keyDir = filepath.Join(opts.HomeDir, ".asylum", "projects", opts.ContainerName, "ssh")
	default:
		keyDir = filepath.Join(opts.HomeDir, ".asylum", "ssh")
	}

	if err := ensureSSHKey(keyDir); err != nil {
		return nil, err
	}

	mounts := []CredentialMount{
		{HostPath: filepath.Join(keyDir, "id_ed25519"), Destination: "~/.ssh/id_ed25519"},
		{HostPath: filepath.Join(keyDir, "id_ed25519.pub"), Destination: "~/.ssh/id_ed25519.pub"},
	}

	knownHosts := filepath.Join(opts.HomeDir, ".ssh", "known_hosts")
	if fileExists(knownHosts) {
		mounts = append(mounts, CredentialMount{
			HostPath:    knownHosts,
			Destination: "~/.ssh/known_hosts",
			Writable:    true,
		})
	}

	return mounts, nil
}

func ensureSSHKey(dir string) error {
	keyPath := filepath.Join(dir, "id_ed25519")
	if fileExists(keyPath) {
		return nil
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create ssh dir: %w", err)
	}

	hostname, _ := os.Hostname()
	comment := fmt.Sprintf("asylum@%s", hostname)

	var captured bytes.Buffer
	cmd := exec.Command("ssh-keygen", "-t", "ed25519", "-f", keyPath, "-C", comment, "-N", "")
	cmd.Stdout = &captured
	cmd.Stderr = &captured
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ssh-keygen: %w\n%s", err, captured.String())
	}

	log.Info("Generated SSH key at %s.pub — see asylum-reference.md for usage.", keyPath)
	return nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
