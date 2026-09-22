package kit

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sync"

	"github.com/inventage-ai/asylum/internal/broker"
	"github.com/inventage-ai/asylum/internal/log"
)

const itermKitName = "iterm"

// itermHookRel is the path iTerm2 writes into the user's Claude settings for
// its status-bar integration, relative to home. iTerm2 maintains it as a
// symlink into its app bundle, so it survives iTerm2 updates. The container
// gets asylum's shim at the same path, leaving the host binary in place for
// host sessions.
const itermHookRel = ".config/iterm2/cc-status"

// itermUtilitiesDir holds it2, which cc-status execs through /usr/bin/env. The
// broker inherits its PATH from whichever process spawned it, so the handler
// puts this directory on the child's PATH rather than trusting that.
const itermUtilitiesDir = "/Applications/iTerm.app/Contents/Resources/utilities"

const (
	itermRoute         = "/iterm-status"
	itermSessionHeader = "X-Asylum-Session"
	itermMaxBody       = 64 << 10
)

// itermSession is the sealed payload: which terminal session the container's
// status updates belong to. The container never sees it in the clear.
//
// cc-status reads TERM_SESSION_ID, not ITERM_SESSION_ID, and derives the bare
// UUID it2 wants by stripping everything through the colon.
type itermSession struct {
	TermSessionID string `json:"term_session_id"`
}

func init() {
	Register(&Kit{
		Name:        itermKitName,
		Description: "Drive the iTerm2 status bar from a containerized session",
		Tier:        TierOptIn,
		ConfigSnippet: `  # iterm:              # iTerm2 status bar integration (macOS)
`,
		ConfigNodes:   configNodes(itermKitName, "iTerm2 status bar integration (macOS)", nil),
		ConfigComment: "iterm:                # iTerm2 status bar integration (macOS)",
		MountFunc:     itermMountFunc,
		RulesSnippet: `### iTerm2 status bar (iterm kit)
Session state is reported to the iTerm2 status bar of the terminal this session runs in — a colored dot for working or idle, and the name of the running tool. It is driven by the hooks iTerm2 already configured; nothing needs to be run by hand.
`,
		Routes: []broker.Route{{Path: itermRoute, Handler: itermStatusHandler}},
	})
}

// itermShim replaces the host's Mach-O status binary inside the container. It
// forwards the hook payload and exits 0 no matter what: Claude Code reads hook
// exit codes, and a status indicator must never be able to fail the action it
// is reporting on.
const itermShim = `#!/bin/sh
# asylum: forwards iTerm2 status hooks to the host broker.
[ -n "$ASYLUM_SESSION" ] || exit 0
if [ -n "$ASYLUM_BROKER_SOCK" ]; then
    curl -fsS --max-time 2 -X POST \
        -H "Authorization: Bearer ${ASYLUM_BROKER_TOKEN}" \
        -H "X-Asylum-Session: ${ASYLUM_SESSION}" \
        --data-binary @- --unix-socket "$ASYLUM_BROKER_SOCK" \
        http://localhost/iterm-status >/dev/null 2>&1
else
    curl -fsS --max-time 2 -X POST \
        -H "Authorization: Bearer ${ASYLUM_BROKER_TOKEN}" \
        -H "X-Asylum-Session: ${ASYLUM_SESSION}" \
        --data-binary @- \
        "http://${ASYLUM_BROKER_HOST}:${ASYLUM_BROKER_PORT}/iterm-status" >/dev/null 2>&1
fi
exit 0
`

// itermMountFunc stages the shim and mounts it over the hook path. It writes
// the file itself rather than returning Content, because staged content is
// written 0600 and the hook has to be executable.
//
// No host binary means the integration is not installed (or this is not a Mac),
// so there is no hook to shadow and the kit contributes nothing.
func itermMountFunc(opts CredentialOpts) ([]CredentialMount, error) {
	if !fileExists(filepath.Join(opts.HomeDir, itermHookRel)) {
		return nil, nil
	}
	dir := filepath.Join(opts.HomeDir, ".asylum", "projects", opts.ContainerName)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	shim := filepath.Join(dir, "cc-status-shim")
	if err := os.WriteFile(shim, []byte(itermShim), 0755); err != nil {
		return nil, err
	}
	return []CredentialMount{{HostPath: shim, Destination: "~/" + itermHookRel}}, nil
}

// ItermSessionEnv returns the per-session environment a starting session needs:
// the sealed identity of the iTerm2 session it was launched from. It returns
// nothing when the kit is inactive or the terminal is not iTerm2, which is what
// disables the integration outside iTerm2 without a config switch.
func ItermSessionEnv(kits []*Kit) map[string]string {
	if !slices.ContainsFunc(kits, func(k *Kit) bool { return k.Name == itermKitName }) {
		return nil
	}
	// ITERM_SESSION_ID marks the terminal as iTerm2; TERM_SESSION_ID is the
	// value cc-status actually reads, and Terminal.app sets it too.
	if os.Getenv("ITERM_SESSION_ID") == "" {
		return nil
	}
	id := os.Getenv("TERM_SESSION_ID")
	if id == "" {
		return nil
	}
	blob, err := broker.Seal(itermSession{TermSessionID: id})
	if err != nil {
		log.Warn("iterm: %v", err)
		return nil
	}
	return map[string]string{broker.SessionEnv: blob}
}

// itermLocks serializes updates per terminal session. Hook events describe a
// sequence of transitions, and applying them out of order leaves the status bar
// showing a state the session is not in.
var (
	itermLocksMu sync.Mutex
	itermLocks   = map[string]*sync.Mutex{}
)

func itermLock(sessionID string) *sync.Mutex {
	itermLocksMu.Lock()
	defer itermLocksMu.Unlock()
	m, ok := itermLocks[sessionID]
	if !ok {
		m = &sync.Mutex{}
		itermLocks[sessionID] = m
	}
	return m
}

// itermStatusHandler runs the host's own status binary on the container's
// behalf. It runs on the host.
//
// Nothing the container sends becomes an argument: the executable path is
// fixed, the argument list is empty, and the payload reaches the binary only on
// stdin. The alternative — exposing it2 — would hand the container
// `it2 session run`, which executes commands in every terminal the user has
// open.
func itermStatusHandler(_ broker.Ctx, w http.ResponseWriter, r *http.Request) {
	var session itermSession
	err := broker.Open(r.Header.Get(itermSessionHeader), &session)
	switch {
	case errors.Is(err, broker.ErrNoEnvelope):
		http.Error(w, "no session envelope", http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, "session envelope could not be opened", http.StatusBadRequest)
		return
	case session.TermSessionID == "":
		http.Error(w, "session envelope names no terminal session", http.StatusBadRequest)
		return
	}

	payload, err := io.ReadAll(http.MaxBytesReader(w, r.Body, itermMaxBody))
	if err != nil {
		http.Error(w, "hook payload too large", http.StatusRequestEntityTooLarge)
		return
	}

	home, err := os.UserHomeDir()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	statusBin := filepath.Join(home, itermHookRel)
	if !fileExists(statusBin) {
		http.Error(w, "iTerm2 Claude Code integration is not installed on the host", http.StatusServiceUnavailable)
		return
	}

	lock := itermLock(session.TermSessionID)
	lock.Lock()
	defer lock.Unlock()

	cmd := exec.Command(statusBin)
	cmd.Stdin = bytes.NewReader(payload)
	// Both are overridden, not just the one cc-status reads today: the broker
	// inherited this session-scoped pair from whichever tab spawned it, and a
	// stale leftover is what made every session paint that tab.
	cmd.Env = append(os.Environ(),
		"TERM_SESSION_ID="+session.TermSessionID,
		"ITERM_SESSION_ID="+session.TermSessionID,
		"PATH="+itermUtilitiesDir+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	if err := cmd.Run(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
