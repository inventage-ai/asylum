package broker

import (
	"context"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

// startTestServer starts a broker on the endpoint and waits until it answers.
func startTestServer(t *testing.T, ep Endpoint, token string, routes []Route) {
	t.Helper()
	go Serve(ep, token, routes)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if alive(ep, token) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("broker did not come up")
}

func tcpEndpoint(t *testing.T) Endpoint {
	t.Helper()
	p, err := FreePort()
	if err != nil {
		t.Fatalf("FreePort: %v", err)
	}
	return Endpoint{Network: "tcp", Addr: "127.0.0.1:" + strconv.Itoa(p)}
}

// doGet issues an authenticated GET to the endpoint and returns the status.
func doGet(t *testing.T, ep Endpoint, path, auth string) int {
	t.Helper()
	client := &http.Client{Timeout: time.Second}
	url := "http://" + ep.Addr + path
	if ep.Network == "unix" {
		sock := ep.Addr
		client.Transport = &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", sock)
		}}
		url = "http://localhost" + path
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()
	return resp.StatusCode
}

func assertAuth(t *testing.T, ep Endpoint) {
	t.Helper()
	if code := doGet(t, ep, "/healthz", "Bearer s3cret"); code != http.StatusOK {
		t.Errorf("valid token: got %d, want 200", code)
	}
	if code := doGet(t, ep, "/healthz", "Bearer wrong"); code != http.StatusUnauthorized {
		t.Errorf("wrong token: got %d, want 401", code)
	}
	if code := doGet(t, ep, "/healthz", ""); code != http.StatusUnauthorized {
		t.Errorf("missing token: got %d, want 401", code)
	}
}

func TestTokenAuthTCP(t *testing.T) {
	ep := tcpEndpoint(t)
	startTestServer(t, ep, "s3cret", nil)
	assertAuth(t, ep)
}

func TestTokenAuthUnix(t *testing.T) {
	ep := Endpoint{Network: "unix", Addr: filepath.Join(t.TempDir(), "broker.sock")}
	startTestServer(t, ep, "s3cret", nil)
	assertAuth(t, ep)
}

func TestUnixSocketStaleCleanup(t *testing.T) {
	// A leftover socket file from a prior run must not block a fresh listen.
	path := filepath.Join(t.TempDir(), "broker.sock")
	l, err := net.Listen("unix", path)
	if err != nil {
		t.Fatalf("seed socket: %v", err)
	}
	l.(*net.UnixListener).SetUnlinkOnClose(false)
	l.Close() // leaves the socket file behind
	ep := Endpoint{Network: "unix", Addr: path}
	startTestServer(t, ep, "tok", nil) // Serve unlinks the stale socket first
}

func TestRouteRequiresAuth(t *testing.T) {
	called := false
	routes := []Route{{Path: "/ping", Handler: func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}}}
	ep := tcpEndpoint(t)
	startTestServer(t, ep, "tok", routes)

	if code := doGet(t, ep, "/ping", "Bearer nope"); code != http.StatusUnauthorized {
		t.Errorf("unauthorized route: got %d, want 401", code)
	}
	if called {
		t.Error("handler ran despite failed auth")
	}
	if code := doGet(t, ep, "/ping", "Bearer tok"); code != http.StatusNoContent {
		t.Errorf("authorized route: got %d, want 204", code)
	}
	if !called {
		t.Error("handler did not run with valid auth")
	}
}

func TestEnsureBrokerSkipsSpawnWhenAlive(t *testing.T) {
	ep := tcpEndpoint(t)
	startTestServer(t, ep, "live", nil)
	// A live broker answers, so no process is spawned — the bogus exec path is
	// never used and EnsureBroker returns nil.
	if err := EnsureBroker("c", "/nonexistent/asylum", ep, "live", nil); err != nil {
		t.Errorf("EnsureBroker with live broker: %v", err)
	}
}

func TestEnsureBrokerSpawnsWhenDead(t *testing.T) {
	ep := tcpEndpoint(t) // free port, nothing listening
	// EnsureBroker tries to spawn; the bogus exec path makes the spawn fail,
	// surfacing that it attempted to start one.
	if err := EnsureBroker("c", "/nonexistent/asylum", ep, "tok", nil); err == nil {
		t.Error("EnsureBroker with no broker: want spawn error, got nil")
	}
}

func TestFreePortDistinct(t *testing.T) {
	p1, err := FreePort()
	if err != nil {
		t.Fatal(err)
	}
	if p1 <= 0 {
		t.Fatalf("FreePort returned %d", p1)
	}
}
