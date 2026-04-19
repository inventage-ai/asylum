package agent

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/inventage-ai/asylum/internal/term"
)

func init() {
	agents["claude"] = Claude{}
	RegisterInstall(&AgentInstall{
		Name:           "claude",
		DockerPriority: 5,
		DockerSnippet: `# Install Claude Code
RUN curl -fsSL https://claude.ai/install.sh | bash && \
    ~/.local/bin/claude --version
`,
		BannerLine: `    echo "Claude:    $(claude --version 2>/dev/null || echo 'not found')"
`,
	})
}

type Claude struct{}

func (Claude) Name() string               { return "claude" }
func (Claude) Binary() string             { return "claude" }
func (Claude) NativeConfigDir() string    { return "~/.claude" }
func (Claude) ContainerConfigDir() string { return "~/.claude" }
func (Claude) AsylumConfigDir() string    { return "~/.asylum/agents/claude" }

func (Claude) EnvVars() map[string]string {
	home, _ := os.UserHomeDir()
	return map[string]string{
		"CLAUDE_CONFIG_DIR": filepath.Join(home, ".claude"),
	}
}

func (Claude) HasSession(configDir, projectPath string) bool {
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

func (Claude) Command(resume bool, extraArgs []string, opts CmdOpts) []string {
	parts := []string{"claude", "--dangerously-skip-permissions"}
	if resume {
		parts = append(parts, "--continue")
	}
	if opts.KitSkillsDir != "" {
		parts = append(parts, "--add-dir", opts.KitSkillsDir)
	}
	parts = append(parts, term.ShellQuoteArgs(extraArgs)...)
	return wrapZsh(strings.Join(parts, " "))
}
