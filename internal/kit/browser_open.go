package kit

import (
	"net/http"
	"net/url"
	"os/exec"
	"runtime"

	"github.com/inventage-ai/asylum/internal/broker"
)

func init() {
	Register(&Kit{
		Name:        "browser-open",
		Description: "Open URLs in the host browser",
		Tier:        TierAlwaysOn,
		ConfigSnippet: `  # browser-open:       # Open URLs in the host browser (on by default)
`,
		ConfigComment: "browser-open:         # Open URLs in the host browser (on by default)",
		// Install a shim that forwards a URL to the host broker, and expose it
		// under the names tools use to open a browser. /usr/local/bin precedes
		// /usr/bin so these shadow any distribution xdg-open.
		DockerSnippet: `# Open URLs in the host browser via the asylum host broker.
# Uses the Unix socket when present (native Linux), else loopback TCP via
# host.docker.internal (Docker Desktop / macOS).
RUN printf '%s\n' \
    '#!/bin/sh' \
    '[ -n "$1" ] || { echo "usage: asylum-open <url>" >&2; exit 1; }' \
    'if [ -n "$ASYLUM_BROKER_SOCK" ]; then' \
    '  exec curl -fsS -X POST -H "Authorization: Bearer ${ASYLUM_BROKER_TOKEN}" --data-urlencode "url=$1" --unix-socket "$ASYLUM_BROKER_SOCK" http://localhost/open' \
    'else' \
    '  exec curl -fsS -X POST -H "Authorization: Bearer ${ASYLUM_BROKER_TOKEN}" --data-urlencode "url=$1" "http://${ASYLUM_BROKER_HOST}:${ASYLUM_BROKER_PORT}/open"' \
    'fi' \
    | sudo tee /usr/local/bin/asylum-open >/dev/null && \
    sudo chmod +x /usr/local/bin/asylum-open && \
    sudo ln -sf /usr/local/bin/asylum-open /usr/local/bin/open && \
    sudo ln -sf /usr/local/bin/asylum-open /usr/local/bin/xdg-open && \
    sudo ln -sf /usr/local/bin/asylum-open /usr/local/bin/sensible-browser
`,
		EnvFunc: func(*SnippetConfig) map[string]string {
			return map[string]string{"BROWSER": "/usr/local/bin/asylum-open"}
		},
		RulesSnippet: `### Opening URLs (browser-open kit)
Run ` + "`open <url>`" + ` (or ` + "`xdg-open <url>`" + `) to open an http(s) URL in the user's browser on the host — for dev servers, previews, and login links. The user's full-screen terminal often blocks text selection, so prefer opening a URL over printing it for them to copy.
`,
		Routes: []broker.Route{{Path: "/open", Handler: openHandler}},
	})
}

// openHandler opens an http(s) URL in the host's default browser. It runs on
// the host. Only http/https is accepted; the URL is passed as a single argument
// to the opener with no shell, so metacharacters are inert.
func openHandler(w http.ResponseWriter, r *http.Request) {
	raw := r.FormValue("url")
	if !validBrowserURL(raw) {
		http.Error(w, "only http(s) URLs may be opened", http.StatusBadRequest)
		return
	}
	opener := "xdg-open"
	if runtime.GOOS == "darwin" {
		opener = "open"
	}
	if err := exec.Command(opener, raw).Start(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// validBrowserURL reports whether raw is a well-formed http(s) URL with a host.
// This blocks file://, application launches, and shell metacharacters.
func validBrowserURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}
