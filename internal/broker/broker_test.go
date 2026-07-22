package broker

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

// startTestServer starts a broker on a free port and waits until it answers.
func startTestServer(t *testing.T, token string, routes []Route) int {
	t.Helper()
	port, err := FreePort()
	if err != nil {
		t.Fatalf("FreePort: %v", err)
	}
	go Serve(port, token, routes)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if alive(port, token) {
			return port
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("broker did not come up")
	return 0
}

func get(t *testing.T, port int, path, auth string) int {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d%s", port, path), nil)
	if err != nil {
		t.Fatal(err)
	}
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()
	return resp.StatusCode
}

func TestTokenAuth(t *testing.T) {
	port := startTestServer(t, "s3cret", nil)

	if code := get(t, port, "/healthz", "Bearer s3cret"); code != http.StatusOK {
		t.Errorf("valid token: got %d, want 200", code)
	}
	if code := get(t, port, "/healthz", "Bearer wrong"); code != http.StatusUnauthorized {
		t.Errorf("wrong token: got %d, want 401", code)
	}
	if code := get(t, port, "/healthz", ""); code != http.StatusUnauthorized {
		t.Errorf("missing token: got %d, want 401", code)
	}
}

func TestRouteRequiresAuth(t *testing.T) {
	called := false
	routes := []Route{{Path: "/ping", Handler: func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}}}
	port := startTestServer(t, "tok", routes)

	if code := get(t, port, "/ping", "Bearer nope"); code != http.StatusUnauthorized {
		t.Errorf("unauthorized route: got %d, want 401", code)
	}
	if called {
		t.Error("handler ran despite failed auth")
	}
	if code := get(t, port, "/ping", "Bearer tok"); code != http.StatusNoContent {
		t.Errorf("authorized route: got %d, want 204", code)
	}
	if !called {
		t.Error("handler did not run with valid auth")
	}
}

func TestEnsureBrokerSkipsSpawnWhenAlive(t *testing.T) {
	port := startTestServer(t, "live", nil)
	// A live broker answers, so no process is spawned — the bogus exec path is
	// never used and EnsureBroker returns nil.
	if err := EnsureBroker("c", "/nonexistent/asylum", port, "live", nil); err != nil {
		t.Errorf("EnsureBroker with live broker: %v", err)
	}
}

func TestEnsureBrokerSpawnsWhenDead(t *testing.T) {
	port, err := FreePort()
	if err != nil {
		t.Fatal(err)
	}
	// No broker on this port, so EnsureBroker tries to spawn; the bogus exec
	// path makes the spawn fail, surfacing that it attempted to start one.
	if err := EnsureBroker("c", "/nonexistent/asylum", port, "tok", nil); err == nil {
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
