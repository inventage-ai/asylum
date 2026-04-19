package kit

func init() {
	Register(&Kit{
		Name:           "ast-grep",
		Description:    "AST-based code search, lint, and rewrite",
		DockerPriority: 44,
		Tier:           TierOptIn,
		Deps:           []string{"node"},
		Tools:          []string{"sg"},
		ProvidesSkills: true,
		ConfigSnippet: `  # ast-grep:           # AST-based code search (sg)
`,
		ConfigNodes:   configNodes("ast-grep", "AST-based code search (sg)", nil),
		ConfigComment: "ast-grep:             # AST-based code search (sg)",
		DockerSnippet: `# Install ast-grep
RUN bash -c 'export PATH="$HOME/.local/share/fnm:$PATH" && eval "$(fnm env)" && \
    npm install -g @ast-grep/cli'
RUN sudo mkdir -p /opt/asylum-skills/.claude/skills && \
    sudo chown -R "$(id -u):$(id -g)" /opt/asylum-skills && \
    bash -c 'export PATH="$HOME/.local/share/fnm:$PATH" && eval "$(fnm env)" && \
        cd /tmp && npx skills add ast-grep/agent-skill --skill ast-grep --yes --copy && \
        mv .claude/skills/ast-grep /opt/asylum-skills/.claude/skills/ast-grep' || true
`,
		RulesSnippet: `### ast-grep (ast-grep kit)
ast-grep (` + "`sg`" + `) is installed for AST-based code search, linting, and rewriting. Use ` + "`sg run`" + ` to search with patterns, ` + "`sg scan`" + ` to lint, and ` + "`sg rewrite`" + ` to apply transformations. Patterns use ` + "`$VAR`" + ` for wildcards. Example: ` + "`sg run -p 'fmt.Errorf($MSG)' -l go`" + `.
`,
		BannerLines: `    echo "ast-grep:  $(sg --version 2>/dev/null | head -1 || echo 'not found')"
`,
	})
}
