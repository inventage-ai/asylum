package kit

import (
	"github.com/inventage-ai/asylum/internal/onboarding"

	"gopkg.in/yaml.v3"
)

func init() {
	Register(&Kit{
		Name:           "node",
		Description:    "Node.js global development packages",
		DockerPriority: 14,
		Tier:           TierAlwaysOn,
		ConfigSnippet: `  node:
    shadow-node-modules: true
    onboarding: false
    # versions:
    #   - 24
    # packages:          # npm packages installed globally
    #   - turbo
`,
		ConfigNodes: configNodes("node", "", []*yaml.Node{
			ScalarNode("shadow-node-modules", ""),
			BoolNode(true),
			ScalarNode("onboarding", ""),
			BoolNode(false),
		}),
		ConfigComment: "versions:\n#   - 24\n# packages:          # npm packages installed globally\n#   - turbo",
		DockerSnippet: `# Install Node.js global packages
RUN bash -c 'export PATH="$HOME/.local/share/fnm:$PATH" && eval "$(fnm env)" && \
    npm install -g \
        typescript \
        @types/node \
        ts-node \
        eslint \
        prettier \
        nodemon'
`,
		RulesSnippet: `### Node.js (node kit)
Node.js LTS is installed via fnm. Global packages: typescript, @types/node, ts-node, eslint, prettier, nodemon. Switch Node versions with ` + "`fnm use <version>`" + `.
`,
		BannerLines: `    echo "Node.js:   $(node --version 2>/dev/null || echo 'not found')"
`,
		SubKits: map[string]*Kit{
			"npm": {
				Name:            "node/npm",
				Description:     "npm with caching and onboarding",
				CacheDirs:       map[string]string{"npm": "~/.npm"},
				OnboardingTasks: []onboarding.Task{onboarding.NPMTask{}},
			},
			"pnpm": {
				Name:           "node/pnpm",
				Description:    "pnpm global install",
				DockerPriority: 14,
				Tools:       []string{"pnpm"},
				DockerSnippet: `# Install pnpm
RUN bash -c 'export PATH="$HOME/.local/share/fnm:$PATH" && eval "$(fnm env)" && npm install -g pnpm'
`,
			},
			"yarn": {
				Name:           "node/yarn",
				Description:    "yarn global install",
				DockerPriority: 14,
				Tools:       []string{"yarn"},
				DockerSnippet: `# Install yarn
RUN bash -c 'export PATH="$HOME/.local/share/fnm:$PATH" && eval "$(fnm env)" && npm install -g yarn'
`,
			},
		},
	})
}
