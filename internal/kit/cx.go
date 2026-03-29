package kit

func init() {
	Register(&Kit{
		Name:        "cx",
		Description: "Semantic code navigation for AI agents",
		Tier:        TierOptIn,
		Tools:       []string{"cx"},
		ConfigSnippet: `  # cx:                # Semantic code navigation
    # packages:        # tree-sitter language grammars
    #   - python
    #   - typescript
    #   - go
`,
		ConfigNodes:   configNodes("cx", "Semantic code navigation", nil),
		ConfigComment: "packages:        # tree-sitter language grammars\n    #   - python\n    #   - typescript\n    #   - go",
		DockerSnippet: `# Install cx
RUN curl -sL https://raw.githubusercontent.com/ind-igo/cx/master/install.sh | sh
`,
		RulesSnippet: `### cx (cx kit)
cx is installed for semantic code navigation. Commands: ` + "`cx overview <file>`" + ` for file table of contents, ` + "`cx symbols`" + ` to search symbols, ` + "`cx definition --name <sym>`" + ` for function bodies, ` + "`cx references --name <sym>`" + ` to find usages. Use ` + "`cx lang add <lang>`" + ` to add language support.
`,
		BannerLines: `    echo "cx:        $(cx --version 2>/dev/null || echo 'not found')"
`,
	})
}
