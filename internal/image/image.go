package image

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/inventage-ai/asylum/assets"
	"github.com/inventage-ai/asylum/internal/agent"
	"github.com/inventage-ai/asylum/internal/docker"
	"github.com/inventage-ai/asylum/internal/log"
	"github.com/inventage-ai/asylum/internal/kit"
)

const baseTag = "asylum:latest"

// writeCore writes a core asset to the builder, ensuring it ends with a blank line separator.
func writeCore(b *strings.Builder, core []byte) {
	b.Write(core)
	if !strings.HasSuffix(string(core), "\n") {
		b.WriteByte('\n')
	}
	b.WriteByte('\n')
}

// assembleDockerfile builds a complete Dockerfile from core + profile snippets + agent snippets + tail.
func assembleDockerfile(profiles []*kit.Kit, agentInstalls []*agent.AgentInstall) []byte {
	var b strings.Builder
	writeCore(&b, assets.DockerfileCore)
	if snippets := kit.AssembleDockerSnippets(profiles); snippets != "" {
		b.WriteString(snippets)
	}
	if snippets := agent.AssembleAgentSnippets(agentInstalls); snippets != "" {
		b.WriteString(snippets)
	}
	b.Write(assets.DockerfileTail)
	return []byte(b.String())
}

// assembleEntrypoint builds a complete entrypoint from core + profile snippets + tail.
// Banner lines from profiles and agents are inserted at the PROFILE_BANNER_PLACEHOLDER marker.
func assembleEntrypoint(profiles []*kit.Kit, agentInstalls []*agent.AgentInstall) []byte {
	var b strings.Builder
	writeCore(&b, assets.EntrypointCore)
	if snippets := kit.AssembleEntrypointSnippets(profiles); snippets != "" {
		b.WriteString(snippets)
		b.WriteByte('\n')
	}

	// Insert banner lines at placeholder in tail
	tail := string(assets.EntrypointTail)
	bannerLines := kit.AssembleBannerLines(profiles) + agent.AssembleAgentBannerLines(agentInstalls)
	tail = strings.Replace(tail, "# PROFILE_BANNER_PLACEHOLDER\n", bannerLines, 1)

	b.WriteString(tail)
	return []byte(b.String())
}

func baseHash(profiles []*kit.Kit, agentInstalls []*agent.AgentInstall) string {
	h := sha256.New()
	h.Write(assets.DockerfileCore)
	h.Write(assets.DockerfileTail)
	h.Write(assets.EntrypointCore)
	h.Write(assets.EntrypointTail)
	h.Write([]byte(kit.AssembleDockerSnippets(profiles)))
	h.Write([]byte(kit.AssembleEntrypointSnippets(profiles)))
	h.Write([]byte(kit.AssembleBannerLines(profiles)))
	h.Write([]byte(agent.AssembleAgentSnippets(agentInstalls)))
	h.Write([]byte(agent.AssembleAgentBannerLines(agentInstalls)))
	// Include host user identity so username/homedir changes trigger rebuild
	h.Write([]byte(fmt.Sprintf("uid=%d gid=%d", os.Getuid(), os.Getgid())))
	if home, err := os.UserHomeDir(); err == nil {
		h.Write([]byte(home))
	}
	if u, err := user.Current(); err == nil {
		h.Write([]byte(u.Username))
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func buildImage(dockerfileContent []byte, extraFiles map[string][]byte, tag string, labels, buildArgs map[string]string, noCache bool) error {
	tmpDir, err := os.MkdirTemp("", "asylum-build-")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	dfPath := filepath.Join(tmpDir, "Dockerfile")
	if err := os.WriteFile(dfPath, dockerfileContent, 0644); err != nil {
		return err
	}
	for name, content := range extraFiles {
		if err := os.WriteFile(filepath.Join(tmpDir, name), content, 0755); err != nil {
			return err
		}
	}

	return docker.Build(tmpDir, dfPath, tag, labels, buildArgs, noCache)
}

func EnsureBase(profiles []*kit.Kit, agentInstalls []*agent.AgentInstall, version string, noCache bool) (bool, error) {
	hash := baseHash(profiles, agentInstalls)

	existing, err := docker.InspectLabel(baseTag, "asylum.hash")
	if err == nil && existing == hash && !noCache {
		log.Info("base image up to date")
		return false, nil
	}

	log.Build("building base image...")

	labels := map[string]string{
		"asylum.hash":    hash,
		"asylum.version": version,
	}
	home, _ := os.UserHomeDir()
	username := "claude"
	if u, err := user.Current(); err == nil {
		username = u.Username
	}

	buildArgs := map[string]string{
		"USER_ID":   fmt.Sprintf("%d", os.Getuid()),
		"GROUP_ID":  fmt.Sprintf("%d", os.Getgid()),
		"USERNAME":  username,
		"USER_HOME": home,
	}

	dockerfile := assembleDockerfile(profiles, agentInstalls)
	entrypoint := assembleEntrypoint(profiles, agentInstalls)

	if err := buildImage(dockerfile, map[string][]byte{"entrypoint.sh": entrypoint}, baseTag, labels, buildArgs, noCache); err != nil {
		return false, fmt.Errorf("build base image: %w", err)
	}

	log.Success("base image built")
	if err := docker.PruneImages("asylum.version"); err != nil {
		log.Error("prune old base images: %v", err)
	}

	return true, nil
}

// Pre-installed Java versions in the base image.
var preinstalledJava = map[string]bool{"17": true, "21": true, "25": true}

func EnsureProject(projectProfiles []*kit.Kit, packages map[string][]string, javaVersion string, version string, baseRebuilt bool, noCache bool) (string, error) {
	profileSnippets := kit.AssembleDockerSnippets(projectProfiles)
	needsCustomJava := javaVersion != "" && !preinstalledJava[javaVersion]

	if len(packages) == 0 && !needsCustomJava && profileSnippets == "" {
		return baseTag, nil
	}

	username := "claude"
	if u, err := user.Current(); err == nil {
		username = u.Username
	}
	dockerfile, err := generateProjectDockerfile(profileSnippets, packages, javaVersion, username)
	if err != nil {
		return "", err
	}
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(dockerfile)))
	tag := "asylum:proj-" + hash[:12]

	existing, err := docker.InspectLabel(tag, "asylum.packages.hash")
	if err == nil && existing == hash && !baseRebuilt && !noCache {
		log.Info("project image up to date")
		return tag, nil
	}

	log.Build("building project image...")

	labels := map[string]string{
		"asylum.packages.hash": hash,
		"asylum.version":       version,
	}

	if err := buildImage([]byte(dockerfile), nil, tag, labels, nil, noCache); err != nil {
		return "", fmt.Errorf("build project image: %w", err)
	}

	log.Success("project image built")
	return tag, nil
}

var knownPackageTypes = map[string]bool{
	"apt": true, "npm": true, "pip": true, "run": true, "cx-lang": true,
}

var validPackageName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9+\-.@:/~_]*$`)
var validJavaVersion = regexp.MustCompile(`^[0-9]+(\.[0-9]+)*$`)

func validatePackageNames(pkgType string, names []string) error {
	for _, name := range names {
		if !validPackageName.MatchString(name) {
			return fmt.Errorf("invalid %s package name %q", pkgType, name)
		}
	}
	return nil
}

func generateProjectDockerfile(profileSnippets string, packages map[string][]string, javaVersion string, username string) (string, error) {
	for k := range packages {
		if !knownPackageTypes[k] {
			return "", fmt.Errorf("unknown package type %q (valid: apt, npm, pip, cx-lang, run)", k)
		}
	}
	for _, pkgType := range []string{"apt", "npm", "pip", "cx-lang"} {
		if err := validatePackageNames(pkgType, packages[pkgType]); err != nil {
			return "", err
		}
	}
	for _, cmd := range packages["run"] {
		if strings.ContainsAny(cmd, "\n\r") {
			return "", fmt.Errorf("invalid run command: must not contain newlines")
		}
		if strings.TrimSpace(cmd) == "" {
			return "", fmt.Errorf("invalid run command: must not be empty")
		}
	}

	var b strings.Builder
	b.WriteString("FROM asylum:latest\n")

	// Project-level profile snippets
	if profileSnippets != "" {
		b.WriteString("\nUSER " + username + "\n")
		b.WriteString(profileSnippets)
	}

	if apt := packages["apt"]; len(apt) > 0 {
		b.WriteString("\nUSER root\n")
		b.WriteString("RUN apt-get update && apt-get install -y --no-install-recommends \\\n    ")
		b.WriteString(strings.Join(apt, " \\\n    "))
		b.WriteString(" \\\n    && rm -rf /var/lib/apt/lists/*\n")
	}

	if npm := packages["npm"]; len(npm) > 0 {
		b.WriteString("\nUSER " + username + "\n")
		b.WriteString("RUN bash -c 'eval \"$(fnm env)\" && npm install -g \\\n    ")
		b.WriteString(strings.Join(npm, " \\\n    "))
		b.WriteString("'\n")
	}

	writeUserRuns := func(prefix string, items []string) {
		if len(items) == 0 {
			return
		}
		b.WriteString("\nUSER " + username + "\n")
		for _, item := range items {
			if strings.TrimSpace(item) == "" {
				continue
			}
			b.WriteString("RUN " + prefix + item + "\n")
		}
	}

	writeUserRuns("$HOME/.local/bin/uv tool install ", packages["pip"])
	writeUserRuns("cx lang add ", packages["cx-lang"])
	writeUserRuns("", packages["run"])

	if javaVersion != "" && !preinstalledJava[javaVersion] {
		if !validJavaVersion.MatchString(javaVersion) {
			return "", fmt.Errorf("invalid java version %q", javaVersion)
		}
		b.WriteString("\nUSER " + username + "\n")
		b.WriteString("RUN $HOME/.local/bin/mise install java@" + javaVersion + " && $HOME/.local/bin/mise use --global java@" + javaVersion + "\n")
	}

	b.WriteString("\nUSER " + username + "\n")

	return b.String(), nil
}
