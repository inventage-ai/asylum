package kit

func init() {
	Register(&Kit{
		Name:        "openspec",
		Description: "OpenSpec CLI",
		Deps:        []string{"node"},
		Tools:       []string{"openspec"},
		Tier: TierOptIn,
		ConfigSnippet: `  # openspec:           # OpenSpec CLI
`,
		ConfigNodes:   configNodes("openspec", "OpenSpec CLI", nil),
		ConfigComment: "openspec:             # OpenSpec CLI",
		DockerSnippet: `# Install OpenSpec CLI
RUN bash -c 'export PATH="$HOME/.local/share/fnm:$PATH" && eval "$(fnm env)" && npm install -g @fission-ai/openspec@latest'
ENV OPENSPEC_TELEMETRY=0
`,
	})
}
