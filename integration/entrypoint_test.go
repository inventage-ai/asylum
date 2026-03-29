//go:build integration

package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEntrypointDefaults(t *testing.T) {
	ensureBaseImage(t)

	// Single container run to verify all default entrypoint behaviors
	script := strings.Join([]string{
		`echo "SAFE_DIR=$(git config --global --get-all safe.directory)"`,
		`echo "GIT_EMAIL=$(git config --global user.email)"`,
		`echo "NODE=$(node --version)"`,
		`echo "PYTHON=$(python3 --version)"`,
		`echo "UV=$(uv --version)"`,
		`echo "JAVA=$(java -version 2>&1 | head -1)"`,
	}, " && ")

	out := dockerRun(t, script)
	results := parseKeyValues(out)

	if v := results["SAFE_DIR"]; v != "*" {
		t.Errorf("safe.directory = %q, want \"*\"", v)
	}
	if v := results["GIT_EMAIL"]; v != "claude@asylum" {
		t.Errorf("user.email = %q, want \"claude@asylum\"", v)
	}
	if v := results["NODE"]; !strings.HasPrefix(v, "v") {
		t.Errorf("node --version = %q, want v*", v)
	}
	if v := results["PYTHON"]; !strings.HasPrefix(v, "Python 3") {
		t.Errorf("python3 --version = %q, want Python 3.*", v)
	}
	if v := results["UV"]; !strings.HasPrefix(v, "uv ") {
		t.Errorf("uv --version = %q, want uv *", v)
	}
	if v := results["JAVA"]; !strings.Contains(v, "openjdk") {
		t.Errorf("java -version missing openjdk: %s", v)
	}
}

func TestEntrypointHostGitConfig(t *testing.T) {
	ensureBaseImage(t)
	tmp := t.TempDir()
	gitconfig := filepath.Join(tmp, "gitconfig")
	if err := os.WriteFile(gitconfig, []byte("[user]\n    email = test@example.com\n    name = Test User\n"), 0644); err != nil {
		t.Fatal(err)
	}
	out := dockerRunWithVolume(t, gitconfig+":/tmp/host_gitconfig:ro", "git config --global user.email")
	if out != "test@example.com" {
		t.Fatalf("user.email = %q, want \"test@example.com\"", out)
	}
}

func TestEntrypointJavaVersionSelection(t *testing.T) {
	ensureBaseImage(t)
	for _, ver := range []string{"17", "21", "25"} {
		t.Run("java"+ver, func(t *testing.T) {
			env := map[string]string{"ASYLUM_JAVA_VERSION": ver}
			out := dockerRunWithEnv(t, env, "java -version 2>&1")
			if !strings.Contains(out, ver) {
				t.Fatalf("java -version with ASYLUM_JAVA_VERSION=%s: %s", ver, out)
			}
		})
	}
}

func TestEntrypointPythonVenvCreation(t *testing.T) {
	ensureBaseImage(t)
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "requirements.txt"), []byte("requests\n"), 0644); err != nil {
		t.Fatal(err)
	}
	out := dockerRunWithProjectDir(t, tmp, "test -d .venv && echo VENV=yes || echo VENV=no")
	if !strings.Contains(out, "VENV=yes") {
		t.Fatalf("venv not created: %s", out)
	}
}

func TestEntrypointSSHPermissions(t *testing.T) {
	ensureBaseImage(t)
	tmp := t.TempDir()
	sshDir := filepath.Join(tmp, ".ssh")
	if err := os.MkdirAll(sshDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sshDir, "id_rsa"), []byte("private"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sshDir, "id_rsa.pub"), []byte("public"), 0644); err != nil {
		t.Fatal(err)
	}

	home, _ := os.UserHomeDir()
	sshDst := filepath.Join(home, ".ssh")
	script := `stat -c '%a' $HOME/.ssh && stat -c '%a' $HOME/.ssh/id_rsa && stat -c '%a' $HOME/.ssh/id_rsa.pub`
	out := dockerRunWithVolume(t, sshDir+":"+sshDst, script)
	lines := strings.Split(out, "\n")
	if len(lines) < 3 {
		t.Fatalf("unexpected output: %s", out)
	}
	if lines[0] != "700" {
		t.Errorf(".ssh dir perms = %s, want 700", lines[0])
	}
	if lines[1] != "600" {
		t.Errorf("id_rsa perms = %s, want 600", lines[1])
	}
	if lines[2] != "644" {
		t.Errorf("id_rsa.pub perms = %s, want 644", lines[2])
	}
}

func TestEntrypointEnvironment(t *testing.T) {
	ensureBaseImage(t)
	out := dockerRunWithEnv(t, map[string]string{"ASYLUM_TEST_VAR": "hello"}, "echo $ASYLUM_TEST_VAR")
	if out != "hello" {
		t.Fatalf("ASYLUM_TEST_VAR = %q, want \"hello\"", out)
	}
}

func TestEntrypointClaudeConfigRestore(t *testing.T) {
	ensureBaseImage(t)
	home, _ := os.UserHomeDir()
	tmp := t.TempDir()
	backupDir := filepath.Join(tmp, "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		t.Fatal(err)
	}
	backup := `{"oauthAccount":"test","hasCompletedOnboarding":true}`
	if err := os.WriteFile(filepath.Join(backupDir, ".claude.json.backup.1000"), []byte(backup), 0644); err != nil {
		t.Fatal(err)
	}

	out := dockerRunWithVolumeAndEnv(t, tmp+":"+filepath.Join(home, ".claude"),
		map[string]string{"CLAUDE_CONFIG_DIR": filepath.Join(home, ".claude")},
		"cat $HOME/.claude/.claude.json")
	if !strings.Contains(out, `"oauthAccount"`) {
		t.Fatalf("expected restored config with oauthAccount, got: %s", out)
	}
}

func TestEntrypointClaudeConfigRestoreSkipsWithoutAuth(t *testing.T) {
	ensureBaseImage(t)
	home, _ := os.UserHomeDir()
	tmp := t.TempDir()
	backupDir := filepath.Join(tmp, "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(backupDir, ".claude.json.backup.1000"), []byte(`{"minimal":true}`), 0644); err != nil {
		t.Fatal(err)
	}

	out := dockerRunWithVolumeAndEnv(t, tmp+":"+filepath.Join(home, ".claude"),
		map[string]string{"CLAUDE_CONFIG_DIR": filepath.Join(home, ".claude")},
		"test -f $HOME/.claude/.claude.json && echo EXISTS || echo MISSING")
	if !strings.Contains(out, "MISSING") {
		t.Fatalf("expected MISSING when backup has no auth, got: %s", out)
	}
}

func TestEntrypointClaudeConfigNoRestoreWhenPresent(t *testing.T) {
	ensureBaseImage(t)
	home, _ := os.UserHomeDir()
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, ".claude.json"), []byte(`{"oauthAccount":"original"}`), 0644); err != nil {
		t.Fatal(err)
	}
	backupDir := filepath.Join(tmp, "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(backupDir, ".claude.json.backup.1000"), []byte(`{"oauthAccount":"backup"}`), 0644); err != nil {
		t.Fatal(err)
	}

	out := dockerRunWithVolumeAndEnv(t, tmp+":"+filepath.Join(home, ".claude"),
		map[string]string{"CLAUDE_CONFIG_DIR": filepath.Join(home, ".claude")},
		"cat $HOME/.claude/.claude.json")
	if !strings.Contains(out, `"original"`) {
		t.Fatalf("expected original config preserved, got: %s", out)
	}
}

