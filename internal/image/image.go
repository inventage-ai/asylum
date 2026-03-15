package image

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/binaryben/asylum/assets"
	"github.com/binaryben/asylum/internal/docker"
	"github.com/binaryben/asylum/internal/log"
	"gopkg.in/yaml.v3"
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

	uid := fmt.Sprintf("%d", os.Getuid())
	gid := fmt.Sprintf("%d", os.Getgid())

	labels := map[string]string{
		"asylum.hash":    hash,
		"asylum.version": version,
	}
	buildArgs := map[string]string{
		"USER_ID":  uid,
		"GROUP_ID": gid,
		"USERNAME": "claude",
	}

	if err := docker.Build(tmpDir, filepath.Join(tmpDir, "Dockerfile"), baseTag, labels, buildArgs, noCache); err != nil {
		return false, fmt.Errorf("build base image: %w", err)
	}

	log.Success("base image built")
	docker.PruneImages("asylum.version")

	return true, nil
}

func EnsureProject(packages map[string][]string, version string, baseRebuilt bool) (string, error) {
	if len(packages) == 0 {
		return baseTag, nil
	}

	hash := packagesHash(packages)
	tag := "asylum:proj-" + hash[:12]

	existing, err := docker.InspectLabel(tag, "asylum.packages.hash")
	if err == nil && existing == hash && !baseRebuilt {
		log.Info("project image up to date")
		return tag, nil
	}

	log.Build("building project image...")

	dockerfile := generateProjectDockerfile(packages)

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

	if err := docker.Build(tmpDir, dfPath, tag, labels, nil, false); err != nil {
		return "", fmt.Errorf("build project image: %w", err)
	}

	log.Success("project image built")
	return tag, nil
}

func packagesHash(packages map[string][]string) string {
	data, _ := yaml.Marshal(packages)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

func generateProjectDockerfile(packages map[string][]string) string {
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

	if pip := packages["pip"]; len(pip) > 0 {
		b.WriteString("\nUSER claude\n")
		b.WriteString("RUN $HOME/.local/bin/uv tool install ")
		b.WriteString(strings.Join(pip, " "))
		b.WriteString("\n")
	}

	if run := packages["run"]; len(run) > 0 {
		b.WriteString("\nUSER claude\n")
		for _, cmd := range run {
			b.WriteString("RUN " + cmd + "\n")
		}
	}

	return b.String()
}
