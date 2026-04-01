package kit

func init() {
	Register(&Kit{
		Name:           "agent-browser",
		Description:    "Browser automation via agent-browser",
		DockerPriority: 34,
		Tier:           TierOptIn,
		Deps:        []string{"node"},
		Tools:       []string{"agent-browser"},
		NeedsMount:  true,
		ConfigSnippet: `  # agent-browser:      # Browser automation via agent-browser
`,
		ConfigNodes:   configNodes("agent-browser", "Browser automation via agent-browser", nil),
		ConfigComment: "agent-browser:        # Browser automation via agent-browser",
		DockerSnippet: `# Install agent-browser and Chrome/Chromium
RUN bash -c 'export PATH="$HOME/.local/share/fnm:$PATH" && eval "$(fnm env)" && \
    npm install -g agent-browser && \
    if [ "$(uname -m)" = "aarch64" ]; then \
        sudo apt-get update && sudo apt-get install -y --no-install-recommends chromium && sudo rm -rf /var/lib/apt/lists/*; \
    else \
        sudo env "PATH=$PATH" agent-browser install --with-deps; \
    fi'
RUN bash -c 'export PATH="$HOME/.local/share/fnm:$PATH" && eval "$(fnm env)" && \
    cd /tmp && npx skills add vercel-labs/agent-browser --skill agent-browser --yes --copy && \
    mv .claude/skills/agent-browser /tmp/asylum-kit-skills-agent-browser' || true
`,
		EntrypointSnippet: `# On ARM64, point agent-browser at system Chromium
if [ "$(uname -m)" = "aarch64" ] && command -v chromium >/dev/null 2>&1; then
    export AGENT_BROWSER_EXECUTABLE_PATH=/usr/bin/chromium
fi
# Mount agent-browser skill into Claude skills directory
if [ -d /tmp/asylum-kit-skills-agent-browser ] && [ -d "$HOME/.claude" ]; then
    mkdir -p "$HOME/.claude/skills/agent-browser"
    sudo mount --bind /tmp/asylum-kit-skills-agent-browser "$HOME/.claude/skills/agent-browser"
fi
`,
		RulesSnippet: `### Browser (agent-browser kit)
Use ` + "`agent-browser`" + ` for web automation. Run ` + "`agent-browser --help`" + ` for all commands.

Core workflow:

1. ` + "`agent-browser open <url>`" + ` - Navigate to page
2. ` + "`agent-browser snapshot -i`" + ` - Get interactive elements with refs (@e1, @e2)
3. ` + "`agent-browser click @e1`" + ` / ` + "`fill @e2 \"text\"`" + ` - Interact using refs
4. Re-snapshot after page changes
`,
		BannerLines: `    echo "agent-browser: $(agent-browser --version 2>/dev/null || echo 'not found')"
`,
	})
}
