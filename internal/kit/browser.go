package kit

func init() {
	Register(&Kit{
		Name:        "browser",
		Description: "Chromium browser via Playwright",
		Tier:        TierOptIn,
		Deps:        []string{"node"},
		Tools:       []string{"playwright"},
		CacheDirs:   map[string]string{"playwright": "/home/claude/.cache/ms-playwright"},
		ConfigSnippet: `  # browser:           # Chromium browser via Playwright
`,
		ConfigNodes:   configNodes("browser", "Chromium browser via Playwright", nil),
		ConfigComment: "browser:             # Chromium browser via Playwright",
		DockerSnippet: `# Install Playwright and Chromium
RUN bash -c 'export PATH="$HOME/.local/share/fnm:$PATH" && eval "$(fnm env)" && \
    npm install -g playwright'
USER root
RUN bash -c 'export PATH="$HOME/.local/share/fnm:$PATH" && eval "$(fnm env)" && \
    npx playwright install --with-deps chromium'
USER claude
`,
		RulesSnippet: `### Browser (browser kit)
Chromium is installed via Playwright for browser automation. Use ` + "`npx playwright`" + ` to launch browsers, take screenshots, navigate pages, and interact with web content.
`,
		BannerLines: `    echo "Chromium:  $(chromium --version 2>/dev/null || echo 'not found')"
`,
	})
}
