package docker

import (
	"fmt"
	"maps"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"
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

func RunDetached(args []string) error {
	cmd := exec.Command("docker", args...)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Exec(container, user string, command ...string) error {
	args := []string{"exec", "-u", user, container}
	args = append(args, command...)
	return exec.Command("docker", args...).Run()
}

// ReadFile reads a file from inside a running container.
func ReadFile(container, path string) (string, error) {
	out, err := exec.Command("docker", "exec", container, "cat", path).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// CopyTo copies the contents of a host directory into a container path.
func CopyTo(container, hostDir, containerPath string) error {
	return exec.Command("docker", "cp", hostDir+"/.", container+":"+containerPath).Run()
}

func RemoveContainer(name string) error {
	cmd := exec.Command("docker", "rm", "-f", name)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func ShowLogs(name string) {
	cmd := exec.Command("docker", "logs", name)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	cmd.Run()
}

// WaitReady polls until the container is running and the entrypoint has
// finished (indicated by /tmp/asylum-path existing), up to timeoutSec seconds.
func WaitReady(name string, timeoutSec int) bool {
	for i := 0; i < timeoutSec; i++ {
		if IsRunning(name) && fileExists(name, "/tmp/asylum-path") {
			return true
		}
		time.Sleep(time.Second)
	}
	return false
}

func fileExists(container, path string) bool {
	return exec.Command("docker", "exec", container, "test", "-f", path).Run() == nil
}

func IsRunning(name string) bool {
	cmd := exec.Command("docker", "inspect", "--format", "{{.State.Running}}", name)
	out, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(out)) == "true"
}

func ListVolumes(prefix string) ([]string, error) {
	cmd := exec.Command("docker", "volume", "ls", "--format", "{{.Name}}")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var volumes []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" && strings.HasPrefix(line, prefix) {
			volumes = append(volumes, line)
		}
	}
	return volumes, nil
}

func RemoveVolumes(volumes ...string) error {
	if len(volumes) == 0 {
		return nil
	}
	args := append([]string{"volume", "rm"}, volumes...)
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// HasOtherSessions checks whether other exec sessions are active in the
// container by running ps inside it and counting processes with PPID=0
// (which identifies docker exec injected processes in the PID namespace).
// PID 1 (docker-init) and the check command itself are excluded.
func HasOtherSessions(containerName string) bool {
	out, err := exec.Command("docker", "exec", containerName,
		"ps", "-o", "pid,ppid", "--no-headers").Output()
	if err != nil {
		return false
	}
	return countExecSessions(string(out)) > 1 // 1 = our check command
}

// countExecSessions counts processes with PPID=0 excluding PID 1.
func countExecSessions(psOutput string) int {
	count := 0
	for _, line := range strings.Split(psOutput, "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		pid, ppid := fields[0], fields[1]
		if ppid == "0" && pid != "1" {
			count++
		}
	}
	return count
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
