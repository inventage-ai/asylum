package kit

import (
	"encoding/base64"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/inventage-ai/asylum/internal/broker"
)

// stubCtx satisfies broker.Ctx for handler tests.
type stubCtx struct{}

func (stubCtx) Container() string         { return "test-container" }
func (stubCtx) ForwardLoopback(int, bool) {}

// itermHome prepares a HOME containing a fake cc-status that records how it was
// invoked, and returns the home and the record path.
func itermHome(t *testing.T) (home, record string) {
	t.Helper()
	return itermHomeWithScript(t, "#!/bin/sh\n"+
		"{ echo \"args:$*\"; echo \"session:$TERM_SESSION_ID\"; echo \"stdin:$(cat)\"; } > \"$ASYLUM_TEST_RECORD\"\n")
}

// itermHomeWithScript prepares a HOME whose cc-status is the given script, and
// returns the home and the path the script is expected to record into.
func itermHomeWithScript(t *testing.T, script string) (home, record string) {
	t.Helper()
	home = t.TempDir()
	t.Setenv("HOME", home)
	record = filepath.Join(home, "invocation")
	t.Setenv("ASYLUM_TEST_RECORD", record)

	bin := filepath.Join(home, itermHookRel)
	if err := os.MkdirAll(filepath.Dir(bin), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(bin, []byte(script), 0755); err != nil {
		t.Fatalf("write fake cc-status: %v", err)
	}
	return home, record
}

func sealedFor(t *testing.T, id string) string {
	t.Helper()
	blob, err := broker.Seal(itermSession{TermSessionID: id})
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	return blob
}

func postStatus(t *testing.T, envelope, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, itermRoute, strings.NewReader(body))
	if envelope != "" {
		req.Header.Set(itermSessionHeader, envelope)
	}
	w := httptest.NewRecorder()
	itermStatusHandler(stubCtx{}, w, req)
	return w
}

func TestItermKitRegistered(t *testing.T) {
	k := Get(itermKitName)
	if k == nil {
		t.Fatal("iterm kit is not registered")
	}
	if k.Tier != TierOptIn {
		t.Errorf("tier = %v, want TierOptIn", k.Tier)
	}
	if len(k.Routes) != 1 || k.Routes[0].Path != itermRoute {
		t.Errorf("routes = %+v, want one route at %s", k.Routes, itermRoute)
	}
	if k.MountFunc == nil {
		t.Error("kit contributes no MountFunc, so nothing shadows the hook path")
	}
}

func TestItermMountFuncWithoutHostBinary(t *testing.T) {
	mounts, err := itermMountFunc(CredentialOpts{HomeDir: t.TempDir(), ContainerName: "asylum-test"})
	if err != nil {
		t.Fatalf("itermMountFunc: %v", err)
	}
	if mounts != nil {
		t.Errorf("mounts = %+v, want none when the host integration is absent", mounts)
	}
}

func TestItermMountFuncStagesExecutableShim(t *testing.T) {
	home, _ := itermHome(t)

	mounts, err := itermMountFunc(CredentialOpts{HomeDir: home, ContainerName: "asylum-test"})
	if err != nil {
		t.Fatalf("itermMountFunc: %v", err)
	}
	if len(mounts) != 1 {
		t.Fatalf("mounts = %+v, want exactly one", mounts)
	}
	if want := "~/" + itermHookRel; mounts[0].Destination != want {
		t.Errorf("destination = %q, want %q", mounts[0].Destination, want)
	}

	info, err := os.Stat(mounts[0].HostPath)
	if err != nil {
		t.Fatalf("stat shim: %v", err)
	}
	if mode := info.Mode().Perm(); mode&0111 == 0 {
		t.Errorf("shim mode is %o, want an executable bit", mode)
	}
	content, err := os.ReadFile(mounts[0].HostPath)
	if err != nil {
		t.Fatalf("read shim: %v", err)
	}
	if string(content) != itermShim {
		t.Error("staged shim does not match the kit's shim")
	}
}

func TestItermSessionEnv(t *testing.T) {
	kits := []*Kit{Get(itermKitName)}

	t.Run("kit inactive", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())
		t.Setenv("ITERM_SESSION_ID", "w0t9p0:ABC-123")
		t.Setenv("TERM_SESSION_ID", "w0t9p0:ABC-123")
		if env := ItermSessionEnv([]*Kit{{Name: "browser-open"}}); env != nil {
			t.Errorf("env = %v, want none when the kit is inactive", env)
		}
	})

	t.Run("not an iTerm2 terminal", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())
		t.Setenv("ITERM_SESSION_ID", "")
		t.Setenv("TERM_SESSION_ID", "w0t0p0:SOME-OTHER-TERMINAL")
		if env := ItermSessionEnv(kits); env != nil {
			t.Errorf("env = %v, want none outside iTerm2", env)
		}
	})

	t.Run("no terminal session id", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())
		t.Setenv("ITERM_SESSION_ID", "w0t9p0:ABC-123")
		t.Setenv("TERM_SESSION_ID", "")
		if env := ItermSessionEnv(kits); env != nil {
			t.Errorf("env = %v, want none without the id cc-status reads", env)
		}
	})

	t.Run("seals the id cc-status reads", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())
		t.Setenv("ITERM_SESSION_ID", "w0t9p0:IGNORED-BY-CC-STATUS")
		t.Setenv("TERM_SESSION_ID", "w0t9p0:ABC-123")

		env := ItermSessionEnv(kits)
		blob, ok := env[broker.SessionEnv]
		if !ok {
			t.Fatalf("env = %v, want a %s entry", env, broker.SessionEnv)
		}
		if strings.Contains(blob, "ABC-123") {
			t.Error("envelope exposes the session id in the clear")
		}
		var got itermSession
		if err := broker.Open(blob, &got); err != nil {
			t.Fatalf("Open: %v", err)
		}
		if got.TermSessionID != "w0t9p0:ABC-123" {
			t.Errorf("sealed id = %q, want TERM_SESSION_ID %q", got.TermSessionID, "w0t9p0:ABC-123")
		}
	})
}

func TestItermStatusHandlerRejects(t *testing.T) {
	home, _ := itermHome(t)
	valid := sealedFor(t, "w0t9p0:ABC-123")

	raw, err := base64.RawURLEncoding.DecodeString(valid)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	tampered := append([]byte(nil), raw...)
	tampered[len(tampered)-1] ^= 0xff

	empty, err := broker.Seal(itermSession{})
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	tests := []struct {
		name     string
		envelope string
		body     string
		want     int
	}{
		{"no envelope", "", "{}", http.StatusBadRequest},
		{"unopenable envelope", "not-an-envelope", "{}", http.StatusBadRequest},
		{"tampered envelope", base64.RawURLEncoding.EncodeToString(tampered), "{}", http.StatusBadRequest},
		{"envelope names no session", empty, "{}", http.StatusBadRequest},
		{"oversized body", valid, strings.Repeat("x", itermMaxBody+1), http.StatusRequestEntityTooLarge},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Remove(filepath.Join(home, "invocation"))
			if got := postStatus(t, tt.envelope, tt.body).Code; got != tt.want {
				t.Errorf("status = %d, want %d", got, tt.want)
			}
			if _, err := os.Stat(filepath.Join(home, "invocation")); err == nil {
				t.Error("the host binary ran despite a rejected request")
			}
		})
	}
}

func TestItermStatusHandlerHostBinaryMissing(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	envelope := sealedFor(t, "w0t9p0:ABC-123")
	if got := postStatus(t, envelope, "{}").Code; got != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", got, http.StatusServiceUnavailable)
	}
}

func TestItermStatusHandlerRunsHostBinary(t *testing.T) {
	_, record := itermHome(t)
	// The broker inherits the environment of whichever session spawned it, so
	// it carries that tab's id for its whole life. The envelope's value must
	// win, or every session would paint the tab that started the container.
	t.Setenv("TERM_SESSION_ID", "w0t0p0:THE-BROKERS-OWN-STALE-TAB")
	envelope := sealedFor(t, "w0t9p0:ABC-123")

	// A payload shaped like command-line flags: none of it may reach argv.
	body := `{"hook_event_name":"PreToolUse","tool_input":{"command":"--session evil --dot-color #000"}}`
	if got := postStatus(t, envelope, body).Code; got != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", got, http.StatusNoContent)
	}

	out, err := os.ReadFile(record)
	if err != nil {
		t.Fatalf("read invocation record: %v", err)
	}
	got := string(out)
	if !strings.Contains(got, "args:\n") {
		t.Errorf("host binary received arguments; record was:\n%s", got)
	}
	if !strings.Contains(got, "session:w0t9p0:ABC-123\n") {
		t.Errorf("ITERM_SESSION_ID not set from the envelope; record was:\n%s", got)
	}
	if strings.Contains(got, "THE-BROKERS-OWN-STALE-TAB") {
		t.Errorf("the broker's inherited session id won over the envelope's; record was:\n%s", got)
	}
	if !strings.Contains(got, body) {
		t.Errorf("payload did not reach stdin; record was:\n%s", got)
	}
}

func TestItermRulesSnippetAssembled(t *testing.T) {
	kits := []*Kit{Get(itermKitName)}
	got := AssembleRulesSnippets(kits, nil)
	if !strings.Contains(got, "iTerm2 status bar") {
		t.Fatalf("assembled rules omit the iterm snippet; got %q", got)
	}
}

// Hook events describe a sequence of state transitions. Two updates for one
// terminal running concurrently could land out of order and leave the status
// bar showing idle while the session works, so the handler serializes them.
func TestItermStatusHandlerSerializesPerSession(t *testing.T) {
	_, record := itermHomeWithScript(t, "#!/bin/sh\n"+
		"marker=$(cat)\n"+
		"echo \"start:$marker\" >> \"$ASYLUM_TEST_RECORD\"\n"+
		"sleep 0.05\n"+
		"echo \"end:$marker\" >> \"$ASYLUM_TEST_RECORD\"\n")
	envelope := sealedFor(t, "w0t9p0:ABC-123")

	const n = 4
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			postStatus(t, envelope, strconv.Itoa(i))
		}()
	}
	wg.Wait()

	data, err := os.ReadFile(record)
	if err != nil {
		t.Fatalf("read record: %v", err)
	}
	lines := strings.Fields(strings.TrimSpace(string(data)))
	if len(lines) != 2*n {
		t.Fatalf("got %d record lines, want %d:\n%s", len(lines), 2*n, data)
	}
	// Serialized execution means every start is immediately followed by the
	// matching end. Any interleaving breaks the pairing.
	for i := 0; i < len(lines); i += 2 {
		start, end := lines[i], lines[i+1]
		if !strings.HasPrefix(start, "start:") || !strings.HasPrefix(end, "end:") {
			t.Fatalf("invocations interleaved at line %d:\n%s", i, data)
		}
		if strings.TrimPrefix(start, "start:") != strings.TrimPrefix(end, "end:") {
			t.Fatalf("mismatched start/end pair at line %d:\n%s", i, data)
		}
	}
}

// The shim is a shell script held as a Go string, so nothing else would catch a
// quoting error in it.
func TestItermShimForwardsAndAlwaysSucceeds(t *testing.T) {
	if _, err := exec.LookPath("curl"); err != nil {
		t.Skip("curl not available")
	}

	type received struct {
		session string
		body    string
	}
	got := make(chan received, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		got <- received{session: r.Header.Get(itermSessionHeader), body: string(b)}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	host, port, err := net.SplitHostPort(strings.TrimPrefix(srv.URL, "http://"))
	if err != nil {
		t.Fatalf("split server address: %v", err)
	}

	shim := filepath.Join(t.TempDir(), "cc-status")
	if err := os.WriteFile(shim, []byte(itermShim), 0755); err != nil {
		t.Fatalf("write shim: %v", err)
	}

	baseEnv := []string{
		"PATH=" + os.Getenv("PATH"),
		"ASYLUM_BROKER_HOST=" + host,
		"ASYLUM_BROKER_PORT=" + port,
		"ASYLUM_BROKER_TOKEN=test-token",
	}
	payload := `{"hook_event_name":"PreToolUse"}`

	exec1 := exec.Command(shim)
	exec1.Stdin = strings.NewReader(payload)
	exec1.Env = append(baseEnv, "ASYLUM_SESSION=sealed-blob")
	if err := exec1.Run(); err != nil {
		t.Fatalf("shim exited non-zero with a healthy broker: %v", err)
	}
	select {
	case r := <-got:
		if r.session != "sealed-blob" {
			t.Errorf("envelope header = %q, want %q", r.session, "sealed-blob")
		}
		if r.body != payload {
			t.Errorf("forwarded body = %q, want %q", r.body, payload)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("broker received no request")
	}

	// Without an envelope the shim must return immediately and send nothing.
	exec2 := exec.Command(shim)
	exec2.Stdin = strings.NewReader(payload)
	exec2.Env = baseEnv
	if err := exec2.Run(); err != nil {
		t.Fatalf("shim exited non-zero without an envelope: %v", err)
	}
	select {
	case r := <-got:
		t.Errorf("shim sent a request without an envelope: %+v", r)
	case <-time.After(200 * time.Millisecond):
	}

	// A dead broker must not fail the hook.
	exec3 := exec.Command(shim)
	exec3.Stdin = strings.NewReader(payload)
	exec3.Env = append([]string{
		"PATH=" + os.Getenv("PATH"),
		"ASYLUM_BROKER_HOST=127.0.0.1",
		"ASYLUM_BROKER_PORT=1",
		"ASYLUM_BROKER_TOKEN=test-token",
	}, "ASYLUM_SESSION=sealed-blob")
	if err := exec3.Run(); err != nil {
		t.Fatalf("shim exited non-zero with an unreachable broker: %v", err)
	}
}
