package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/inventage-ai/asylum/internal/config"
	"github.com/inventage-ai/asylum/internal/kit"
	"github.com/inventage-ai/asylum/internal/log"
	"github.com/inventage-ai/asylum/internal/term"
	"github.com/inventage-ai/asylum/internal/tui"
)

// confirmRejected handles config entries that were dropped because their file
// does not accept them. In a terminal it asks once per set of entries whether
// to start without them; without one it warns and continues, which is safe
// because dropping entries only narrows the sandbox.
func confirmRejected(rejected []config.Rejection, stateFile string) {
	if len(rejected) == 0 {
		return
	}
	hash := rejectedHash(rejected)
	if prev, err := os.ReadFile(stateFile); err == nil && strings.TrimSpace(string(prev)) == hash {
		var entries []string
		for _, r := range rejected {
			entries = append(entries, fmt.Sprintf("%s (%s:%d)", r.Path, r.File, r.Line))
		}
		log.Warn("ignoring config options their file does not accept: %s", strings.Join(entries, ", "))
		return
	}
	warnRejected(rejected)
	if !term.IsTerminal() {
		return
	}
	choice, err := tui.Select("The config options above are ignored", []tui.Option{
		{Label: "Start without them", Description: "Asks again when the ignored options change"},
		{Label: "Abort"},
	}, 0)
	if errors.Is(err, tui.ErrCancelled) || (err == nil && choice != 0) {
		die("aborted")
	}
	if err != nil {
		return // no usable terminal after all: treat as non-interactive
	}
	if err := os.MkdirAll(filepath.Dir(stateFile), 0755); err != nil {
		log.Warn("remember decision: %v", err)
		return
	}
	if err := os.WriteFile(stateFile, []byte(hash+"\n"), 0644); err != nil {
		log.Warn("remember decision: %v", err)
	}
}

// warnRejected logs each dropped entry with the YAML to move and where to.
func warnRejected(rejected []config.Rejection) {
	for _, r := range rejected {
		reason := "only personal config files may widen the sandbox"
		if r.Scope == kit.ScopeGlobal {
			reason = "it configures asylum itself, not a single project"
		}
		log.Warn("%s:%d: %s is ignored here (%s). Move it to %s:\n%s",
			r.File, r.Line, r.Path, reason, r.AcceptedIn(), indent(r.Snippet, "    "))
	}
}

// rejectedHash identifies a set of dropped entries. Line numbers are left out,
// so editing unrelated parts of a file does not ask again.
func rejectedHash(rejected []config.Rejection) string {
	entries := make([]string, len(rejected))
	for i, r := range rejected {
		entries[i] = r.File + "\x00" + r.Path + "\x00" + r.Snippet
	}
	slices.Sort(entries)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(entries, "\n"))))
}

func indent(s, prefix string) string {
	return prefix + strings.ReplaceAll(s, "\n", "\n"+prefix)
}
