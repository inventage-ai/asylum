package kit

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

func init() {
	k := &Kit{
		Name:           "java",
		Description:    "Java via mise with JDK 17/21/25",
		DockerPriority: 10,
		ConfigSnippet: `  java:
    versions:
      - 17
      - 21
      - 25
    default-version: 21
`,
		ConfigComment: "java:\n  versions:\n    - 17\n    - 21\n    - 25\n  default-version: 21",
		ConfigNodes: configNodes("java", "", []*yaml.Node{
			ScalarNode("versions", ""),
			SeqNode("17", "21", "25"),
			ScalarNode("default-version", ""),
			IntNode(21),
		}),
		DockerSnippet: `# Install Java versions via mise
RUN ~/.local/bin/mise install java@17 java@21 java@25 && \
    ~/.local/bin/mise use --global java@21
`,
		RulesSnippet: `### Java (java kit)
JDK 17, 21, and 25 are installed via mise. The default is 21. Switch versions with ` + "`mise use java@<version>`" + `.
`,
		EntrypointSnippet: `# Select Java version if configured
if [ -n "${ASYLUM_JAVA_VERSION:-}" ] && command -v mise >/dev/null 2>&1; then
    mise use --global java@"${ASYLUM_JAVA_VERSION}" >/dev/null 2>&1
    eval "$(mise env)"
fi
`,
		BannerLines: `    echo "Java:      $(java -version 2>&1 | head -1 | cut -d'"' -f2 || echo 'not found')"
`,
		SubKits: map[string]*Kit{
			"maven": {
				Name:            "java/maven",
				Description:     "Maven with dependency caching",
				DockerPriority:  10,
				Tools:           []string{"mvn"},
				CredentialFunc:  mavenCredentialFunc,
				CredentialLabel: "Java/Maven",
				DockerSnippet: `# Install Maven
USER root
RUN apt-get update && apt-get install -y --no-install-recommends maven && rm -rf /var/lib/apt/lists/*
USER ${USERNAME}
`,
				CacheDirs: map[string]string{"maven": "~/.m2"},
			},
			"gradle": {
				Name:           "java/gradle",
				Description:    "Gradle via mise with dependency caching",
				DockerPriority: 10,
				Tools:       []string{"gradle"},
				RulesSnippet: `### Gradle (java/gradle kit)
Gradle is installed via mise. Dependencies are cached across container restarts.
`,
				DockerSnippet: `# Install Gradle via mise
RUN ~/.local/bin/mise install gradle@latest && \
    ~/.local/bin/mise use --global gradle@latest
`,
				CacheDirs: map[string]string{"gradle": "~/.gradle"},
			},
		},
	}

	k.ConfigureFunc = func(versions []string, defaultVersion string) {
		if len(versions) == 0 {
			return
		}
		if defaultVersion == "" {
			defaultVersion = versions[0]
		}

		// Build "java@17 java@21" style args
		var installs []string
		for _, v := range versions {
			installs = append(installs, "java@"+v)
		}

		k.DockerSnippet = fmt.Sprintf("# Install Java versions via mise\nRUN ~/.local/bin/mise install %s && \\\n    ~/.local/bin/mise use --global java@%s\n",
			strings.Join(installs, " "), defaultVersion)

		k.RulesSnippet = fmt.Sprintf("### Java (java kit)\nJDK %s installed via mise. The default is %s. Switch versions with `mise use java@<version>`.\n",
			strings.Join(versions, ", "), defaultVersion)

		k.Description = fmt.Sprintf("Java via mise with JDK %s", strings.Join(versions, "/"))
	}

	Register(k)
}
