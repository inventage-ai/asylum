package kit

func init() {
	Register(&Kit{
		Name:        "ports",
		Description: "Automatic port forwarding for web services",
		Tier:        TierAlwaysOn,
		ConfigSnippet: `  # ports:              # Automatic port forwarding (on by default)
  #   count: 5          # Number of ports to allocate
`,
		ConfigComment: "ports:                # Automatic port forwarding (on by default)\n  count: 5            # Number of ports to allocate",
	})
}
