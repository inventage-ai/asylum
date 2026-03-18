package image

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/inventage-ai/asylum/assets"
	"github.com/inventage-ai/asylum/internal/docker"
	"github.com/inventage-ai/asylum/internal/log"
)

const baseTag = "asylum:latest"

func assetHash() string {
	h := sha256.New()
	h.Write(assets.Dockerfile)
	h.Write(assets.Entrypoint)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func EnsureBase(version string, noCache bool) (bool, error) {
	hash := assetHash()

	existing, err := docker.InspectLabel(baseTag, "asylum.hash")
	if err == nil && existing == hash && !noCache {
		log.Info("base image up to date")
		return false, nil
	}

	log.Build("building base image...")

	tmpDir, err := os.MkdirTemp("", "asylum-build-")
	if err != nil {
		return false, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := os.WriteFile(filepath.Join(tmpDir, "Dockerfile"), assets.Dockerfile, 0644); err != nil {
		return false, err
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "entrypoint.sh"), assets.Entrypoint, 0755); err != nil {
		return false, err
	}

	labels := map[string]string{
		"asylum.hash":    hash,
		"asylum.version": version,
	}
	buildArgs := map[string]string{
		"USER_ID":  fmt.Sprintf("%d", os.Getuid()),
		"GROUP_ID": fmt.Sprintf("%d", os.Getgid()),
		"USERNAME": "claude",
	}

	if err := docker.Build(tmpDir, filepath.Join(tmpDir, "Dockerfile"), baseTag, labels, buildArgs, noCache); err != nil {
		return false, fmt.Errorf("build base image: %w", err)
	}

	log.Success("base image built")
	if err := docker.PruneImages("asylum.version"); err != nil {
		log.Error("prune old base images: %v", err)
	}

	return true, nil
}

func EnsureProject(packages map[string][]string, version string, baseRebuilt bool, noCache bool) (string, error) {
	if len(packages) == 0 {
		return baseTag, nil
	}

	dockerfile, err := generateProjectDockerfile(packages)
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

	tmpDir, err := os.MkdirTemp("", "asylum-proj-")
	if err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	dfPath := filepath.Join(tmpDir, "Dockerfile")
	if err := os.WriteFile(dfPath, []byte(dockerfile), 0644); err != nil {
		return "", err
	}

	labels := map[string]string{
		"asylum.packages.hash": hash,
		"asylum.version":       version,
	}

	if err := docker.Build(tmpDir, dfPath, tag, labels, nil, noCache); err != nil {
		return "", fmt.Errorf("build project image: %w", err)
	}

	log.Success("project image built")
	return tag, nil
}

var knownPackageTypes = map[string]bool{
	"apt": true, "npm": true, "pip": true, "run": true,
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

func generateProjectDockerfile(packages map[string][]string) (string, error) {
	for k := range packages {
		if !knownPackageTypes[k] {
			return "", fmt.Errorf("unknown package type %q (valid: apt, npm, pip, run)", k)
		}
	}
	for _, pkgType := range []string{"apt", "npm", "pip"} {
		if err := validatePackageNames(pkgType, packages[pkgType]); err != nil {
			return "", err
		}
	}

	var b strings.Builder
	b.WriteString("FROM asylum:latest\n")

	if apt := packages["apt"]; len(apt) > 0 {
		b.WriteString("\nUSER root\n")
		b.WriteString("RUN apt-get update && apt-get install -y --no-install-recommends \\\n    ")
		b.WriteString(strings.Join(apt, " \\\n    "))
		b.WriteString(" \\\n    && rm -rf /var/lib/apt/lists/*\n")
	}

	if npm := packages["npm"]; len(npm) > 0 {
		b.WriteString("\nUSER claude\n")
		b.WriteString("RUN bash -c \"source $HOME/.nvm/nvm.sh && npm install -g \\\n    ")
		b.WriteString(strings.Join(npm, " \\\n    "))
		b.WriteString("\"\n")
	}

	writeUserRuns := func(prefix string, items []string) {
		if len(items) == 0 {
			return
		}
		b.WriteString("\nUSER claude\n")
		for _, item := range items {
			b.WriteString("RUN " + prefix + item + "\n")
		}
	}

	writeUserRuns("$HOME/.local/bin/uv tool install ", packages["pip"])
	writeUserRuns("", packages["run"])

	// Ensure we always end as the non-root user
	b.WriteString("\nUSER claude\n")

	return b.String(), nil
}
