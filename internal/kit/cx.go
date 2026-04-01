package kit

func init() {
	Register(&Kit{
		Name:           "cx",
		Description:    "Semantic code navigation for AI agents",
		DockerPriority: 42,
		Tier:           TierOptIn,
		Tools:       []string{"cx"},
		NeedsMount:  true,
		ConfigSnippet: `  # cx:                 # Semantic code navigation
  #   packages:        # tree-sitter language grammars
  #     - python
  #     - typescript
  #     - go
`,
		ConfigNodes:   configNodes("cx", "Semantic code navigation", nil),
		ConfigComment: "packages:        # tree-sitter language grammars\n    #   - python\n    #   - typescript\n    #   - go",
		DockerSnippet: `# Install cx
RUN curl -sL https://raw.githubusercontent.com/ind-igo/cx/master/install.sh | sh
RUN mkdir -p /tmp/asylum-kit-rules && cx skill > /tmp/asylum-kit-rules/cx.md || true
`,
		EntrypointSnippet: `# Mount cx rules file into agent rules directory
if [ -s /tmp/asylum-kit-rules/cx.md ] && [ -d "$HOME/.claude/rules" ]; then
    touch "$HOME/.claude/rules/cx.md"
    sudo mount --bind /tmp/asylum-kit-rules/cx.md "$HOME/.claude/rules/cx.md"
fi
`,
		RulesSnippet: `### cx (cx kit)
Semantic code navigation. If cx reports a missing grammar, add it to ` + "`.asylum`" + ` under ` + "`kits.cx.packages`" + ` so it persists across container rebuilds. ` + "`cx lang add <lang>`" + ` only lasts for the current session.
`,
		BannerLines: `    echo "cx:        $(cx --version 2>/dev/null || echo 'not found')"
`,
	})
}
