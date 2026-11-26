package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

var localConfigNames = []string{".asylum.local", ".asylum.local.yaml"}

// checkUntracked refuses a local config file that git tracks: a repository that
// ships .asylum.local could set options a committed .asylum may not. Tracked
// files are matched by identity, not by name, because a case-insensitive
// filesystem reads a committed .ASYLUM.LOCAL as .asylum.local. When git cannot
// answer (no repository, git not installed) the file counts as untracked.
func checkUntracked(path string) error {
	dir := filepath.Dir(path)
	out, err := git(dir, "ls-files", "-z", "--", ":(glob)*")
	if err != nil {
		return nil
	}
	local, err := os.Stat(path)
	if err != nil {
		return err
	}
	for name := range strings.SplitSeq(string(out), "\x00") {
		if tracked, err := os.Stat(filepath.Join(dir, name)); err == nil && os.SameFile(local, tracked) {
			return fmt.Errorf("%s is tracked by git, but it holds personal settings that must not be shared.\n"+
				"Untrack it with `git rm --cached %s` and keep it ignored", path, name)
		}
	}
	return nil
}

// ExcludeLocalConfig adds the local config filenames that git does not ignore
// yet to the repository's local exclude file, so they are not committed by
// accident. It returns the names it added; outside a git repository, or without
// git, it does nothing.
func ExcludeLocalConfig(projectDir string) ([]string, error) {
	// check-ignore exits 1 when none of the names is ignored; its output lists the ignored ones.
	out, err := git(projectDir, append([]string{"check-ignore", "--"}, localConfigNames...)...)
	if err != nil && !isExitCode(err, 1) {
		return nil, nil
	}
	ignored := strings.Fields(string(out))
	var missing []string
	for _, name := range localConfigNames {
		if !slices.Contains(ignored, name) {
			missing = append(missing, name)
		}
	}
	if len(missing) == 0 {
		return nil, nil
	}

	out, err = git(projectDir, "rev-parse", "--git-path", "info/exclude")
	if err != nil {
		return nil, nil
	}
	path := strings.TrimSpace(string(out))
	if !filepath.IsAbs(path) {
		path = filepath.Join(projectDir, path)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	entry := "# asylum: personal config, never commit\n" + strings.Join(missing, "\n") + "\n"
	if len(existing) > 0 && !strings.HasSuffix(string(existing), "\n") {
		entry = "\n" + entry
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if _, err := f.WriteString(entry); err != nil {
		return nil, err
	}
	return missing, nil
}

// git runs git in dir with fsmonitor switched off: it runs a command from the
// repository's config, which the agent can edit, and asylum runs on the host.
func git(dir string, args ...string) ([]byte, error) {
	return exec.Command("git", append([]string{"-C", dir, "-c", "core.fsmonitor=false"}, args...)...).Output()
}

func isExitCode(err error, code int) bool {
	exitErr, ok := err.(*exec.ExitError)
	return ok && exitErr.ExitCode() == code
}
