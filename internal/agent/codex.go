package agent

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/inventage-ai/asylum/internal/log"
	"github.com/inventage-ai/asylum/internal/term"
)

func init() {
	agents["codex"] = Codex{}
	RegisterInstall(&AgentInstall{
		Name:           "codex",
		DockerPriority: 6,
		DockerSnippet: `# Install Codex
RUN bash -c 'export PATH="$HOME/.local/share/fnm:$PATH" && eval "$(fnm env)" && npm install -g @openai/codex'
`,
		KitDeps: []string{"node"},
		BannerLine: `    echo "Codex:     $(codex --version 2>/dev/null || echo 'not found')"
`,
	})
}

type Codex struct{}

func (Codex) Name() string               { return "codex" }
func (Codex) Binary() string             { return "codex" }
func (Codex) NativeConfigDir() string    { return "~/.codex" }
func (Codex) ContainerConfigDir() string { return "~/.codex" }
func (Codex) AsylumConfigDir() string    { return "~/.asylum/agents/codex" }

func (Codex) EnvVars() map[string]string {
	home, _ := os.UserHomeDir()
	return map[string]string{
		"CODEX_HOME": filepath.Join(home, ".codex"),
	}
}

func (Codex) markerDir(configDir, projectPath string) string {
	encoded := strings.ReplaceAll(projectPath, "/", "-")
	return filepath.Join(configDir, "projects", encoded)
}

func (c Codex) HasSession(configDir, projectPath string) bool {
	_, err := os.Stat(filepath.Join(c.markerDir(configDir, projectPath), ".has_session"))
	return err == nil
}

func (c Codex) WriteMarker(configDir, projectPath string) error {
	dir := c.markerDir(configDir, projectPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, ".has_session"), nil, 0644)
}

func (Codex) Command(resume bool, extraArgs []string, _ CmdOpts) []string {
	if resume {
		if len(extraArgs) == 0 {
			return wrapZsh("codex resume --last --yolo")
		}
		log.Warn("codex: resume skipped because extra args were provided")
	}
	parts := []string{"codex", "--yolo"}
	parts = append(parts, term.ShellQuoteArgs(extraArgs)...)
	return wrapZsh(strings.Join(parts, " "))
}
