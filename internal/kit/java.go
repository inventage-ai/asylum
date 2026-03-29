package kit

func init() {
	Register(&Kit{
		Name:        "java",
		Description: "Java via mise with JDK 17/21/25",
		DockerSnippet: `# Install Java versions via mise
RUN ~/.local/bin/mise install java@17 java@21 java@25 && \
    ~/.local/bin/mise use --global java@21
`,
		RulesSnippet: `### Java (java kit)
JDK 17, 21, and 25 are installed via mise. The default is 21. Switch versions with ` + "`mise use java@<version>`" + `.
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
		SubKits: map[string]*Kit{
			"maven": {
				Name:        "java/maven",
				Description: "Maven with dependency caching",
				Tools:       []string{"mvn"},
				DockerSnippet: `# Install Maven
USER root
RUN apt-get update && apt-get install -y --no-install-recommends maven && rm -rf /var/lib/apt/lists/*
USER claude
`,
				CacheDirs: map[string]string{"maven": "/home/claude/.m2"},
			},
			"gradle": {
				Name:        "java/gradle",
				Description: "Gradle via mise with dependency caching",
				Tools:       []string{"gradle"},
				RulesSnippet: `### Gradle (java/gradle kit)
Gradle is installed via mise. Dependencies are cached across container restarts.
`,
				DockerSnippet: `# Install Gradle via mise
RUN ~/.local/bin/mise install gradle@latest && \
    ~/.local/bin/mise use --global gradle@latest
`,
				CacheDirs: map[string]string{"gradle": "/home/claude/.gradle"},
			},
		},
	})
}
