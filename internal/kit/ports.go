package kit

func init() {
	Register(&Kit{
		Name:        "ports",
		Description: "Automatic port forwarding for web services",
		DefaultOn:   true,
		ConfigSnippet: `  # ports:              # Automatic port forwarding (5 ports per project)
  #   count: 5          # Number of ports to allocate
`,
	})
}
