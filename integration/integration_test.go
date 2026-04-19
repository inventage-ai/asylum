//go:build integration

package integration_test

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/inventage-ai/asylum/internal/docker"
	"github.com/inventage-ai/asylum/internal/image"
	"github.com/inventage-ai/asylum/internal/kit"
)

var testVersion = fmt.Sprintf("test-%d", time.Now().Unix())

var baseOnce sync.Once
var baseErr error
var baseKits []*kit.Kit

func ensureBaseImage(t *testing.T) {
	t.Helper()
	baseOnce.Do(func() {
		baseKits, baseErr = kit.Resolve([]string{"java", "python"}, nil)
		if baseErr != nil {
			return
		}
		_, _, baseErr = image.EnsureBase(baseKits, nil, nil, testVersion, false, nil)
	})
	if baseErr != nil {
		t.Fatalf("base image build failed: %v", baseErr)
	}
}

func TestMain(m *testing.M) {
	if err := docker.DockerAvailable(); err != nil {
		fmt.Fprintf(os.Stderr, "skipping integration tests: %v\n", err)
		os.Exit(0)
	}

	code := m.Run()

	// Cleanup: remove all images labeled with our test version
	images, err := docker.ListImages("asylum:*")
	if err == nil {
		var toRemove []string
		for _, img := range images {
			v, err := docker.InspectLabel(img, "asylum.version")
			if err == nil && v == testVersion {
				toRemove = append(toRemove, img)
			}
		}
		if len(toRemove) > 0 {
			docker.RemoveImages(toRemove...)
		}
	}

	os.Exit(code)
}

func dockerRun(t *testing.T, script string) string {
	t.Helper()
	cmd := exec.Command("docker", "run", "--rm", "asylum:latest", "bash", "-c", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("docker run failed: %v\noutput: %s", err, out)
	}
	return strings.TrimSpace(string(out))
}

func dockerRunWithEnv(t *testing.T, env map[string]string, script string) string {
	t.Helper()
	args := []string{"run", "--rm"}
	for k, v := range env {
		args = append(args, "-e", k+"="+v)
	}
	args = append(args, "asylum:latest", "bash", "-c", script)
	cmd := exec.Command("docker", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("docker run failed: %v\noutput: %s", err, out)
	}
	return strings.TrimSpace(string(out))
}

func dockerRunWithVolume(t *testing.T, volume, script string) string {
	t.Helper()
	cmd := exec.Command("docker", "run", "--rm", "-v", volume, "asylum:latest", "bash", "-c", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("docker run failed: %v\noutput: %s", err, out)
	}
	return strings.TrimSpace(string(out))
}

func dockerRunWithProjectDir(t *testing.T, projDir, script string) string {
	t.Helper()
	cmd := exec.Command("docker", "run", "--rm",
		"-v", projDir+":"+projDir,
		"-w", projDir,
		"-e", "HOST_PROJECT_DIR="+projDir,
		"asylum:latest", "bash", "-c", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("docker run failed: %v\noutput: %s", err, out)
	}
	return strings.TrimSpace(string(out))
}

func dockerRunWithVolumeAndEnv(t *testing.T, volume string, env map[string]string, script string) string {
	t.Helper()
	args := []string{"run", "--rm", "-v", volume}
	for k, v := range env {
		args = append(args, "-e", k+"="+v)
	}
	args = append(args, "asylum:latest", "bash", "-c", script)
	cmd := exec.Command("docker", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("docker run failed: %v\noutput: %s", err, out)
	}
	return strings.TrimSpace(string(out))
}

func parseKeyValues(output string) map[string]string {
	m := make(map[string]string)
	for _, line := range strings.Split(output, "\n") {
		k, v, ok := strings.Cut(line, "=")
		if ok {
			m[k] = v
		}
	}
	return m
}
