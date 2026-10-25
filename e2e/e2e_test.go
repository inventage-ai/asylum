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

	"github.com/inventage-ai/asylum/internal/config"
	"github.com/inventage-ai/asylum/internal/docker"
	"github.com/inventage-ai/asylum/internal/kit"
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

	// Use a stable HOME path so Docker build cache persists across runs.
	// A random temp dir would change the Dockerfile's USER_HOME build arg
	// each time, invalidating the Docker layer cache from user creation onward.
	testHome = "/tmp/asylum-e2e-home"
	os.RemoveAll(testHome)
	os.MkdirAll(testHome, 0755)
	configDir := filepath.Join(testHome, ".asylum")
	os.MkdirAll(configDir, 0755)
	os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(`version: "0.2"
agent: echo
kits: {}
agents: {}
`), 0644)

	// Mark every kit as known so kit sync does not try to rewrite the config.
	config.SaveState(configDir, config.State{KnownKits: kit.All()})

	// Create agent config dir so EnsureAgentConfig doesn't prompt
	os.MkdirAll(filepath.Join(configDir, "agents", "echo"), 0755)

	// Stable project dir for deterministic container names and Docker cache.
	projectDir = "/tmp/asylum-e2e-project"
	os.RemoveAll(projectDir)
	os.MkdirAll(projectDir, 0755)

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

// --- CLI validation tests (no Docker needed) ---

func TestHelp(t *testing.T) {
	r := runAsylumSuccess(t, "--help")
	if !strings.Contains(r.stdout, "Usage:") {
		t.Errorf("help output should contain 'Usage:', got:\n%s", r.stdout)
	}
}

func TestHelpAlias(t *testing.T) {
	r := runAsylumSuccess(t, "-h")
	if !strings.Contains(r.stdout, "Usage:") {
		t.Errorf("-h output should contain 'Usage:', got:\n%s", r.stdout)
	}
}

func TestVersion(t *testing.T) {
	r := runAsylumSuccess(t, "--version")
	if !strings.Contains(r.stdout, "asylum") {
		t.Errorf("version output should contain 'asylum', got:\n%s", r.stdout)
	}
}

func TestVersionSubcommand(t *testing.T) {
	r := runAsylumSuccess(t, "version")
	if !strings.Contains(r.stdout, "asylum") {
		t.Errorf("version subcommand should contain 'asylum', got:\n%s", r.stdout)
	}
}

func TestVersionShort(t *testing.T) {
	r := runAsylumSuccess(t, "version", "--short")
	if strings.Contains(r.stdout, "asylum") {
		t.Errorf("short version should not contain 'asylum', got:\n%s", r.stdout)
	}
	if strings.TrimSpace(r.stdout) == "" {
		t.Error("short version should not be empty")
	}
}

func TestVersionFlagShort(t *testing.T) {
	r := runAsylumSuccess(t, "--version", "--short")
	if strings.Contains(r.stdout, "asylum") {
		t.Errorf("--version --short should not contain 'asylum', got:\n%s", r.stdout)
	}
}

func TestUnknownFlag(t *testing.T) {
	r := runAsylum(t, "--nonexistent")
	if r.exitCode == 0 {
		t.Error("unknown flag should exit non-zero")
	}
	if !strings.Contains(r.stderr, "unknown flag") {
		t.Errorf("stderr should mention 'unknown flag', got:\n%s", r.stderr)
	}
}

func TestUnexpectedArgument(t *testing.T) {
	r := runAsylum(t, "notacommand")
	if r.exitCode == 0 {
		t.Error("unexpected argument should exit non-zero")
	}
	if !strings.Contains(r.stderr, "unexpected argument") {
		t.Errorf("stderr should mention 'unexpected argument', got:\n%s", r.stderr)
	}
}

func TestRunNoCommand(t *testing.T) {
	r := runAsylum(t, "run")
	if r.exitCode == 0 {
		t.Error("run with no command should exit non-zero")
	}
	if !strings.Contains(r.stderr, "requires a command") {
		t.Errorf("stderr should mention 'requires a command', got:\n%s", r.stderr)
	}
}

// --- Run mode tests (Docker) ---

func TestRunMode(t *testing.T) {
	r := runAsylumSuccess(t, "run", "echo", "ok")
	if !strings.Contains(r.stdout, "ok") {
		t.Errorf("run output should contain 'ok', got:\n%s", r.stdout)
	}
}

func TestRunModeReusesImage(t *testing.T) {
	start := time.Now()
	r := runAsylumSuccess(t, "run", "echo", "cached")
	elapsed := time.Since(start)
	if !strings.Contains(r.stdout, "cached") {
		t.Errorf("output should contain 'cached', got:\n%s", r.stdout)
	}
	if elapsed > 30*time.Second {
		t.Logf("warning: second run took %s (expected <30s with cached image)", elapsed)
	}
}

func TestRunModeExitCode(t *testing.T) {
	r := runAsylum(t, "run", "false")
	if r.exitCode == 0 {
		t.Error("'run false' should exit non-zero")
	}
}

func TestRunModeMultipleCommands(t *testing.T) {
	r := runAsylumSuccess(t, "run", "sh", "-c", "echo foo && echo bar")
	if !strings.Contains(r.stdout, "foo") || !strings.Contains(r.stdout, "bar") {
		t.Errorf("output should contain 'foo' and 'bar', got:\n%s", r.stdout)
	}
}

func TestRunModeWorkingDir(t *testing.T) {
	r := runAsylumSuccess(t, "run", "pwd")
	if !strings.Contains(r.stdout, projectDir) {
		t.Errorf("working dir should be %s, got:\n%s", projectDir, r.stdout)
	}
}

func TestRunModeReadFile(t *testing.T) {
	testFile := filepath.Join(projectDir, "e2e-test-read.txt")
	os.WriteFile(testFile, []byte("e2e-file-content"), 0644)
	defer os.Remove(testFile)

	r := runAsylumSuccess(t, "run", "cat", testFile)
	if !strings.Contains(r.stdout, "e2e-file-content") {
		t.Errorf("should read file from project dir, got:\n%s", r.stdout)
	}
}

// --- Agent mode tests (Docker) ---

func TestAgentMode(t *testing.T) {
	r := runAsylumSuccess(t)
	if r.exitCode != 0 {
		t.Errorf("agent mode with echo should succeed, got exit %d", r.exitCode)
	}
}

func TestAgentModeWithArgs(t *testing.T) {
	r := runAsylumSuccess(t, "--", "hello", "world")
	if !strings.Contains(r.stdout, "hello world") {
		t.Errorf("agent should pass args to echo, got:\n%s", r.stdout)
	}
}

func TestAgentFlag(t *testing.T) {
	r := runAsylumSuccess(t, "-a", "echo", "--", "agent-flag-test")
	if !strings.Contains(r.stdout, "agent-flag-test") {
		t.Errorf("output should contain 'agent-flag-test', got:\n%s", r.stdout)
	}
}

func TestAgentFlagShort(t *testing.T) {
	r := runAsylumSuccess(t, "-aecho", "--", "short-flag-test")
	if !strings.Contains(r.stdout, "short-flag-test") {
		t.Errorf("output should contain 'short-flag-test', got:\n%s", r.stdout)
	}
}

func TestNewSessionFlag(t *testing.T) {
	r := runAsylumSuccess(t, "--new", "--", "new-session-test")
	if !strings.Contains(r.stdout, "new-session-test") {
		t.Errorf("output should contain 'new-session-test', got:\n%s", r.stdout)
	}
}

func TestSkipOnboarding(t *testing.T) {
	r := runAsylumSuccess(t, "--skip-onboarding", "--", "skip-onboarding-test")
	if !strings.Contains(r.stdout, "skip-onboarding-test") {
		t.Errorf("output should contain 'skip-onboarding-test', got:\n%s", r.stdout)
	}
}

// --- Environment variable tests (Docker) ---

func TestEnvVar(t *testing.T) {
	r := runAsylumSuccess(t, "-e", "TEST_E2E_VAR=hello123", "run", "sh", "-c", "echo $TEST_E2E_VAR")
	if !strings.Contains(r.stdout, "hello123") {
		t.Errorf("should see env var value, got:\n%s", r.stdout)
	}
}

func TestMultipleEnvVars(t *testing.T) {
	r := runAsylumSuccess(t, "-e", "E2E_A=alpha", "-e", "E2E_B=beta", "run", "sh", "-c", "echo $E2E_A $E2E_B")
	if !strings.Contains(r.stdout, "alpha") || !strings.Contains(r.stdout, "beta") {
		t.Errorf("should see both env vars, got:\n%s", r.stdout)
	}
}

func TestEnvVarLastWins(t *testing.T) {
	r := runAsylumSuccess(t, "-e", "E2E_DUP=first", "-e", "E2E_DUP=second", "run", "sh", "-c", "echo $E2E_DUP")
	if !strings.Contains(r.stdout, "second") {
		t.Errorf("last-wins: expected 'second' in output, got:\n%s", r.stdout)
	}
}

// --- Volume and port tests (Docker) ---

func TestVolumeMount(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "asylum-e2e-vol-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	if err := os.WriteFile(filepath.Join(tmpDir, "vol-test.txt"), []byte("volume-data-xyz"), 0644); err != nil {
		t.Fatal(err)
	}

	r := runAsylumSuccess(t, "-v", tmpDir+":/mnt/testvol:ro", "run", "cat", "/mnt/testvol/vol-test.txt")
	if !strings.Contains(r.stdout, "volume-data-xyz") {
		t.Errorf("should read from volume mount, got:\n%s", r.stdout)
	}
}

func TestPortFlag(t *testing.T) {
	r := runAsylumSuccess(t, "-p", "19876", "run", "echo", "port-ok")
	if !strings.Contains(r.stdout, "port-ok") {
		t.Errorf("port flag should not break execution, got:\n%s", r.stdout)
	}
}

// --- Error cases ---

func TestInvalidAgent(t *testing.T) {
	r := runAsylum(t, "-a", "nonexistent", "run", "echo", "hi")
	if r.exitCode == 0 {
		t.Error("invalid agent should exit non-zero")
	}
	if !strings.Contains(r.stderr, "nonexistent") {
		t.Errorf("error should mention invalid agent name, got:\n%s", r.stderr)
	}
}

// --- Rebuild (Docker, slow) ---

func TestRebuildFlag(t *testing.T) {
	r := runAsylumSuccess(t, "--rebuild", "run", "echo", "rebuilt")
	if !strings.Contains(r.stdout, "rebuilt") {
		t.Errorf("output should contain 'rebuilt', got:\n%s", r.stdout)
	}
}

// --- Lifecycle tests (Docker) ---

func TestContainerCleanedUp(t *testing.T) {
	runAsylumSuccess(t, "run", "echo", "cleanup-test")
	time.Sleep(time.Second)

	out, _ := exec.Command("docker", "ps", "--filter", "name=asylum-", "--format", "{{.Names}}").Output()
	names := strings.TrimSpace(string(out))
	if names != "" {
		t.Errorf("containers still running after exit: %s", names)
	}
}

func TestCleanupProject(t *testing.T) {
	runAsylumSuccess(t, "run", "echo", "setup-for-cleanup")
	time.Sleep(time.Second)

	r := runAsylum(t, "cleanup")
	combined := r.stdout + r.stderr
	if strings.Contains(combined, "panic") {
		t.Errorf("cleanup should not panic:\n%s", combined)
	}
	if r.exitCode != 0 {
		t.Errorf("cleanup should exit 0, got %d\nstdout: %s\nstderr: %s", r.exitCode, r.stdout, r.stderr)
	}
}
