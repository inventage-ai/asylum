package kit

func init() {
	Register(&Kit{
		Name:        "title",
		Description: "Terminal tab title and agent title configuration",
		Tier:        TierAlwaysOn,
		ConfigSnippet: `  # title:              # Terminal tab title configuration (on by default)
  #   # Placeholders: {project}, {agent}, {mode}
  #   tab-title: "🤖 {project}"
  #   allow-agent-terminal-title: false
`,
		ConfigComment: "title:                # Terminal tab title configuration (on by default)\n  # Placeholders: {project}, {agent}, {mode}\n  tab-title: \"🤖 {project}\"\n  allow-agent-terminal-title: false",
	})
}
