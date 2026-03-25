package agent

import (
	"os"
	"path/filepath"
	"strings"
)

func init() {
	RegisterInstall(&AgentInstall{
		Name: "gemini",
		DockerSnippet: `# Install Gemini CLI
RUN bash -c 'export PATH="$HOME/.local/share/fnm:$PATH" && eval "$(fnm env)" && npm install -g @google/gemini-cli'
`,
		ProfileDeps: []string{"node"},
		BannerLine: `    echo "Gemini:    $(gemini --version 2>/dev/null || echo 'not found')"
`,
	})
}

type Gemini struct{}

func (Gemini) Name() string             { return "gemini" }
func (Gemini) Binary() string           { return "gemini" }
func (Gemini) NativeConfigDir() string  { return "~/.gemini" }
func (Gemini) ContainerConfigDir() string { return "/home/claude/.gemini" }
func (Gemini) AsylumConfigDir() string  { return "~/.asylum/agents/gemini" }

func (Gemini) EnvVars() map[string]string { return nil }

func (g Gemini) HasSession(projectPath string) bool {
	configDir, err := resolveConfigDir(g)
	if err != nil {
		return false
	}
	tmpDir := filepath.Join(configDir, "tmp")
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		root, err := os.ReadFile(filepath.Join(tmpDir, e.Name(), ".project_root"))
		if err != nil {
			continue
		}
		if strings.TrimSpace(string(root)) != projectPath {
			continue
		}
		chats, err := os.ReadDir(filepath.Join(tmpDir, e.Name(), "chats"))
		if err != nil {
			continue
		}
		if len(chats) > 0 {
			return true
		}
	}
	return false
}

func (Gemini) Command(resume bool, extraArgs []string) []string {
	parts := []string{"gemini", "--yolo"}
	if resume {
		parts = append(parts, "--resume")
	}
	parts = append(parts, quoteArgs(extraArgs)...)
	return wrapZsh(strings.Join(parts, " "))
}
