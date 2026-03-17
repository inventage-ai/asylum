package ssh

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/inventage-ai/asylum/internal/log"
)

func Init() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("home dir: %w", err)
	}

	sshDir := filepath.Join(home, ".asylum", "ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return fmt.Errorf("create ssh dir: %w", err)
	}

	// Copy known_hosts if it exists
	knownHosts := filepath.Join(home, ".ssh", "known_hosts")
	if info, err := os.Stat(knownHosts); err == nil && !info.IsDir() {
		data, err := os.ReadFile(knownHosts)
		if err == nil {
			dst := filepath.Join(sshDir, "known_hosts")
			if err := os.WriteFile(dst, data, 0600); err != nil {
				return fmt.Errorf("write known_hosts: %w", err)
			}
			log.Success("copied known_hosts")
		}
	}

	keyPath := filepath.Join(sshDir, "id_ed25519")
	if _, err := os.Stat(keyPath); err == nil {
		log.Info("SSH key already exists at %s", keyPath)
		log.Info("replace with your own keys if needed")
		return nil
	}

	hostname, _ := os.Hostname()
	comment := fmt.Sprintf("asylum@%s", hostname)

	cmd := exec.Command("ssh-keygen", "-t", "ed25519", "-f", keyPath, "-C", comment, "-N", "")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ssh-keygen: %w", err)
	}

	pubKey, err := os.ReadFile(keyPath + ".pub")
	if err != nil {
		return fmt.Errorf("read public key: %w", err)
	}

	log.Success("SSH key generated")
	fmt.Printf("\nPublic key:\n%s\n", pubKey)
	log.Info("add this key to your Git hosting provider")
	log.Info("or replace with your own keys at %s", sshDir)

	return nil
}
