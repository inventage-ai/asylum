package kit

func init() {
	Register(&Kit{
		Name:        "openspec",
		Description: "OpenSpec CLI",
		DefaultOn:   true,
		Deps:        []string{"node"},
		Tools:       []string{"openspec"},
		DockerSnippet: `# Install OpenSpec CLI
RUN bash -c 'export PATH="$HOME/.local/share/fnm:$PATH" && eval "$(fnm env)" && npm install -g @fission-ai/openspec@latest'
`,
	})
}
