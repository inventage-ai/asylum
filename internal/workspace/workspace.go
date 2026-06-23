// Package workspace guards against sandboxing directories that cannot be
// safely bind-mounted (the home directory or filesystem root) by redirecting
// to a freshly created dated workspace under ~/asylum-workspace/.
package workspace

import (
	"bufio"
	"bytes"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/inventage-ai/asylum/assets"
)

var words = parseWords(assets.WorkspaceWords)

func parseWords(data []byte) []string {
	var out []string
	sc := bufio.NewScanner(bytes.NewReader(data))
	for sc.Scan() {
		if w := strings.TrimSpace(sc.Text()); w != "" {
			out = append(out, w)
		}
	}
	return out
}

// unsafe reports whether dir cannot be safely sandboxed: the exact home
// directory or the filesystem root. Subdirectories of home are safe.
func unsafe(dir, home string) bool {
	dir = filepath.Clean(dir)
	return dir == filepath.Clean(home) || dir == string(filepath.Separator)
}

func name(date string, r *rand.Rand) string {
	w := func() string { return words[r.Intn(len(words))] }
	return date + "-" + w() + "-" + w() + "-" + w()
}

// Resolve returns projectDir unchanged when it is safe to sandbox. When it is
// the home directory or filesystem root, it creates a fresh workspace under
// ~/asylum-workspace/<YYYY-MM-DD>-<three-words>/ and returns that path with
// redirected=true.
func Resolve(projectDir, home string) (string, bool, error) {
	if !unsafe(projectDir, home) {
		return projectDir, false, nil
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return resolveAt(filepath.Join(home, "asylum-workspace"), time.Now().Format("2006-01-02"), r)
}

func resolveAt(base, date string, r *rand.Rand) (string, bool, error) {
	for {
		dir := filepath.Join(base, name(date, r))
		_, err := os.Stat(dir)
		if err == nil {
			continue // collision — re-roll the words
		}
		if !os.IsNotExist(err) {
			return "", false, err
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", false, err
		}
		return dir, true, nil
	}
}
