package docker

import (
	"fmt"
	"maps"
	"os"
	"os/exec"
	"slices"
	"strings"
)

func DockerAvailable() error {
	if err := exec.Command("docker", "info").Run(); err != nil {
		return fmt.Errorf("docker daemon not running (is Docker installed and started?): %w", err)
	}
	return nil
}

func Build(contextDir, dockerfilePath, tag string, labels, buildArgs map[string]string, noCache bool) error {
	args := []string{"build", "-f", dockerfilePath, "-t", tag}
	for _, k := range slices.Sorted(maps.Keys(labels)) {
		args = append(args, "--label", k+"="+labels[k])
	}
	for _, k := range slices.Sorted(maps.Keys(buildArgs)) {
		args = append(args, "--build-arg", k+"="+buildArgs[k])
	}
	if noCache {
		args = append(args, "--no-cache")
	}
	args = append(args, contextDir)

	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func InspectLabel(image, label string) (string, error) {
	tmpl := fmt.Sprintf("{{index .Config.Labels %q}}", label)
	cmd := exec.Command("docker", "inspect", "--format", tmpl, image)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("inspect %s: %w", image, err)
	}
	return strings.TrimSpace(string(out)), nil
}

func RemoveImages(images ...string) error {
	if len(images) == 0 {
		return nil
	}
	args := append([]string{"rmi", "-f"}, images...)
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func PruneImages(filterLabel string) error {
	cmd := exec.Command("docker", "image", "prune", "-f", "--filter", "label="+filterLabel)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func IsRunning(name string) bool {
	cmd := exec.Command("docker", "inspect", "--format", "{{.State.Running}}", name)
	out, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(out)) == "true"
}

func ListImages(filter string) ([]string, error) {
	cmd := exec.Command("docker", "images", "--format", "{{.Repository}}:{{.Tag}}", "--filter", "reference="+filter)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var images []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			images = append(images, line)
		}
	}
	return images, nil
}
