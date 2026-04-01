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
# Generate rtk hooks for Claude Code
RUN mkdir -p "$HOME/.claude" && \
    $HOME/.local/bin/rtk init -g && \
    mkdir -p /tmp/asylum-kit-rtk && \
    cp -r "$HOME/.claude/hooks" /tmp/asylum-kit-rtk/hooks && \
    cp "$HOME/.claude/RTK.md" /tmp/asylum-kit-rtk/RTK.md && \
    rm -rf "$HOME/.claude"
`,
		EntrypointSnippet: `# Mount rtk hooks into Claude config directory
if [ -d /tmp/asylum-kit-rtk ] && [ -d "$HOME/.claude" ]; then
    # Mount hook script
    if [ -d /tmp/asylum-kit-rtk/hooks ]; then
        mkdir -p "$HOME/.claude/hooks"
        for f in /tmp/asylum-kit-rtk/hooks/*; do
            [ -f "$f" ] && touch "$HOME/.claude/hooks/$(basename "$f")" && \
                sudo mount --bind "$f" "$HOME/.claude/hooks/$(basename "$f")"
        done
    fi

    # Mount RTK.md awareness doc
    if [ -f /tmp/asylum-kit-rtk/RTK.md ]; then
        touch "$HOME/.claude/RTK.md"
        sudo mount --bind /tmp/asylum-kit-rtk/RTK.md "$HOME/.claude/RTK.md"
    fi

    # Register PreToolUse hook in settings.json
    settings="$HOME/.claude/settings.json"
    hook_path="$HOME/.claude/hooks/rtk-rewrite.sh"
    if [ -f "$hook_path" ]; then
        if [ ! -f "$settings" ]; then
            echo '{}' > "$settings"
        fi
        jq --arg cmd "$hook_path" '
            .hooks.PreToolUse //= [] |
            if (.hooks.PreToolUse | map(select(.matcher == "Bash" and (.hooks // [] | any(.command | test("rtk"))))) | length) == 0 then
                .hooks.PreToolUse += [{"matcher": "Bash", "hooks": [{"type": "command", "command": $cmd}]}]
            else . end
        ' "$settings" > "${settings}.tmp" && mv "${settings}.tmp" "$settings"
    fi
fi
`,
		RulesSnippet: `### rtk (rtk kit)
rtk is installed for transparent token reduction. It intercepts shell commands and compresses output, reducing token usage by 60-90%. Use ` + "`rtk gain`" + ` to see token savings statistics and ` + "`rtk discover`" + ` to identify optimization opportunities.
`,
		BannerLines: `    echo "rtk:       $(rtk --version 2>/dev/null | head -1 || echo 'not found')"
`,
	})
}
