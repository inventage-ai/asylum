package agent

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/inventage-ai/asylum/internal/term"
)

func init() {
	agents["copilot"] = Copilot{}
	RegisterInstall(&AgentInstall{
		Name: "copilot",
		DockerSnippet: `# Install GitHub Copilot CLI
RUN curl -fsSL https://gh.io/copilot-install | bash && \
    ~/.local/bin/copilot --version
`,
		BannerLine: `    echo "Copilot:   $(copilot --version 2>/dev/null || echo 'not found')"
`,
	})
}

// Copilot is the GitHub Copilot CLI adapter.
type Copilot struct{}

func (Copilot) Name() string               { return "copilot" }
func (Copilot) Binary() string             { return "copilot" }
func (Copilot) NativeConfigDir() string    { return "~/.copilot" }
func (Copilot) ContainerConfigDir() string { return "~/.copilot" }
func (Copilot) AsylumConfigDir() string    { return "~/.asylum/agents/copilot" }

// EnvVars sets COPILOT_HOME so the CLI uses the mounted config dir, which is
// where session-state/ and mcp-config.json live.
func (Copilot) EnvVars() map[string]string {
	home, _ := os.UserHomeDir()
	return map[string]string{
		"COPILOT_HOME": filepath.Join(home, ".copilot"),
	}
}

// markerDir is the Asylum-owned per-project directory used to track whether
// copilot has been launched in this project before. It lives alongside
// copilot's own state so it travels with the same config-dir isolation.
func (Copilot) markerDir(configDir, projectPath string) string {
	encoded := strings.ReplaceAll(projectPath, "/", "-")
	return filepath.Join(configDir, "asylum-projects", encoded)
}

// HasSession reports whether copilot has been launched in this project before.
// Copilot's own session-state directory is global to the config dir and the
// `--resume` picker lists every recent session regardless of project, so
// auto-passing `--resume` based on session-state would expose unrelated
// projects' context. Instead we require an Asylum-owned per-project marker.
func (c Copilot) HasSession(configDir, projectPath string) bool {
	_, err := os.Stat(filepath.Join(c.markerDir(configDir, projectPath), ".has_session"))
	return err == nil
}

// WriteMarker records that copilot has been launched in this project, so
// future runs in the same project may pass `--resume`. Called by the
// container launcher after a successful agent-mode startup.
func (c Copilot) WriteMarker(configDir, projectPath string) error {
	dir := c.markerDir(configDir, projectPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, ".has_session"), nil, 0644)
}

// Command launches copilot. When `gh` is available in the container (github
// kit active), export GH_TOKEN from the host-mounted gh credential so copilot
// authenticates non-interactively (per the COPILOT_GITHUB_TOKEN → GH_TOKEN →
// GITHUB_TOKEN precedence). If gh is absent or not authenticated, copilot
// falls back to its own interactive flow.
func (Copilot) Command(resume bool, extraArgs []string, _ CmdOpts) []string {
	parts := []string{"copilot"}
	if resume {
		parts = append(parts, "--resume")
	}
	parts = append(parts, term.ShellQuoteArgs(extraArgs)...)
	cmd := strings.Join(parts, " ")

	preamble := `if command -v gh >/dev/null 2>&1 && [ -z "${GH_TOKEN:-}" ]; then export GH_TOKEN="$(gh auth token 2>/dev/null || true)"; fi`
	return []string{"zsh", "-c", "source ~/.zshrc && " + preamble + " && exec " + cmd}
}
