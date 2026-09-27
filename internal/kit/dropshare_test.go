package kit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// dropshareEcho builds a callback the way Dropshare does: the file value
// path-encoded, with & + = left raw.
func dropshareEcho(id, file, shareURL string) string {
	return dropshareScheme + ":///done/" + id + "?file=" + (&url.URL{Path: file}).EscapedPath() + "&url=" + shareURL
}

func TestParseDropshareCallback(t *testing.T) {
	tests := []struct {
		name, raw, file, url string
		wantErr              bool
	}{
		{
			name: "plain path",
			raw:  "asylum-dropshare-poc:///done?file=/Users/simon/Tools/asylum/build/dropshare-poc/poc-upload.txt&url=https://dl.dplx.ch/u9XunWboZ6.txt",
			file: "/Users/simon/Tools/asylum/build/dropshare-poc/poc-upload.txt",
			url:  "https://dl.dplx.ch/u9XunWboZ6.txt",
		},
		{
			name: "reserved characters as observed",
			raw:  "asylum-dropshare-poc:///done?file=/Users/simon/Tools/asylum/build/dropshare-poc/a%20&%20b%231+c=%C3%BC%2520.txt&url=https://dl.dplx.ch/o5RHVultSh.txt",
			file: "/Users/simon/Tools/asylum/build/dropshare-poc/a & b#1+c=ü%20.txt",
			url:  "https://dl.dplx.ch/o5RHVultSh.txt",
		},
		{
			name: "path containing &url=",
			raw:  "asylum-dropshare:///done/x?file=/p/a&url=b.txt&url=https://dl.example.com/k.txt",
			file: "/p/a&url=b.txt",
			url:  "https://dl.example.com/k.txt",
		},
		{name: "missing file", raw: "asylum-dropshare:///done/x?url=https://dl.example.com/k.txt", wantErr: true},
		{name: "missing url", raw: "asylum-dropshare:///done/x?file=/p/a.txt", wantErr: true},
		{name: "http url", raw: "asylum-dropshare:///done/x?file=/p/a.txt&url=http://dl.example.com/k.txt", wantErr: true},
		{name: "relative url", raw: "asylum-dropshare:///done/x?file=/p/a.txt&url=k.txt", wantErr: true},
		{name: "bad escape", raw: "asylum-dropshare:///done/x?file=/p/%zz&url=https://dl.example.com/k.txt", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, u, err := parseDropshareCallback(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Errorf("want error, got file=%q url=%q", file, u)
				}
				return
			}
			if err != nil || file != tt.file || u != tt.url {
				t.Errorf("got (%q, %q, %v), want (%q, %q)", file, u, err, tt.file, tt.url)
			}
		})
	}
}

func TestDropshareUploadURLRoundTrip(t *testing.T) {
	id := strings.Repeat("ab", 16)
	for _, file := range []string{
		"/Users/simon/Tools/asylum/build/dropshare-poc/poc-upload.txt",
		"/Users/simon/Tools/asylum/build/dropshare-poc/a & b#1+c=ü%20.txt",
		"/p/?x=1;y'z\".txt",
	} {
		u, err := url.Parse(dropshareUploadURL(file, id))
		if err != nil {
			t.Fatalf("%q: %v", file, err)
		}
		q := u.Query()
		if q.Get("file") != file {
			t.Errorf("file param decodes to %q, want %q", q.Get("file"), file)
		}
		if want := dropshareScheme + ":///done/" + id; q.Get("callback") != want {
			t.Errorf("callback param = %q, want %q", q.Get("callback"), want)
		}
		got, _, err := parseDropshareCallback(dropshareEcho(id, q.Get("file"), "https://dl.example.com/k.txt"))
		if err != nil || got != file {
			t.Errorf("echoed callback parses to (%q, %v), want %q", got, err, file)
		}
	}
}

func TestDropshareAppletHash(t *testing.T) {
	a := dropshareAppletHash("/Users/a/.asylum/dropshare")
	if a != dropshareAppletHash("/Users/a/.asylum/dropshare") {
		t.Error("hash is not stable")
	}
	if a == dropshareAppletHash("/Users/b/.asylum/dropshare") {
		t.Error("hash ignores the install directory")
	}
}

// A current applet must be left alone. On Linux a rebuild attempt fails
// because osacompile is missing, so any error here means it tried to rebuild.
func TestEnsureDropshareAppletSkipsCurrent(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, dropshareAppName, "Contents"), 0755)
	os.WriteFile(filepath.Join(dir, ".applet-sha256"), []byte(dropshareAppletHash(dir)), 0600)
	if err := ensureDropshareApplet(dir); err != nil {
		t.Errorf("current applet was rebuilt: %v", err)
	}
}

func TestWaitDropshareCallback(t *testing.T) {
	defer swap(&dropsharePoll, time.Millisecond)()
	inbox := t.TempDir()
	os.WriteFile(filepath.Join(inbox, "other"), []byte("not mine\n"), 0600)
	go func() {
		time.Sleep(20 * time.Millisecond)
		writeRecord(inbox, "mine", "the callback")
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	got, err := waitDropshareCallback(ctx, inbox, "mine")
	if err != nil || got != "the callback" {
		t.Fatalf("got (%q, %v), want the callback", got, err)
	}
	if fileExists(filepath.Join(inbox, "mine")) {
		t.Error("consumed record was not removed")
	}
	if !fileExists(filepath.Join(inbox, "other")) {
		t.Error("another request's record was removed")
	}
}

func TestPruneOlderThan(t *testing.T) {
	dir := t.TempDir()
	old, fresh, oldStage := filepath.Join(dir, "old"), filepath.Join(dir, "fresh"), filepath.Join(dir, "old-stage")
	os.WriteFile(old, nil, 0600)
	os.WriteFile(fresh, nil, 0600)
	os.MkdirAll(filepath.Join(oldStage, "sub"), 0700)
	for _, p := range []string{old, oldStage} {
		os.Chtimes(p, time.Now().Add(-2*time.Hour), time.Now().Add(-2*time.Hour))
	}

	pruneOlderThan(dir, time.Hour)
	if fileExists(old) || dirExists(oldStage) || !fileExists(fresh) {
		t.Errorf("after prune: old=%v old-stage=%v fresh=%v", fileExists(old), dirExists(oldStage), fileExists(fresh))
	}
}

func TestDropshareCallbackScriptRejectsBadIDs(t *testing.T) {
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "inbox"), 0700)
	script := filepath.Join(dir, "callback.sh")
	os.WriteFile(script, []byte(dropshareCallbackScript), 0644)
	good := strings.Repeat("0f", 16)

	for _, u := range []string{
		dropshareScheme + ":///done/" + good + "?file=/p/a.txt&url=https://x/y",
		dropshareScheme + ":///done/../../evil?file=/p&url=https://x/y",
		dropshareScheme + ":///done/" + good[:30] + "?file=/p&url=https://x/y",
		dropshareScheme + ":///done/" + strings.ToUpper(good) + "?file=/p&url=https://x/y",
		"other:///done/" + good,
	} {
		if out, err := exec.Command("/bin/sh", script, u).CombinedOutput(); err != nil {
			t.Fatalf("%s: %v: %s", u, err, out)
		}
	}
	entries, _ := os.ReadDir(filepath.Join(dir, "inbox"))
	if len(entries) != 1 || entries[0].Name() != good {
		t.Errorf("inbox holds %v, want only %s", entries, good)
	}
	if fileExists(filepath.Join(dir, "evil")) {
		t.Error("crafted ID wrote outside the inbox")
	}
}

func TestDropshareDockerSnippetStagesCommandAndSkill(t *testing.T) {
	k := Get(dropshareKitName)
	if k == nil || !k.ProvidesSkills || k.Tier != TierOptIn {
		t.Fatalf("dropshare kit = %+v, want opt-in with skills", k)
	}
	for _, want := range []string{"/usr/local/bin/asylum-dropshare", "/opt/asylum-skills/.claude/skills/dropshare/SKILL.md"} {
		if !strings.Contains(k.DockerSnippet, want) {
			t.Errorf("Docker snippet does not stage %s", want)
		}
	}
}

func TestHoldForSize(t *testing.T) {
	tests := []struct {
		size int64
		want time.Duration
	}{
		{0, 10 * time.Second},
		{300 << 10, 10 * time.Second},
		{100 << 20, 110 * time.Second},
		{10 << 30, 10250 * time.Second},
	}
	for _, tt := range tests {
		if got := holdForSize(tt.size); got != tt.want {
			t.Errorf("holdForSize(%d) = %v, want %v", tt.size, got, tt.want)
		}
	}
}

// fakeHost stands in for the Mac. A temp directory plays the container's
// filesystem, tar -h plays `docker cp -L`, and the opener hands every upload
// to answer, whose return value, if not empty, becomes the callback record.
type fakeHost struct {
	container string // the simulated container filesystem
	dir       string // ~/.asylum/dropshare
}

func newFakeHost(t *testing.T, answer func(id, file string) string) *fakeHost {
	t.Helper()
	h := &fakeHost{container: t.TempDir(), dir: t.TempDir()}
	for _, sub := range []string{"inbox", "staging"} {
		os.Mkdir(filepath.Join(h.dir, sub), 0700)
	}
	restore := []func(){
		swap(&dropshareSupported, true),
		swap(&dropsharePoll, time.Millisecond),
		swap(&dropsharePrepare, func() (string, error) { return h.dir, nil }),
		swap(&dropshareCopyOut, func(_, path string) *exec.Cmd {
			return exec.Command("tar", "-chf", "-", "-C", filepath.Dir(path), filepath.Base(path))
		}),
		swap(&dropshareOpen, func(raw string) (func(), error) {
			id, file := uploadParams(raw)
			if cb := answer(id, file); cb != "" {
				go writeRecord(filepath.Join(h.dir, "inbox"), id, cb)
			}
			return func() {}, nil
		}),
	}
	t.Cleanup(func() {
		for _, r := range restore {
			r()
		}
	})
	return h
}

func (h *fakeHost) file(t *testing.T, name, content string) string {
	t.Helper()
	p := filepath.Join(h.container, name)
	os.MkdirAll(filepath.Dir(p), 0755)
	os.WriteFile(p, []byte(content), 0644)
	return p
}

func uploadParams(raw string) (id, file string) {
	u, _ := url.Parse(raw)
	return strings.TrimPrefix(u.Query().Get("callback"), dropshareScheme+":///done/"), u.Query().Get("file")
}

func echoBase(id, file string) string {
	return dropshareEcho(id, file, "https://dl.example.com/"+filepath.Base(file))
}

func postUpload(path string) *httptest.ResponseRecorder {
	return postUploadCtx(context.Background(), path)
}

func postUploadCtx(ctx context.Context, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, dropshareRoute, strings.NewReader(url.Values{"path": {path}}.Encode())).WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	dropshareHandler(stubCtx{}, w, req)
	return w
}

func TestDropshareHandler(t *testing.T) {
	h := newFakeHost(t, func(id, file string) string {
		switch filepath.Base(file) {
		case "slow.txt":
			return ""
		case "wrong.txt":
			return dropshareEcho(id, "/elsewhere.txt", "https://dl.example.com/w.txt")
		}
		return echoBase(id, file)
	})
	defer swap(&dropshareTimeout, 50*time.Millisecond)()
	weird := h.file(t, "a & b#1+c=ü%20.txt", "x")
	slow, wrong := h.file(t, "slow.txt", "x"), h.file(t, "wrong.txt", "x")
	h.file(t, "sub/inner.txt", "x")

	tests := []struct {
		name, path string
		code       int
		body       string
	}{
		{"uploads", weird, http.StatusOK, "https://dl.example.com/a & b#1+c=ü%20.txt\n"},
		{"relative path", "a.txt", http.StatusBadRequest, "not absolute"},
		{"missing file", filepath.Join(h.container, "nope.txt"), http.StatusBadRequest, "nope.txt"},
		{"directory", filepath.Join(h.container, "sub"), http.StatusBadRequest, "not a regular file"},
		{"no callback", slow, http.StatusGatewayTimeout, "host clipboard"},
		{"callback for another file", wrong, http.StatusBadGateway, "callback names"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := postUpload(tt.path)
			if w.Code != tt.code || !strings.Contains(w.Body.String(), tt.body) {
				t.Errorf("got %d %q, want %d containing %q", w.Code, w.Body.String(), tt.code, tt.body)
			}
		})
	}
}

func TestDropshareHandlerRequiresMac(t *testing.T) {
	defer swap(&dropshareSupported, false)()
	if w := postUpload("/p/a.txt"); w.Code != http.StatusNotImplemented {
		t.Errorf("non-macOS host: got %d, want 501", w.Code)
	}
}

// Dropshare must read a copy the container cannot touch, so swapping the
// source for a symlink to a host file after the request changes nothing.
func TestDropshareUploadsStagedCopy(t *testing.T) {
	var h *fakeHost
	var uploaded, staged string
	h = newFakeHost(t, func(id, file string) string {
		staged = file
		os.Remove(filepath.Join(h.container, "shot.png"))
		os.Symlink(filepath.Join(h.container, "secret"), filepath.Join(h.container, "shot.png"))
		b, _ := os.ReadFile(file)
		uploaded = string(b)
		return echoBase(id, file)
	})
	h.file(t, "secret", "host-only key")
	shot := h.file(t, "shot.png", "screenshot")

	if w := postUpload(shot); w.Code != http.StatusOK {
		t.Fatalf("got %d %q, want 200", w.Code, w.Body.String())
	}
	if uploaded != "screenshot" {
		t.Errorf("Dropshare read %q, want the staged screenshot", uploaded)
	}
	if !strings.HasPrefix(staged, filepath.Join(h.dir, "staging")+"/") || filepath.Base(staged) != "shot.png" {
		t.Errorf("Dropshare was handed %q, want a staged copy named shot.png", staged)
	}
	if fileExists(staged) {
		t.Error("staged copy was not removed after the callback")
	}
}

func TestDropshareRejectedUploadLeavesNoStaging(t *testing.T) {
	h := newFakeHost(t, echoBase)
	h.file(t, "tree/a.txt", "x")
	postUpload(filepath.Join(h.container, "tree"))
	postUpload(filepath.Join(h.container, "missing.txt"))
	if entries, _ := os.ReadDir(filepath.Join(h.dir, "staging")); len(entries) != 0 {
		t.Errorf("staging holds %d entries after rejected uploads, want none", len(entries))
	}
}

func TestDropshareUploadsTakeTurns(t *testing.T) {
	var inFlight atomic.Int32
	var overlapped atomic.Bool
	var h *fakeHost
	h = newFakeHost(t, func(id, file string) string {
		if inFlight.Add(1) > 1 {
			overlapped.Store(true)
		}
		go func() {
			time.Sleep(30 * time.Millisecond)
			inFlight.Add(-1)
			writeRecord(filepath.Join(h.dir, "inbox"), id, echoBase(id, file))
		}()
		return ""
	})
	names := []string{"one.txt", "two.txt", "three.txt"}
	results := make(chan [2]string, len(names))
	for _, n := range names {
		p := h.file(t, n, n)
		go func() { results <- [2]string{n, postUpload(p).Body.String()} }()
	}
	for range names {
		r := <-results
		if want := "https://dl.example.com/" + r[0] + "\n"; r[1] != want {
			t.Errorf("%s got %q, want %q", r[0], r[1], want)
		}
	}
	if overlapped.Load() {
		t.Error("an upload started while another was still in flight")
	}
}

func TestDropshareHoldReleasesStuckUpload(t *testing.T) {
	handedOff := make(chan struct{})
	h := newFakeHost(t, func(id, file string) string {
		if filepath.Base(file) == "stuck.txt" {
			close(handedOff)
			return ""
		}
		return echoBase(id, file)
	})
	defer swap(&dropshareHold, func(int64) time.Duration { return 20 * time.Millisecond })()
	defer swap(&dropshareTimeout, time.Second)()
	stuckPath, next := h.file(t, "stuck.txt", "x"), h.file(t, "next.txt", "x")

	stuck := make(chan int)
	go func() { stuck <- postUpload(stuckPath).Code }()
	<-handedOff
	start := time.Now()
	if w := postUpload(next); w.Code != http.StatusOK {
		t.Errorf("upload behind a stuck one got %d %q, want 200", w.Code, w.Body.String())
	}
	if d := time.Since(start); d < 15*time.Millisecond || d > 500*time.Millisecond {
		t.Errorf("upload behind a stuck one waited %v, want about the 20ms hold", d)
	}
	if c := <-stuck; c != http.StatusGatewayTimeout {
		t.Errorf("stuck upload got %d, want 504", c)
	}
}

func TestDropshareHoldCountsFromHandoff(t *testing.T) {
	h := newFakeHost(t, func(string, string) string { return "" })
	defer swap(&dropshareHold, func(int64) time.Duration { return 30 * time.Millisecond })()
	defer swap(&dropshareTimeout, 300*time.Millisecond)()
	var mu sync.Mutex
	var handedOff, secondStart time.Time
	opening := make(chan struct{})
	defer swap(&dropshareOpen, func(raw string) (func(), error) {
		_, file := uploadParams(raw)
		if filepath.Base(file) == "second.txt" {
			mu.Lock()
			secondStart = time.Now()
			mu.Unlock()
			return func() {}, nil
		}
		close(opening)
		time.Sleep(50 * time.Millisecond) // a slow handoff
		mu.Lock()
		handedOff = time.Now()
		mu.Unlock()
		return func() {}, nil
	})()
	first, second := h.file(t, "first.txt", "x"), h.file(t, "second.txt", "x")

	done := make(chan struct{})
	go func() { postUpload(first); close(done) }()
	<-opening
	postUpload(second)
	<-done
	mu.Lock()
	defer mu.Unlock()
	if gap := secondStart.Sub(handedOff); gap < 30*time.Millisecond {
		t.Errorf("second upload started %v after the first handoff, want at least the 30ms hold", gap)
	}
}

func TestDropshareQueueTimeout(t *testing.T) {
	h := newFakeHost(t, echoBase)
	defer swap(&dropshareTimeout, 50*time.Millisecond)()
	unlock, err := lockDropshareUploads(context.Background(), filepath.Join(h.dir, ".upload.lock"))
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()

	w := postUpload(h.file(t, "a.txt", "a"))
	if w.Code != http.StatusServiceUnavailable || !strings.Contains(w.Body.String(), "nothing was uploaded") {
		t.Errorf("got %d %q, want 503 saying nothing was uploaded", w.Code, w.Body.String())
	}
}

func TestDropshareLockErrorIsReported(t *testing.T) {
	h := newFakeHost(t, echoBase)
	os.Mkdir(filepath.Join(h.dir, ".upload.lock"), 0700) // a directory cannot be opened for writing
	w := postUpload(h.file(t, "a.txt", "a"))
	if w.Code != http.StatusInternalServerError {
		t.Errorf("got %d %q, want 500", w.Code, w.Body.String())
	}
}

func TestDropshareCanceledRequest(t *testing.T) {
	ids := make(chan string, 1)
	h := newFakeHost(t, func(id, _ string) string {
		ids <- id
		return ""
	})
	file := h.file(t, "a.txt", "a")

	ctx, cancel := context.WithCancel(context.Background())
	returned := make(chan struct{})
	go func() { postUploadCtx(ctx, file); close(returned) }()
	id := <-ids
	cancel()
	select {
	case <-returned:
	case <-time.After(time.Second):
		t.Fatal("handler kept waiting after the request was canceled")
	}

	// The lock is free at once, not after the 10 s hold.
	lockCtx, lockCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer lockCancel()
	unlock, err := lockDropshareUploads(lockCtx, filepath.Join(h.dir, ".upload.lock"))
	if err != nil {
		t.Fatalf("lock still held after cancel: %v", err)
	}
	unlock()

	inbox := filepath.Join(h.dir, "inbox")
	writeRecord(inbox, id, echoBase(id, file))
	time.Sleep(20 * time.Millisecond)
	if !fileExists(filepath.Join(inbox, id)) {
		t.Error("a callback arriving after cancel was consumed")
	}
}

func swap[T any](p *T, v T) func() {
	old := *p
	*p = v
	return func() { *p = old }
}

// writeRecord stores a callback record the way the applet does, by renaming a
// finished file into place.
func writeRecord(inbox, id, callback string) {
	tmp := filepath.Join(inbox, "."+id+".tmp")
	os.WriteFile(tmp, []byte(callback+"\n"), 0600)
	os.Rename(tmp, filepath.Join(inbox, id))
}

// The upload finishes as soon as the callback arrives, and only then is the
// opener stopped, since Dropshare needs it alive until it has named the caller.
func TestDropshareStopsOpenerAfterCallback(t *testing.T) {
	var h *fakeHost
	var stopped atomic.Bool
	var stoppedBeforeCallback atomic.Bool
	h = newFakeHost(t, echoBase)
	defer swap(&dropshareOpen, func(raw string) (func(), error) {
		id, file := uploadParams(raw)
		go func() {
			time.Sleep(20 * time.Millisecond)
			if stopped.Load() {
				stoppedBeforeCallback.Store(true)
			}
			writeRecord(filepath.Join(h.dir, "inbox"), id, echoBase(id, file))
		}()
		return func() { stopped.Store(true) }, nil
	})()

	start := time.Now()
	if w := postUpload(h.file(t, "a.txt", "a")); w.Code != http.StatusOK {
		t.Fatalf("got %d %q, want 200", w.Code, w.Body.String())
	}
	if d := time.Since(start); d > 500*time.Millisecond {
		t.Errorf("upload took %v, want about the 20ms until the callback", d)
	}
	if !stopped.Load() || stoppedBeforeCallback.Load() {
		t.Errorf("opener stopped=%v, stopped before the callback=%v; want stopped only after it", stopped.Load(), stoppedBeforeCallback.Load())
	}
}
