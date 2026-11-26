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
// ships .asylum.local could set options a committed .asylum may not. When git
// cannot answer (no repository, git not installed) the file counts as untracked.
func checkUntracked(path string) error {
	dir, name := filepath.Split(path)
	out, err := exec.Command("git", "-C", dir, "ls-files", "-z", "--", name).Output()
	if err != nil || len(out) == 0 {
		return nil
	}
	return fmt.Errorf("%s is tracked by git, but it holds personal settings that must not be shared.\n"+
		"Untrack it with `git rm --cached %s` and keep it ignored", path, name)
}

// ExcludeLocalConfig adds the local config filenames that git does not ignore
// yet to the repository's local exclude file, so they are not committed by
// accident. It returns the names it added; outside a git repository, or without
// git, it does nothing.
func ExcludeLocalConfig(projectDir string) ([]string, error) {
	// check-ignore exits 1 when none of the names is ignored; its output lists the ignored ones.
	out, err := exec.Command("git", append([]string{"-C", projectDir, "check-ignore", "--"}, localConfigNames...)...).Output()
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

	out, err = exec.Command("git", "-C", projectDir, "rev-parse", "--git-path", "info/exclude").Output()
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

func isExitCode(err error, code int) bool {
	exitErr, ok := err.(*exec.ExitError)
	return ok && exitErr.ExitCode() == code
}
