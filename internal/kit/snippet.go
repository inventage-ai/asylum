package kit

import "strings"

// Kits author ConfigSnippet in either active or commented form, matching the
// tier they default to. Both the first-run writer and the kit-sync prompt have
// to be able to flip that form when the user's choice differs from the
// authored default, so the transformations live here beside the snippets.

// SnippetIsActive reports whether the snippet's first non-blank line is
// uncommented.
func SnippetIsActive(snippet string) bool {
	for _, line := range strings.Split(snippet, "\n") {
		t := strings.TrimLeft(line, " ")
		if t == "" {
			continue
		}
		return !strings.HasPrefix(t, "#")
	}
	return false
}

// UncommentSnippet strips a single `# ` (or `#` alone) following the leading
// indent on each line. Blank lines are preserved unchanged. Lines without a
// `#` after the indent are returned as-is so inner commented details inside an
// otherwise-active snippet survive correctly.
func UncommentSnippet(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		idx := leadingSpaceLen(line)
		rest := line[idx:]
		switch {
		case rest == "":
			// blank line, leave it
		case strings.HasPrefix(rest, "# "):
			lines[i] = line[:idx] + rest[2:]
		case rest == "#":
			lines[i] = line[:idx]
		}
	}
	return strings.Join(lines, "\n")
}

// CommentSnippet inserts a `# ` after the leading indent on each non-blank line
// that isn't already a comment. Lines already commented at the YAML level are
// left alone — the whole block ends up commented either way, and skipping them
// avoids `# # versions:` artifacts when an authored active snippet contains
// inner hint-comment lines.
func CommentSnippet(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		idx := leadingSpaceLen(line)
		if idx == len(line) {
			continue // blank
		}
		if strings.HasPrefix(line[idx:], "#") {
			continue
		}
		lines[i] = line[:idx] + "# " + line[idx:]
	}
	return strings.Join(lines, "\n")
}

func leadingSpaceLen(s string) int {
	i := 0
	for i < len(s) && s[i] == ' ' {
		i++
	}
	return i
}
