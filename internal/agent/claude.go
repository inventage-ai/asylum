package agent

import (
	"os"
	"path/filepath"
	"strings"
)

func init() {
	RegisterInstall(&AgentInstall{
		Name: "claude",
		DockerSnippet: `# Install Claude Code
RUN curl -fsSL https://claude.ai/install.sh | bash && \
    ~/.local/bin/claude --version
`,
		BannerLine: `    echo "Claude:    $(claude --version 2>/dev/null || echo 'not found')"
`,
	})
}

type Claude struct{}

func (Claude) Name() string             { return "claude" }
func (Claude) Binary() string           { return "claude" }
func (Claude) NativeConfigDir() string  { return "~/.claude" }
func (Claude) ContainerConfigDir() string { return "/home/claude/.claude" }
func (Claude) AsylumConfigDir() string  { return "~/.asylum/agents/claude" }

func (Claude) EnvVars() map[string]string {
	return map[string]string{
		"CLAUDE_CONFIG_DIR": "/home/claude/.claude",
	}
}

func (c Claude) HasSession(projectPath string) bool {
	configDir, err := resolveConfigDir(c)
	if err != nil {
		return false
	}
	// Claude encodes project paths by replacing "/" with "-"
	encoded := strings.ReplaceAll(projectPath, "/", "-")
	projDir := filepath.Join(configDir, "projects", encoded)
	entries, err := os.ReadDir(projDir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".jsonl") {
			return true
		}
	}
	return false
}

func (Claude) Command(resume bool, extraArgs []string) []string {
	parts := []string{"claude", "--dangerously-skip-permissions"}
	if resume {
		parts = append(parts, "--continue")
	}
	parts = append(parts, quoteArgs(extraArgs)...)
	return wrapZsh(strings.Join(parts, " "))
}
