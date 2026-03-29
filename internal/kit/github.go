package kit

import (
	"fmt"
	"os/exec"
	"strings"
)

func init() {
	Register(&Kit{
		Name:            "github",
		Description:     "GitHub CLI",
		Tools:           []string{"gh"},
		CredentialFunc:  githubCredentialFunc,
		CredentialLabel: "GitHub",
		ConfigSnippet: `  github:               # GitHub CLI (gh)
`,
		ConfigNodes:   configNodes("github", "GitHub CLI (gh)", nil),
		ConfigComment: "github:               # GitHub CLI (gh)",
		DockerSnippet: `# Install GitHub CLI
USER root
RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | \
    gpg --dearmor -o /usr/share/keyrings/githubcli-archive-keyring.gpg && \
    chmod 644 /usr/share/keyrings/githubcli-archive-keyring.gpg && \
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" \
    > /etc/apt/sources.list.d/github-cli.list && \
    apt-get update && \
    apt-get install -y gh && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*
USER ${USERNAME}
`,
	})
}

// githubCredentialFunc extracts the gh auth token from the host (which may
// be in the system keyring) and generates a hosts.yml for the container.
func githubCredentialFunc(opts CredentialOpts) ([]CredentialMount, error) {
	out, err := exec.Command("gh", "auth", "token").Output()
	if err != nil {
		return nil, nil // gh not installed or not authenticated
	}
	token := strings.TrimSpace(string(out))
	if token == "" {
		return nil, nil
	}

	hostsYAML := fmt.Sprintf("github.com:\n    oauth_token: %s\n    git_protocol: https\n", token)
	return []CredentialMount{
		{
			Content:     []byte(hostsYAML),
			FileName:    "hosts.yml",
			Destination: "~/.config/gh",
			Writable:    true, // gh writes config.yml and migrates hosts.yml
		},
	}, nil
}
