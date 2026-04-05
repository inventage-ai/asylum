package kit

func init() {
	Register(&Kit{
		Name:           "ast-grep",
		Description:    "AST-based code search, lint, and rewrite",
		DockerPriority: 44,
		Tier:           TierOptIn,
		Deps:        []string{"node"},
		Tools:       []string{"sg"},
		NeedsMount:  true,
		ConfigSnippet: `  # ast-grep:           # AST-based code search (sg)
`,
		ConfigNodes:   configNodes("ast-grep", "AST-based code search (sg)", nil),
		ConfigComment: "ast-grep:             # AST-based code search (sg)",
		DockerSnippet: `# Install ast-grep
RUN bash -c 'export PATH="$HOME/.local/share/fnm:$PATH" && eval "$(fnm env)" && \
    npm install -g @ast-grep/cli'
RUN bash -c 'export PATH="$HOME/.local/share/fnm:$PATH" && eval "$(fnm env)" && \
    cd /tmp && npx skills add ast-grep/agent-skill --skill ast-grep --yes --copy && \
    mv .claude/skills/ast-grep /tmp/asylum-kit-skills-ast-grep' || true
`,
		EntrypointSnippet: `# Mount ast-grep skill into Claude skills directory (skip if already present, e.g. shared config)
if [ -d /tmp/asylum-kit-skills-ast-grep ] && [ -d "$HOME/.claude" ] && [ ! -e "$HOME/.claude/skills/ast-grep" ] && [ ! -L "$HOME/.claude/skills/ast-grep" ]; then
    mkdir -p "$HOME/.claude/skills/ast-grep"
    sudo mount --bind /tmp/asylum-kit-skills-ast-grep "$HOME/.claude/skills/ast-grep"
fi
`,
		RulesSnippet: `### ast-grep (ast-grep kit)
ast-grep (` + "`sg`" + `) is installed for AST-based code search, linting, and rewriting. Use ` + "`sg run`" + ` to search with patterns, ` + "`sg scan`" + ` to lint, and ` + "`sg rewrite`" + ` to apply transformations. Patterns use ` + "`$VAR`" + ` for wildcards. Example: ` + "`sg run -p 'fmt.Errorf($MSG)' -l go`" + `.
`,
		BannerLines: `    echo "ast-grep:  $(sg --version 2>/dev/null | head -1 || echo 'not found')"
`,
	})
}
