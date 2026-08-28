package kit

import (
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"

	"github.com/inventage-ai/asylum/internal/broker"
)

// OpenSchemesEnv names the environment variable that carries the extra URL
// schemes the /open handler accepts. The broker runs as a detached host process
// that never reads the project config, so the session passes the merged
// allowlist through its environment at spawn time.
const OpenSchemesEnv = "ASYLUM_OPEN_SCHEMES"

var schemeSyntax = regexp.MustCompile(`^[a-z][a-z0-9+.-]*$`)

// NormalizeSchemes splits configured scheme entries into the extra schemes the
// open handler may accept and the entries that had to be rejected. Matching is
// case-insensitive and entries may be written in any of the forms tools
// document their scheme in (dropshare5, dropshare5:, dropshare5://,
// dropshare5:///). http and https are always allowed and never appear in the
// extra list.
func NormalizeSchemes(entries []string) (schemes, rejected []string) {
	seen := map[string]bool{"http": true, "https": true}
	for _, entry := range entries {
		s := strings.ToLower(strings.TrimSpace(entry))
		s = strings.TrimSuffix(strings.TrimRight(s, "/"), ":")
		if !schemeSyntax.MatchString(s) {
			rejected = append(rejected, entry)
			continue
		}
		if seen[s] {
			continue
		}
		seen[s] = true
		schemes = append(schemes, s)
	}
	return schemes, rejected
}

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
		RulesSnippetFunc: openRulesSnippet,
		Routes:           []broker.Route{{Path: "/open", Handler: openHandler}},
	})
}

// openRulesSnippet describes the opener to the agent, naming the extra schemes
// the user allowlisted. sc is nil whenever the always-on kit has no config
// entry, which is the common case.
func openRulesSnippet(sc *SnippetConfig) string {
	rules := "### Opening URLs (browser-open kit)\n" +
		"Run `open <url>` (or `xdg-open <url>`) to open an http(s) URL in the user's browser on the host — for dev servers, previews, and login links. " +
		"The user's full-screen terminal often blocks text selection, so prefer opening a URL over printing it for them to copy.\n"
	if sc == nil {
		return rules
	}
	schemes, _ := NormalizeSchemes(sc.Schemes)
	if len(schemes) == 0 {
		return rules
	}
	var prefixed []string
	for _, s := range schemes {
		prefixed = append(prefixed, "`"+s+":`")
	}
	return rules + "The same command also hands these schemes to the app registered for them on the host: " + strings.Join(prefixed, ", ") + ".\n"
}

// openHandler opens a URL in the host's default browser or in the host app
// registered for an allowlisted scheme. It runs on the host. Only http/https
// and the schemes the user allowlisted are accepted; the URL is passed as a
// single argument to the opener with no shell, so metacharacters are inert.
// When the URL carries a loopback redirect_uri (an OAuth callback), it also
// asks the broker to bridge that callback port from the host into the
// container — best-effort, never blocking the open.
func openHandler(ctx broker.Ctx, w http.ResponseWriter, r *http.Request) {
	raw := r.FormValue("url")
	if !allowedURL(raw, allowedSchemes()) {
		http.Error(w, "only http(s) URLs and schemes allowlisted in the browser-open kit may be opened", http.StatusBadRequest)
		return
	}
	if port, ipv6, ok := detectLoopbackCallback(raw); ok {
		ctx.ForwardLoopback(port, ipv6)
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

// allowedSchemes returns the extra schemes this broker process was started
// with. The value is re-filtered so a hand-started broker cannot be widened by
// a malformed one.
func allowedSchemes() []string {
	raw := os.Getenv(OpenSchemesEnv)
	if raw == "" {
		return nil
	}
	schemes, _ := NormalizeSchemes(strings.Split(raw, ","))
	return schemes
}

// allowedURL reports whether raw may be handed to the host opener: a
// well-formed http(s) URL with a host, or a URL whose scheme the user
// allowlisted. Application schemes commonly use the scheme:///path form, so an
// allowlisted scheme is not required to carry a host. Everything else — no
// scheme, unparseable input, file:// unless allowlisted — is blocked.
func allowedURL(raw string, extra []string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	if u.Scheme == "http" || u.Scheme == "https" {
		return u.Host != ""
	}
	return u.Scheme != "" && slices.Contains(extra, u.Scheme)
}

// detectLoopbackCallback inspects an OAuth authorize URL for a redirect_uri (or
// redirect_url) pointing at a loopback host with an explicit port — the pattern
// of a localhost callback flow. It returns the port and whether it is IPv6.
func detectLoopbackCallback(rawURL string) (port int, ipv6 bool, ok bool) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return 0, false, false
	}
	q := u.Query()
	for _, key := range []string{"redirect_uri", "redirect_url"} {
		v := q.Get(key)
		if v == "" {
			continue
		}
		ru, err := url.Parse(v)
		if err != nil {
			continue
		}
		p, err := strconv.Atoi(ru.Port())
		if err != nil || p <= 0 || p > 65535 {
			continue
		}
		switch ru.Hostname() {
		case "localhost", "127.0.0.1":
			return p, false, true
		case "::1":
			return p, true, true
		}
	}
	return 0, false, false
}
