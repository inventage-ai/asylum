package kit

func init() {
	Register(&Kit{
		Name:        "rtk",
		Description: "Token-reduction proxy for LLM agents",
		Tier:        TierOptIn,
		Tools:       []string{"rtk"},
		NeedsMount:  true,
		ConfigSnippet: `  # rtk:                # Token-reduction proxy (rtk)
`,
		ConfigNodes:   configNodes("rtk", "Token-reduction proxy (rtk)", nil),
		ConfigComment: "rtk:                  # Token-reduction proxy (rtk)",
		DockerSnippet: `# Install rtk
RUN curl -fsSL https://raw.githubusercontent.com/rtk-ai/rtk/refs/heads/master/install.sh | sh
# Stage RTK.md awareness doc. rtk init -g writes it into ~/.claude/.
# Recent rtk no longer creates a hooks/ directory — it expects its hook to be
# registered as the command "rtk hook claude" in settings.json (handled below).
RUN mkdir -p "$HOME/.claude" && \
    $HOME/.local/bin/rtk init -g && \
    mkdir -p /tmp/asylum-kit-rtk && \
    cp "$HOME/.claude/RTK.md" /tmp/asylum-kit-rtk/RTK.md && \
    rm -rf "$HOME/.claude"
`,
		EntrypointSnippet: `# Mount RTK.md and register rtk's PreToolUse hook in settings.json
if [ -d /tmp/asylum-kit-rtk ] && [ -d "$HOME/.claude" ]; then
    if [ -f /tmp/asylum-kit-rtk/RTK.md ]; then
        touch "$HOME/.claude/RTK.md"
        sudo mount --bind /tmp/asylum-kit-rtk/RTK.md "$HOME/.claude/RTK.md"
    fi

    settings="$HOME/.claude/settings.json"
    if [ ! -f "$settings" ]; then
        echo '{}' > "$settings"
    fi
    # Drop any existing rtk PreToolUse Bash entry (upgrades stale file-path
    # hooks from older asylum versions) and append the canonical command.
    jq '
        .hooks.PreToolUse //= [] |
        .hooks.PreToolUse |= [.[] | select(.matcher != "Bash" or ((.hooks // []) | any(.command | test("rtk")) | not))] |
        .hooks.PreToolUse += [{"matcher": "Bash", "hooks": [{"type": "command", "command": "rtk hook claude"}]}]
    ' "$settings" > "${settings}.tmp" && mv "${settings}.tmp" "$settings"
fi
`,
		RulesSnippet: `### rtk (rtk kit)
rtk is installed for transparent token reduction. It intercepts shell commands and compresses output, reducing token usage by 60-90%. Use ` + "`rtk gain`" + ` to see token savings statistics and ` + "`rtk discover`" + ` to identify optimization opportunities.
`,
		BannerLines: `    echo "rtk:       $(rtk --version 2>/dev/null | head -1 || echo 'not found')"
`,
	})
}
