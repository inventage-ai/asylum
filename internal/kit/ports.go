package kit

func init() {
	Register(&Kit{
		Name:        "ports",
		Description: "Automatic port forwarding for web services",
		Tier:        TierAlwaysOn,
		ConfigSnippet: `  # ports:              # Automatic port forwarding (5 ports per project)
  #   count: 5          # Number of ports to allocate
`,
		ConfigComment: "ports:              # Automatic port forwarding (5 ports per project)\n  count: 5          # Number of ports to allocate",
	})
}
