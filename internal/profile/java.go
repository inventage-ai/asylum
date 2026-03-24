package profile

import "github.com/inventage-ai/asylum/internal/config"

func init() {
	Register(&Profile{
		Name:        "java",
		Description: "Java via mise with JDK 17/21/25",
		DockerSnippet: `# Install Java versions via mise
RUN ~/.local/bin/mise install java@17 java@21 java@25 && \
    ~/.local/bin/mise use --global java@21
`,
		EntrypointSnippet: `# Activate mise for Java/Gradle
if command -v mise >/dev/null 2>&1; then
    eval "$(mise activate bash)"
    if [ -n "${ASYLUM_JAVA_VERSION:-}" ]; then
        mise use --global java@"${ASYLUM_JAVA_VERSION}" >/dev/null 2>&1
        eval "$(mise env)"
    fi
fi
`,
		BannerLines: `    echo "Java:      $(java -version 2>&1 | head -1 | cut -d'"' -f2 || echo 'not found')"
`,
		Config: config.Config{
			Versions: map[string]string{"java": "21"},
		},
		SubProfiles: map[string]*Profile{
			"maven": {
				Name:        "java/maven",
				Description: "Maven dependency caching",
				CacheDirs:   map[string]string{"maven": "/home/claude/.m2"},
			},
			"gradle": {
				Name:        "java/gradle",
				Description: "Gradle via mise with dependency caching",
				DockerSnippet: `# Install Gradle via mise
RUN ~/.local/bin/mise install gradle@latest && \
    ~/.local/bin/mise use --global gradle@latest
`,
				CacheDirs: map[string]string{"gradle": "/home/claude/.gradle"},
			},
		},
	})
}
