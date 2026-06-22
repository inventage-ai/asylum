package kit

func init() {
	Register(&Kit{
		Name:           "openspec",
		Description:    "OpenSpec CLI",
		DockerPriority: 46,
		Deps:           []string{"node"},
		Tools:       []string{"openspec"},
		Tier: TierDefault,
		ConfigSnippet: `  openspec:             # OpenSpec CLI
`,
		ConfigNodes:   configNodes("openspec", "OpenSpec CLI", nil),
		ConfigComment: "openspec:             # OpenSpec CLI",
		DockerSnippet: `# Install OpenSpec CLI
RUN bash -c 'export PATH="$HOME/.local/share/fnm:$PATH" && eval "$(fnm env)" && npm install -g @fission-ai/openspec@latest'
ENV OPENSPEC_TELEMETRY=0
# Seed preferred OpenSpec workflow profile (custom profile with verify instead of sync)
RUN mkdir -p "$HOME/.config/openspec" && \
    printf '%s' '{"profile":"custom","workflows":["propose","explore","apply","verify","archive"]}' > "$HOME/.config/openspec/config.json"
# Install asylum-openspec-init: non-interactive setup using the active agent and preferred profile
RUN printf '%s\n' \
    '#!/usr/bin/env bash' \
    'set -euo pipefail' \
    'tools="${ASYLUM_AGENT:-claude}"' \
    '[ "$tools" = copilot ] && tools=github-copilot' \
    'if [ -d openspec ]; then' \
    '  openspec update --force' \
    'else' \
    '  openspec init --tools "$tools" --force' \
    'fi' \
    | sudo tee /usr/local/bin/asylum-openspec-init >/dev/null && \
    sudo chmod +x /usr/local/bin/asylum-openspec-init
`,
		RulesSnippet: `### OpenSpec (openspec kit)
The ` + "`openspec`" + ` CLI is installed for structured, spec-driven change management. If the user wants to use OpenSpec in a project where it is not yet set up (no ` + "`openspec/`" + ` directory), run ` + "`asylum-openspec-init`" + ` — it initializes OpenSpec non-interactively with the preferred profile and the agent's toolset. It is safe to re-run (refreshes an existing setup).
`,
	})
}
