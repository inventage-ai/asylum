package kit

func init() {
	Register(&Kit{
		Name:        "apt",
		Description: "Extra apt packages installed in the project image",
		Tier:        TierOptIn,
		ConfigSnippet: `  # apt:                # System packages installed via apt-get
  #   packages:
  #     - imagemagick
  #     - ffmpeg
`,
		ConfigComment: "apt:                  # System packages installed via apt-get\n  packages:\n    - imagemagick\n    - ffmpeg",
	})
}
