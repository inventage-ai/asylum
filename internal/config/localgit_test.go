package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// gitRepo creates a repository isolated from the user's git config, so a
// global excludes file cannot change the result.
func gitRepo(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	dir := t.TempDir()
	runGit(t, dir, "init", "-q")
	return dir
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir, "-c", "user.name=t", "-c", "user.email=t@t"}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestCheckUntracked(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		setup   func(t *testing.T, repo, dir string)
		wantErr bool
	}{
		{"untracked", ".asylum.local", func(*testing.T, string, string) {}, false},
		{"committed", ".asylum.local", func(t *testing.T, repo, dir string) {
			runGit(t, repo, "add", "-f", ".")
			runGit(t, repo, "commit", "-qm", "x")
		}, true},
		{"staged alias", ".asylum.local.yaml", func(t *testing.T, repo, dir string) {
			runGit(t, repo, "add", "-f", ".")
		}, true},
	}
	for _, tt := range tests {
		for _, sub := range []string{"", "sub/project"} {
			t.Run(tt.name+"/"+sub, func(t *testing.T) {
				repo := gitRepo(t)
				dir := filepath.Join(repo, sub)
				os.MkdirAll(dir, 0755)
				path := filepath.Join(dir, tt.file)
				os.WriteFile(path, []byte("agent: codex\n"), 0644)
				tt.setup(t, repo, dir)
				err := checkUntracked(path)
				if (err != nil) != tt.wantErr {
					t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
				}
				if err != nil && !strings.Contains(err.Error(), "git rm --cached "+tt.file) {
					t.Errorf("error does not show the fix: %v", err)
				}
			})
		}
	}
}

// On a case-insensitive filesystem a committed .ASYLUM.LOCAL is read as
// .asylum.local. A hard link reproduces that on a case-sensitive one.
func TestCheckUntrackedOtherName(t *testing.T) {
	repo := gitRepo(t)
	tracked := filepath.Join(repo, ".ASYLUM.LOCAL")
	os.WriteFile(tracked, []byte("volumes: [~/.ssh:/x]\n"), 0644)
	runGit(t, repo, "add", "-f", ".")
	path := filepath.Join(repo, ".asylum.local")
	if !fileExists(path) {
		if err := os.Link(tracked, path); err != nil {
			t.Fatal(err)
		}
	}
	err := checkUntracked(path)
	if err == nil || !strings.Contains(err.Error(), "git rm --cached .ASYLUM.LOCAL") {
		t.Errorf("err = %v, want refusal naming .ASYLUM.LOCAL", err)
	}
}

func TestCheckUntrackedOutsideRepo(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".asylum.local")
	os.WriteFile(path, []byte("agent: codex\n"), 0644)
	if err := checkUntracked(path); err != nil {
		t.Errorf("err = %v", err)
	}
}

func TestLoadRefusesTrackedLocalConfig(t *testing.T) {
	repo := gitRepo(t)
	t.Setenv("HOME", t.TempDir())
	os.WriteFile(filepath.Join(repo, ".asylum.local"), []byte("agent: codex\n"), 0644)
	runGit(t, repo, "add", ".asylum.local")
	if _, err := Load(repo, CLIFlags{}, ""); err == nil || !strings.Contains(err.Error(), "tracked by git") {
		t.Errorf("Load err = %v, want tracked error", err)
	}
}

func readExclude(t *testing.T, repo string) string {
	t.Helper()
	data, _ := os.ReadFile(filepath.Join(repo, ".git", "info", "exclude"))
	return string(data)
}

func TestExcludeLocalConfig(t *testing.T) {
	t.Run("not yet ignored", func(t *testing.T) {
		repo := gitRepo(t)
		os.WriteFile(filepath.Join(repo, ".git", "info", "exclude"), []byte("*.tmp"), 0644)
		added, err := ExcludeLocalConfig(repo)
		if err != nil || !slices.Equal(added, localConfigNames) {
			t.Fatalf("added = %v, err = %v", added, err)
		}
		if got := readExclude(t, repo); !strings.HasPrefix(got, "*.tmp\n") || !strings.HasSuffix(got, ".asylum.local\n.asylum.local.yaml\n") {
			t.Errorf("exclude = %q", got)
		}
		runGit(t, repo, "check-ignore", "-q", ".asylum.local")
		runGit(t, repo, "check-ignore", "-q", ".asylum.local.yaml")
		if added, _ := ExcludeLocalConfig(repo); added != nil {
			t.Errorf("second run added %v", added)
		}
	})
	t.Run("already ignored", func(t *testing.T) {
		repo := gitRepo(t)
		os.WriteFile(filepath.Join(repo, ".gitignore"), []byte(".asylum.local*\n"), 0644)
		before := readExclude(t, repo)
		if added, err := ExcludeLocalConfig(repo); added != nil || err != nil {
			t.Errorf("added = %v, err = %v", added, err)
		}
		if readExclude(t, repo) != before {
			t.Error("exclude file modified")
		}
	})
	t.Run("only alias missing", func(t *testing.T) {
		repo := gitRepo(t)
		os.WriteFile(filepath.Join(repo, ".gitignore"), []byte(".asylum.local\n"), 0644)
		if added, _ := ExcludeLocalConfig(repo); !slices.Equal(added, []string{".asylum.local.yaml"}) {
			t.Errorf("added = %v", added)
		}
	})
	t.Run("missing info directory", func(t *testing.T) {
		repo := gitRepo(t)
		os.RemoveAll(filepath.Join(repo, ".git", "info"))
		if _, err := ExcludeLocalConfig(repo); err != nil {
			t.Fatal(err)
		}
		runGit(t, repo, "check-ignore", "-q", ".asylum.local")
	})
	t.Run("project subdirectory", func(t *testing.T) {
		repo := gitRepo(t)
		dir := filepath.Join(repo, "sub")
		os.MkdirAll(dir, 0755)
		if _, err := ExcludeLocalConfig(dir); err != nil {
			t.Fatal(err)
		}
		runGit(t, dir, "check-ignore", "-q", ".asylum.local")
	})
	t.Run("linked worktree", func(t *testing.T) {
		repo := gitRepo(t)
		runGit(t, repo, "commit", "-q", "--allow-empty", "-m", "x")
		wt := filepath.Join(t.TempDir(), "wt")
		runGit(t, repo, "worktree", "add", "-q", wt)
		if added, err := ExcludeLocalConfig(wt); err != nil || added == nil {
			t.Fatalf("added = %v, err = %v", added, err)
		}
		if !strings.Contains(readExclude(t, repo), ".asylum.local") {
			t.Error("entries not written to the shared repository")
		}
		runGit(t, wt, "check-ignore", "-q", ".asylum.local")
	})
	t.Run("not a repository", func(t *testing.T) {
		dir := t.TempDir()
		if added, err := ExcludeLocalConfig(dir); added != nil || err != nil {
			t.Errorf("added = %v, err = %v", added, err)
		}
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			t.Error(".git created")
		}
	})
}
