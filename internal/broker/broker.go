// Package broker implements a host-side HTTP server, scoped to a container's
// lifetime, that serves token-authenticated routes contributed by kits. It lets
// a sandboxed container ask the host to perform actions it cannot do itself
// (e.g. open a URL in the real browser).
package broker

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Route is a broker endpoint contributed by a kit. Handlers run on the host.
type Route struct {
	Path    string
	Handler http.HandlerFunc
}

// Token returns a random hex token used to authenticate broker requests.
func Token() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// FreePort asks the OS for an unused TCP port.
func FreePort() (int, error) {
	l, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// Serve mounts the routes under token authentication and serves on
// 0.0.0.0:port until the process exits or the bind fails. A bind failure
// (another broker already owns the port) is returned so the caller can treat
// it as a clean "already running" exit.
func Serve(port int, token string, routes []Route) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", authWrap(token, func(w http.ResponseWriter, r *http.Request) {}))
	for _, rt := range routes {
		mux.HandleFunc(rt.Path, authWrap(token, rt.Handler))
	}
	srv := &http.Server{Addr: fmt.Sprintf("0.0.0.0:%d", port), Handler: mux}
	return srv.ListenAndServe()
}

// authWrap rejects any request whose bearer token does not match, in constant
// time, before the wrapped handler runs.
func authWrap(token string, h http.HandlerFunc) http.HandlerFunc {
	want := []byte("Bearer " + token)
	return func(w http.ResponseWriter, r *http.Request) {
		got := []byte(r.Header.Get("Authorization"))
		if subtle.ConstantTimeCompare(got, want) != 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		h(w, r)
	}
}

// EnsureBroker guarantees a broker for the container is running. It probes the
// port; if no live broker answers it spawns a detached `asylum __broker`
// process. Spawning when one already runs is harmless — the new process's bind
// fails and it exits.
//
// The token is passed to the child via its environment, not argv: process
// command lines are world-readable (/proc/<pid>/cmdline), while environ is
// readable only by the owning user, so this keeps the token off multi-user
// hosts' ps output.
func EnsureBroker(cname, execPath string, port int, token string, kitNames []string) error {
	if alive(port, token) {
		return nil
	}
	args := []string{"__broker", "--container", cname, "--port", strconv.Itoa(port)}
	if len(kitNames) > 0 {
		args = append(args, "--kits", strings.Join(kitNames, ","))
	}
	cmd := exec.Command(execPath, args...)
	cmd.Env = append(os.Environ(), "ASYLUM_BROKER_TOKEN="+token)
	// Detach from the session so the broker outlives the current asylum process.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	return cmd.Start()
}

// alive reports whether a broker answers an authenticated health check on the
// host loopback. The port is bound on 0.0.0.0, so 127.0.0.1 reaches it here.
func alive(port int, token string) bool {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/healthz", port), nil)
	if err != nil {
		return false
	}
	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
