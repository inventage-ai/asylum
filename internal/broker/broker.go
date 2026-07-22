// Package broker implements a host-side HTTP server, scoped to a container's
// lifetime, that serves token-authenticated routes contributed by kits. It lets
// a sandboxed container ask the host to perform actions it cannot do itself
// (e.g. open a URL in the real browser).
//
// The broker never binds a publicly reachable address. On a native Linux engine
// it listens on a Unix domain socket bind-mounted into the one container; on a
// VM-backed engine (Docker Desktop, macOS) it listens on 127.0.0.1, reached from
// the container via host.docker.internal. Both are captured by Endpoint.
package broker

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// SockName is the broker's Unix socket filename, within the per-container
// directory on the host and the bind-mounted directory in the container.
const SockName = "broker.sock"

// Endpoint describes how to listen for and dial the broker.
type Endpoint struct {
	Network string // "unix" or "tcp"
	Addr    string // socket path, or "127.0.0.1:<port>"
}

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

// FreePort asks the OS for an unused TCP port (used only by the TCP transport).
func FreePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// Serve mounts the routes under token authentication and serves on the endpoint
// until the process exits or the bind fails. For a Unix socket it unlinks a
// stale socket before listening. A bind failure (another broker already owns the
// endpoint) is returned so the caller can treat it as a clean "already running"
// exit.
func Serve(ep Endpoint, token string, routes []Route) error {
	if ep.Network == "unix" {
		os.Remove(ep.Addr) // clear a stale socket left by a previous run
	}
	l, err := net.Listen(ep.Network, ep.Addr)
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", authWrap(token, func(w http.ResponseWriter, r *http.Request) {}))
	for _, rt := range routes {
		mux.HandleFunc(rt.Path, authWrap(token, rt.Handler))
	}
	return http.Serve(l, mux)
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
// endpoint; if no live broker answers it spawns a detached `asylum __broker`
// process. Spawning when one already runs is harmless — the new process's bind
// fails and it exits.
//
// The token is passed to the child via its environment, not argv: process
// command lines are world-readable (/proc/<pid>/cmdline), while environ is
// readable only by the owning user, so this keeps the token off multi-user
// hosts' ps output.
func EnsureBroker(cname, execPath string, ep Endpoint, token string, kitNames []string) error {
	if alive(ep, token) {
		return nil
	}
	args := []string{"__broker", "--container", cname, "--net", ep.Network, "--addr", ep.Addr}
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
// endpoint.
func alive(ep Endpoint, token string) bool {
	client := &http.Client{Timeout: 500 * time.Millisecond}
	url := "http://" + ep.Addr + "/healthz"
	if ep.Network == "unix" {
		sock := ep.Addr
		client.Transport = &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, "unix", sock)
			},
		}
		url = "http://localhost/healthz"
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
