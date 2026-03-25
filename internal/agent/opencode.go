package agent

import (
	"strings"

	"github.com/inventage-ai/asylum/internal/log"
	"github.com/inventage-ai/asylum/internal/term"
)

func init() {
	agents["opencode"] = Opencode{}
	RegisterInstall(&AgentInstall{
		Name: "opencode",
		DockerSnippet: `# Install Opencode
RUN curl -fsSL https://opencode.ai/install | bash
`,
		BannerLine: `    echo "Opencode:  $(opencode --version 2>/dev/null || echo 'not found')"
`,
	})
}

type Opencode struct{}

func (Opencode) Name() string               { return "opencode" }
func (Opencode) Binary() string             { return "opencode" }
func (Opencode) NativeConfigDir() string    { return "~/.opencode" }
func (Opencode) ContainerConfigDir() string { return "/home/claude/.opencode" }
func (Opencode) AsylumConfigDir() string    { return "~/.asylum/agents/opencode" }

func (Opencode) EnvVars() map[string]string { return nil }

func (Opencode) HasSession(_ string) bool { return false }

func (Opencode) Command(resume bool, extraArgs []string) []string {
	if resume {
		log.Warn("opencode: resume not supported, starting fresh session")
	}
	parts := []string{"opencode"}
	parts = append(parts, term.ShellQuoteArgs(extraArgs)...)
	return wrapZsh(strings.Join(parts, " "))
}
