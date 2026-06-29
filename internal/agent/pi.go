package agent

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/inventage-ai/asylum/internal/term"
)

func init() {
	agents["pi"] = Pi{}
	RegisterInstall(&AgentInstall{
		Name:           "pi",
		DockerPriority: 6,
		DockerSnippet: `# Install Pi
RUN bash -c 'export PATH="$HOME/.local/share/fnm:$PATH" && eval "$(fnm env)" && npm install -g @earendil-works/pi-coding-agent'
`,
		KitDeps: []string{"node"},
		BannerLine: `    echo "Pi:        $(pi --version 2>/dev/null || echo 'not found')"
`,
	})
}

type Pi struct{}

func (Pi) Name() string               { return "pi" }
func (Pi) Binary() string             { return "pi" }
func (Pi) NativeConfigDir() string    { return "~/.pi" }
func (Pi) ContainerConfigDir() string { return "~/.pi" }
func (Pi) AsylumConfigDir() string    { return "~/.asylum/agents/pi" }

func (Pi) EnvVars() map[string]string { return nil }

func (Pi) HasSession(configDir, projectPath string) bool {
	encoded := "--" + strings.ReplaceAll(projectPath, "/", "-") + "--"
	sessDir := filepath.Join(configDir, "agent", "sessions", encoded)
	_, err := os.Stat(sessDir)
	return err == nil
}

func (Pi) Command(resume bool, extraArgs []string, _ CmdOpts) []string {
	parts := []string{"pi"}
	if resume {
		parts = append(parts, "--continue")
	}
	parts = append(parts, term.ShellQuoteArgs(extraArgs)...)
	return wrapZsh(strings.Join(parts, " "))
}
