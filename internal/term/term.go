package term

import (
	"io"
	"os"
	"strings"
)

// IsTerminal reports whether stdin is connected to a terminal.
func IsTerminal() bool { return IsTTY(os.Stdin) }

// IsTTY reports whether f is a terminal.
func IsTTY(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// ShellQuote wraps s in single quotes, escaping embedded single quotes
// so the result is safe for use in a shell command.
func ShellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// ShellQuoteArgs applies ShellQuote to each element of args.
func ShellQuoteArgs(args []string) []string {
	quoted := make([]string, len(args))
	for i, a := range args {
		quoted[i] = ShellQuote(a)
	}
	return quoted
}

// RestoreModes turns off terminal modes a TUI inside the container may have
// left on, because a stopped container gives the TUI no chance to clean up.
// The leading cursor save makes leaving the alternate screen a no-op when the
// terminal is already on the main screen.
func RestoreModes() {
	if !IsTTY(os.Stdout) {
		return
	}
	io.WriteString(os.Stdout, "\x1b7\x1b[?1000l\x1b[?1002l\x1b[?1003l\x1b[?1006l\x1b[?1004l\x1b[?1l\x1b[?2004l\x1b[?1049l\x1b[?25h")
}
