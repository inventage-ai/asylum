package kit

import (
	"os/exec"
	"strings"
	"testing"
)

func TestGithubCredentialFunc_NoGh(t *testing.T) {
	// If gh isn't authenticated (or not installed), should return empty
	// We can't easily mock exec.Command, so just verify the function
	// doesn't error when gh auth token fails
	mounts, err := githubCredentialFunc(CredentialOpts{HomeDir: t.TempDir()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check if gh is actually authenticated
	out, ghErr := exec.Command("gh", "auth", "token").Output()
	if ghErr != nil || strings.TrimSpace(string(out)) == "" {
		// Not authenticated — should return empty
		if len(mounts) != 0 {
			t.Fatalf("expected 0 mounts when gh not authenticated, got %d", len(mounts))
		}
	} else {
		// Authenticated — should return a hosts.yml mount
		if len(mounts) != 1 {
			t.Fatalf("expected 1 mount when gh authenticated, got %d", len(mounts))
		}
		if mounts[0].Destination != "~/.config/gh/hosts.yml" {
			t.Errorf("Destination = %q, want %q", mounts[0].Destination, "~/.config/gh/hosts.yml")
		}
		if !strings.Contains(string(mounts[0].Content), "oauth_token") {
			t.Error("expected hosts.yml to contain oauth_token")
		}
	}
}
