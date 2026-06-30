package agent

import (
	"strings"
	"testing"
)

func TestVersionedSnippet(t *testing.T) {
	tests := []struct {
		agentName  string
		snippet    string
		versions   map[string]string
		wantLine   string // expected first line
	}{
		{
			agentName: "gemini",
			snippet:   "RUN npm install -g @google/gemini-cli\n",
			versions:  map[string]string{"gemini": "0.8.0"},
			wantLine:  "ARG GEMINI_VERSION=0.8.0",
		},
		{
			agentName: "gemini",
			snippet:   "RUN npm install -g @google/gemini-cli\n",
			versions:  map[string]string{},
			wantLine:  "RUN npm install -g @google/gemini-cli",
		},
		{
			agentName: "claude",
			snippet:   "RUN curl -fsSL https://claude.ai/install.sh | bash && \\\n    ~/.local/bin/claude --version\n",
			versions:  map[string]string{"claude": "v2.1.195"},
			wantLine:  "ARG CLAUDE_VERSION=v2.1.195",
		},
		{
			agentName: "copilot",
			snippet:   "RUN curl -fsSL https://gh.io/copilot-install | bash && \\\n    ~/.local/bin/copilot --version\n",
			versions:  map[string]string{"copilot": "v1.0.65"},
			wantLine:  "ARG COPILOT_VERSION=v1.0.65",
		},
		{
			agentName: "opencode",
			snippet:   "RUN curl -fsSL https://opencode.ai/install | bash\n",
			versions:  map[string]string{"opencode": "v0.0.55"},
			wantLine:  "ARG OPENCODE_VERSION=v0.0.55",
		},
		{
			agentName: "pi",
			snippet:   "RUN bash -c 'export PATH=\"$HOME/.local/share/fnm:$PATH\" && eval \"$(fnm env)\" && npm install -g @earendil-works/pi-coding-agent'\n",
			versions:  map[string]string{"pi": "0.13.0"},
			wantLine:  "ARG PI_VERSION=0.13.0",
		},
		{
			agentName: "echo",
			snippet:   "RUN echo hello\n",
			versions:  map[string]string{"echo": "1.0.0"},
			wantLine:  "RUN echo hello",
		},
		{
			agentName: "echo",
			snippet:   "RUN echo hello\n",
			versions:  map[string]string{},
			wantLine:  "RUN echo hello",
		},
	}
	for _, tt := range tests {
		t.Run(tt.agentName, func(t *testing.T) {
			install := &AgentInstall{
				Name:        tt.agentName,
				DockerSnippet: tt.snippet,
			}
			result := install.VersionedSnippet(tt.versions)

			// Check the first line
			lines := strings.Split(strings.TrimSpace(result), "\n")
			if len(lines) == 0 {
				t.Fatalf("VersionedSnippet() returned empty string")
			}

			if _, hasVersion := tt.versions[tt.agentName]; !hasVersion {
				// No version available — should fall through to original snippet
				if !strings.Contains(result, tt.wantLine) {
					t.Errorf("VersionedSnippet() = %q, want to contain %q", result, tt.wantLine)
				}
				return
			}

			if lines[0] != tt.wantLine {
				t.Errorf("VersionedSnippet() first line = %q, want %q", lines[0], tt.wantLine)
			}
		})
	}
}

func TestVersionArgName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"gemini", "GEMINI_VERSION"},
		{"codex", "CODEX_VERSION"},
		{"pi", "PI_VERSION"},
		{"claude", "CLAUDE_VERSION"},
		{"opencode", "OPENCODE_VERSION"},
	}
	for _, tt := range tests {
		if got := versionArgName(tt.name); got != tt.want {
			t.Errorf("versionArgName(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}
