package docker

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func DockerAvailable() error {
	cmd := exec.Command("docker", "info")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker daemon not running (is Docker installed and started?): %w", err)
	}
	return nil
}

func Build(contextDir, dockerfilePath, tag string, labels, buildArgs map[string]string, noCache bool) error {
	args := []string{"build", "-f", dockerfilePath, "-t", tag}
	for k, v := range labels {
		args = append(args, "--label", k+"="+v)
	}
	for k, v := range buildArgs {
		args = append(args, "--build-arg", k+"="+v)
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
