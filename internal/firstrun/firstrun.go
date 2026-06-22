package firstrun

import (
	"os"
	"path/filepath"
)

// IsFirstRun reports whether `~/.asylum/config.yaml` is absent — the signal
// that no prior asylum invocation has written its defaults. Capture this
// value before any code path that might create the file (notably
// `config.WriteDefaults`), otherwise the answer will always be false.
func IsFirstRun(home string) bool {
	_, err := os.Stat(filepath.Join(home, ".asylum", "config.yaml"))
	return os.IsNotExist(err)
}
