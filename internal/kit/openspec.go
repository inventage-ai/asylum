package kit

func init() {
	Register(&Kit{
		Name:        "openspec",
		Description: "OpenSpec CLI",
		Deps:        []string{"node"},
		Tools:       []string{"openspec"},
		ConfigSnippet: `  openspec:            # OpenSpec CLI (requires node)
`,
		DockerSnippet: `# Install OpenSpec CLI
RUN bash -c 'export PATH="$HOME/.local/share/fnm:$PATH" && eval "$(fnm env)" && npm install -g @fission-ai/openspec@latest'
ENV OPENSPEC_TELEMETRY=0
`,
	})
}
