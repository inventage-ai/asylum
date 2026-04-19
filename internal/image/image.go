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

// assembleDockerfile builds a complete Dockerfile from core + ordered snippets + tail.
// Snippets are written in the order specified by orderedIDs, which is computed
// by computeSourceOrder to optimize Docker layer caching.
func assembleDockerfile(orderedIDs []string, snippetOf map[string]string) []byte {
	var b strings.Builder
	writeCore(&b, assets.DockerfileCore)
	for _, id := range orderedIDs {
		s := snippetOf[id]
		if s != "" {
			b.WriteString(s)
			if !strings.HasSuffix(s, "\n") {
				b.WriteByte('\n')
			}
		}
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

// assembleProjectEntrypoint builds an optional entrypoint script from project-level kit snippets.
// Returns nil if no project kits have EntrypointSnippets or BannerLines.
func assembleProjectEntrypoint(projectKits []*kit.Kit) []byte {
	snippets := kit.AssembleEntrypointSnippets(projectKits)
	bannerLines := kit.AssembleBannerLines(projectKits)
	if snippets == "" && bannerLines == "" {
		return nil
	}

	var b strings.Builder
	b.WriteString("#!/bin/bash\nset -e\n\n")
	if snippets != "" {
		b.WriteString(snippets)
		if !strings.HasSuffix(snippets, "\n") {
			b.WriteByte('\n')
		}
		b.WriteByte('\n')
	}
	if bannerLines != "" {
		b.WriteString("export PROJECT_BANNER='")
		b.WriteString(strings.ReplaceAll(bannerLines, "'", "'\\''"))
		b.WriteString("'\n")
	}
	return []byte(b.String())
}

func baseHash(orderedIDs []string, snippetOf map[string]string, profiles []*kit.Kit, agentInstalls []*agent.AgentInstall) string {
	h := sha256.New()
	h.Write(assets.DockerfileCore)
	h.Write(assets.DockerfileTail)
	h.Write(assets.EntrypointCore)
	h.Write(assets.EntrypointTail)
	// Hash ordered snippets so that reordering triggers a rebuild
	for _, id := range orderedIDs {
		h.Write([]byte(id))
		h.Write([]byte(snippetOf[id]))
	}
	h.Write([]byte(kit.AssembleEntrypointSnippets(profiles)))
	h.Write([]byte(kit.AssembleBannerLines(profiles)))
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

// EnsureBase builds the base image if needed, using computed source ordering
// for optimal Docker layer caching. previousOrder is the source order from the
// last successful build (from state.json). Returns (rebuilt, newOrder, err)
// where newOrder should be saved to state on success.
func EnsureBase(profiles []*kit.Kit, agentInstalls []*agent.AgentInstall, kitConfig func(string) *kit.SnippetConfig, version string, noCache bool, previousOrder []string) (bool, []string, error) {
	sources := collectSources(profiles, kitConfig, agentInstalls)
	orderedIDs := computeSourceOrder(sources, previousOrder)

	snippetOf := map[string]string{}
	for _, s := range sources {
		snippetOf[s.ID] = s.Snippet
	}

	hash := baseHash(orderedIDs, snippetOf, profiles, agentInstalls)

	existing, err := docker.InspectLabel(baseTag, "asylum.hash")
	if err == nil && existing == hash && !noCache {
		log.Info("base image up to date")
		return false, orderedIDs, nil
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

	dockerfile := assembleDockerfile(orderedIDs, snippetOf)
	entrypoint := assembleEntrypoint(profiles, agentInstalls)

	if err := buildImage(dockerfile, map[string][]byte{"entrypoint.sh": entrypoint}, baseTag, labels, buildArgs, noCache); err != nil {
		return false, nil, fmt.Errorf("build base image: %w", err)
	}

	log.Success("base image built")
	if err := docker.PruneImages("asylum.version"); err != nil {
		log.Error("prune old base images: %v", err)
	}

	return true, orderedIDs, nil
}

func EnsureProject(projectProfiles []*kit.Kit, allKits []*kit.Kit, packages map[string][]string, kitConfig func(string) *kit.SnippetConfig, version string, baseRebuilt bool, noCache bool) (string, error) {
	profileSnippets := kit.AssembleDockerSnippets(projectProfiles, kitConfig)
	projectEntrypoint := assembleProjectEntrypoint(projectProfiles)
	kitProjectSnippets := kit.AssembleProjectSnippets(allKits, kitConfig)

	if len(packages) == 0 && kitProjectSnippets == "" && profileSnippets == "" && projectEntrypoint == nil {
		return baseTag, nil
	}

	username := "claude"
	if u, err := user.Current(); err == nil {
		username = u.Username
	}
	dockerfile, err := generateProjectDockerfile(profileSnippets, packages, kitProjectSnippets, username, projectEntrypoint != nil)
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

	var extraFiles map[string][]byte
	if projectEntrypoint != nil {
		extraFiles = map[string][]byte{"project-entrypoint.sh": projectEntrypoint}
	}

	if err := buildImage([]byte(dockerfile), extraFiles, tag, labels, nil, noCache); err != nil {
		return "", fmt.Errorf("build project image: %w", err)
	}

	log.Success("project image built")
	return tag, nil
}

var knownPackageTypes = map[string]bool{
	"apt": true, "npm": true, "pip": true, "run": true, "cx-lang": true,
}

var validPackageName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9+\-.@:/~_]*$`)

func validatePackageNames(pkgType string, names []string) error {
	for _, name := range names {
		if !validPackageName.MatchString(name) {
			return fmt.Errorf("invalid %s package name %q", pkgType, name)
		}
	}
	return nil
}

func generateProjectDockerfile(profileSnippets string, packages map[string][]string, kitProjectSnippets string, username string, hasProjectEntrypoint bool) (string, error) {
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

	if kitProjectSnippets != "" {
		b.WriteString("\nUSER " + username + "\n")
		b.WriteString(kitProjectSnippets)
	}

	if hasProjectEntrypoint {
		b.WriteString("\nUSER root\n")
		b.WriteString("COPY --chmod=755 project-entrypoint.sh /usr/local/bin/project-entrypoint.sh\n")
	}

	b.WriteString("\nUSER " + username + "\n")

	return b.String(), nil
}
