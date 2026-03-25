//go:build e2e

package e2e_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/inventage-ai/asylum/internal/docker"
)

var (
	binaryPath string
	testHome   string
	projectDir string
)

func TestMain(m *testing.M) {
	if err := docker.DockerAvailable(); err != nil {
		fmt.Fprintf(os.Stderr, "skipping e2e tests: %v\n", err)
		os.Exit(0)
	}

	// Build the binary
	tmpDir, err := os.MkdirTemp("", "asylum-e2e-")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create temp dir: %v\n", err)
		os.Exit(1)
	}

	binaryPath = filepath.Join(tmpDir, "asylum")
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/asylum")
	cmd.Dir = findRepoRoot()
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v\n%s\n", err, out)
		os.Exit(1)
	}

	// Set up test HOME with minimal config
	testHome, _ = os.MkdirTemp("", "asylum-e2e-home-")
	configDir := filepath.Join(testHome, ".asylum")
	os.MkdirAll(configDir, 0755)
	os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(`version: "0.2"
agent: echo
kits: {}
agents: {}
`), 0644)

	// Create agent config dir so EnsureAgentConfig doesn't prompt
	os.MkdirAll(filepath.Join(configDir, "agents", "echo"), 0755)

	// Set up test project directory
	projectDir, _ = os.MkdirTemp("", "asylum-e2e-project-")

	code := m.Run()

	// Cleanup
	cleanupContainers()
	cleanupImages()
	os.RemoveAll(tmpDir)
	os.RemoveAll(testHome)
	os.RemoveAll(projectDir)

	os.Exit(code)
}

func findRepoRoot() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "."
		}
		dir = parent
	}
}

func cleanupContainers() {
	out, err := exec.Command("docker", "ps", "-a", "--filter", "name=asylum-", "--format", "{{.Names}}").Output()
	if err != nil {
		return
	}
	for _, name := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if name != "" {
			exec.Command("docker", "rm", "-f", name).Run()
		}
	}
}

func cleanupImages() {
	images, err := docker.ListImages("asylum:*")
	if err != nil {
		return
	}
	if len(images) > 0 {
		docker.RemoveImages(images...)
	}
}

type result struct {
	stdout   string
	stderr   string
	exitCode int
}

func runAsylum(t *testing.T, args ...string) result {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = projectDir
	cmd.Env = append(os.Environ(),
		"HOME="+testHome,
	)
	cmd.Stdin = strings.NewReader("") // no TTY — prevents docker exec -t

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to run asylum: %v", err)
		}
	}

	return result{
		stdout:   stdout.String(),
		stderr:   stderr.String(),
		exitCode: exitCode,
	}
}

func runAsylumSuccess(t *testing.T, args ...string) result {
	t.Helper()
	r := runAsylum(t, args...)
	if r.exitCode != 0 {
		t.Fatalf("asylum %v exited %d\nstdout: %s\nstderr: %s", args, r.exitCode, r.stdout, r.stderr)
	}
	return r
}

// --- Tests ---

func TestHelp(t *testing.T) {
	r := runAsylumSuccess(t, "--help")
	if !strings.Contains(r.stdout, "Usage:") {
		t.Errorf("help output should contain 'Usage:', got:\n%s", r.stdout)
	}
}

func TestVersion(t *testing.T) {
	r := runAsylumSuccess(t, "--version")
	if !strings.Contains(r.stdout, "asylum") {
		t.Errorf("version output should contain 'asylum', got:\n%s", r.stdout)
	}
}

func TestRunMode(t *testing.T) {
	r := runAsylumSuccess(t, "run", "echo", "ok")
	if !strings.Contains(r.stdout, "ok") {
		t.Errorf("run output should contain 'ok', got:\n%s", r.stdout)
	}
}

func TestRunModeReusesImage(t *testing.T) {
	// Second run should be faster (image cached)
	start := time.Now()
	r := runAsylumSuccess(t, "run", "echo", "cached")
	elapsed := time.Since(start)
	if !strings.Contains(r.stdout, "cached") {
		t.Errorf("output should contain 'cached', got:\n%s", r.stdout)
	}
	// Should be under 30s if image is cached (first build might be 60s+)
	if elapsed > 30*time.Second {
		t.Logf("warning: second run took %s (expected <30s with cached image)", elapsed)
	}
}

func TestContainerCleanedUp(t *testing.T) {
	runAsylumSuccess(t, "run", "echo", "cleanup-test")

	// Give a moment for cleanup
	time.Sleep(time.Second)

	// No asylum containers should be running for this project
	out, _ := exec.Command("docker", "ps", "--filter", "name=asylum-", "--format", "{{.Names}}").Output()
	names := strings.TrimSpace(string(out))
	if names != "" {
		t.Errorf("containers still running after exit: %s", names)
	}
}
