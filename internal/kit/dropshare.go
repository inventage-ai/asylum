package kit

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	_ "embed"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/inventage-ai/asylum/internal/broker"
	"github.com/inventage-ai/asylum/internal/docker"
)

const (
	dropshareKitName  = "dropshare"
	dropshareRoute    = "/dropshare-upload"
	dropshareScheme   = "asylum-dropshare"
	dropshareAppName  = "AsylumDropshare.app"
	dropshareBundleID = "ch.inventage.asylum.dropshare"
)

// dropshareBundleIDs are the Dropshare 5 builds, direct and Setapp.
var dropshareBundleIDs = []string{"net.mkswap.Dropshare5", "net.mkswap.Dropshare-setapp"}

// Host hooks, replaced in tests.
var (
	dropshareSupported = runtime.GOOS == "darwin"
	dropsharePrepare   = prepareDropshareHost
	dropshareCopyOut   = docker.CopyOut
	dropshareOpen      = openDropshareURL
	dropsharePoll      = 200 * time.Millisecond
	dropshareTimeout   = 5 * time.Minute
	dropshareHold      = holdForSize
)

//go:embed dropshare_skill.md
var dropshareSkill string

func init() {
	Register(&Kit{
		Name:           dropshareKitName,
		Description:    "Upload files through Dropshare on the host",
		Tier:           TierOptIn,
		Tools:          []string{"asylum-dropshare"},
		ProvidesSkills: true,
		ConfigSnippet: `  # dropshare:          # Upload files through Dropshare on the host (macOS)
`,
		ConfigNodes:   configNodes(dropshareKitName, "Upload files through Dropshare on the host (macOS)", nil),
		ConfigComment: "dropshare:            # Upload files through Dropshare on the host (macOS)",
		DockerSnippet: dropshareDockerSnippet(),
		RulesSnippet: `### Dropshare (dropshare kit)
Run ` + "`asylum-dropshare <file>`" + ` to publish a file through Dropshare on the user's Mac; it prints the share URL. Read the ` + "`dropshare`" + ` skill before the first upload.
`,
		Routes: []broker.Route{{Path: dropshareRoute, Handler: dropshareHandler}},
	})
}

// dropshareCommand runs in the container. Exit 2 marks a timeout, after which
// the upload may still finish.
const dropshareCommand = `#!/bin/sh
# asylum: uploads a file through Dropshare on the host and prints the share URL.
[ $# -eq 1 ] || { echo "usage: asylum-dropshare <file>" >&2; exit 1; }
case "$1" in /*) p=$1 ;; *) p=$PWD/$1 ;; esac
if [ -n "$ASYLUM_BROKER_SOCK" ]; then
    set -- --unix-socket "$ASYLUM_BROKER_SOCK" http://localhost/dropshare-upload
else
    set -- "http://${ASYLUM_BROKER_HOST}:${ASYLUM_BROKER_PORT}/dropshare-upload"
fi
out=$(mktemp) || exit 1
code=$(curl -sS --max-time 330 -o "$out" -w '%{http_code}' -X POST \
    -H "Authorization: Bearer ${ASYLUM_BROKER_TOKEN}" --data-urlencode "path=$p" "$@")
case "$code" in
    200) cat "$out"; rc=0 ;;
    504) cat "$out" >&2; rc=2 ;;
    *)   if [ -s "$out" ]; then cat "$out" >&2; else echo "asylum-dropshare: broker request failed (HTTP $code)" >&2; fi; rc=1 ;;
esac
rm -f "$out"
exit $rc
`

// dropshareDockerSnippet stages the command and the skill. Both travel as
// base64 because the Dockerfile has no heredoc support.
func dropshareDockerSnippet() string {
	b64 := func(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }
	return `# Dropshare uploads via the asylum host broker.
RUN echo '` + b64(dropshareCommand) + `' | base64 -d | sudo tee /usr/local/bin/asylum-dropshare >/dev/null && \
    sudo chmod +x /usr/local/bin/asylum-dropshare && \
    sudo mkdir -p /opt/asylum-skills/.claude/skills/dropshare && \
    echo '` + b64(dropshareSkill) + `' | base64 -d | sudo tee /opt/asylum-skills/.claude/skills/dropshare/SKILL.md >/dev/null && \
    sudo chown -R "$(id -u):$(id -g)" /opt/asylum-skills
`
}

// dropshareHandler uploads a file from the container through the host's
// Dropshare and responds with the share URL once Dropshare calls back.
func dropshareHandler(ctx broker.Ctx, w http.ResponseWriter, r *http.Request) {
	if !dropshareSupported {
		http.Error(w, "the dropshare kit requires a macOS host", http.StatusNotImplemented)
		return
	}
	path := r.FormValue("path")
	if !filepath.IsAbs(path) {
		http.Error(w, fmt.Sprintf("path %q is not absolute", path), http.StatusBadRequest)
		return
	}
	dir, err := dropsharePrepare()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id, err := dropshareID()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	stage := filepath.Join(dir, "staging", id)
	file, size, err := stageDropshareFile(ctx.Container(), filepath.Clean(path), stage)
	if err != nil {
		os.RemoveAll(stage)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	wait, cancel := context.WithTimeout(r.Context(), dropshareTimeout)
	defer cancel()
	unlock, err := lockDropshareUploads(wait, filepath.Join(dir, ".upload.lock"))
	if err != nil {
		os.RemoveAll(stage)
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			http.Error(w, fmt.Sprintf("Dropshare stayed busy with another upload for %v; nothing was uploaded, try again later", dropshareTimeout), http.StatusServiceUnavailable)
		case wait.Err() == nil:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	stop, err := dropshareOpen(dropshareUploadURL(file, id))
	if err != nil {
		unlock()
		os.RemoveAll(stage)
		http.Error(w, "hand upload to Dropshare: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer stop()
	// The hold counts from the handoff. Past it, a Dropshare failure that
	// never calls back stops blocking other uploads.
	hold := time.AfterFunc(dropshareHold(size), unlock)
	defer func() {
		if hold.Stop() {
			unlock()
		}
	}()

	// Without a callback Dropshare may still be reading the staged copy, so
	// only the prune removes it.
	raw, err := waitDropshareCallback(wait, filepath.Join(dir, "inbox"), id)
	if errors.Is(err, context.DeadlineExceeded) {
		http.Error(w, fmt.Sprintf("Dropshare did not report a link within %v; if the upload finishes, the link will be on the host clipboard", dropshareTimeout), http.StatusGatewayTimeout)
		return
	}
	if err != nil {
		return // the caller went away
	}
	os.RemoveAll(stage)
	gotFile, shareURL, err := parseDropshareCallback(raw)
	if err == nil && gotFile != file {
		err = fmt.Errorf("callback names %q, want %q", gotFile, file)
	}
	if err != nil {
		http.Error(w, "unexpected Dropshare callback: "+err.Error(), http.StatusBadGateway)
		return
	}
	fmt.Fprintln(w, shareURL)
}

var errNotRegular = errors.New("not a regular file; only regular files can be uploaded")

// stageDropshareFile copies path out of the container into stage and returns
// the copy with its size. Docker resolves the path and its symlinks inside
// the container, so only what the container can see is reachable, and
// Dropshare reads a copy the container cannot change. The copy streams as a
// tar archive, so a directory is refused at its first header instead of
// being copied whole.
func stageDropshareFile(container, path, stage string) (string, int64, error) {
	cmd := dropshareCopyOut(container, path)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.StdoutPipe()
	if err != nil {
		return "", 0, err
	}
	if err := cmd.Start(); err != nil {
		return "", 0, err
	}
	file := filepath.Join(stage, filepath.Base(path))
	size, err := writeStagedFile(tar.NewReader(out), file)
	if err != nil {
		cmd.Process.Kill()
	}
	if werr := cmd.Wait(); err == nil {
		err = werr
	}
	switch {
	case errors.Is(err, errNotRegular):
		return "", 0, fmt.Errorf("%s is %w", path, err)
	case err != nil && stderr.Len() > 0:
		return "", 0, errors.New(strings.TrimSpace(stderr.String()))
	case err != nil:
		return "", 0, err
	}
	return file, size, nil
}

func writeStagedFile(tr *tar.Reader, file string) (int64, error) {
	hdr, err := tr.Next()
	if err != nil {
		return 0, err
	}
	if hdr.Typeflag != tar.TypeReg {
		return 0, errNotRegular
	}
	if err := os.MkdirAll(filepath.Dir(file), 0700); err != nil {
		return 0, err
	}
	f, err := os.OpenFile(file, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return 0, err
	}
	n, err := io.Copy(f, tr)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	return n, err
}

// lockDropshareUploads waits for exclusive use of Dropshare. Dropshare drops
// the callback of an upload that starts while another is still transferring,
// so uploads from every project take turns. The kernel releases the lock when
// its holder exits, so a crashed broker cannot leave it stuck.
func lockDropshareUploads(ctx context.Context, path string) (unlock func(), err error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	t := time.NewTicker(dropsharePoll)
	defer t.Stop()
	for {
		err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return func() { f.Close() }, nil
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) {
			f.Close()
			return nil, err
		}
		select {
		case <-ctx.Done():
			f.Close()
			return nil, ctx.Err()
		case <-t.C:
		}
	}
}

// holdForSize is how long an upload may keep Dropshare to itself: long
// enough to transfer the file at a pessimistic 1 MiB/s.
func holdForSize(size int64) time.Duration {
	return 10*time.Second + time.Duration(size>>20)*time.Second
}

// dropshareUploadURL builds the upload action. The callback path carries the
// request ID so the callback finds its way back to this request.
func dropshareUploadURL(file, id string) string {
	return "dropshare5:///action/upload?file=" + escapeAll(file) +
		"&callback=" + escapeAll(dropshareScheme+":///done/"+id)
}

// escapeAll percent-encodes every byte outside the RFC 3986 unreserved set.
func escapeAll(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9' || strings.IndexByte("-._~", c) >= 0 {
			b.WriteByte(c)
		} else {
			fmt.Fprintf(&b, "%%%02X", c)
		}
	}
	return b.String()
}

// parseDropshareCallback extracts the uploaded path and share URL. Dropshare
// path-encodes the file value but leaves & + = raw, so a query parser would
// split filenames apart. url comes last and is generated by Dropshare, which
// makes the last "&url=" the real separator.
func parseDropshareCallback(raw string) (file, shareURL string, err error) {
	_, query, _ := strings.Cut(raw, "?")
	rest, ok := strings.CutPrefix(query, "file=")
	i := strings.LastIndex(rest, "&url=")
	if !ok || i < 0 {
		return "", "", fmt.Errorf("missing file or url in %q", raw)
	}
	if file, err = url.PathUnescape(rest[:i]); err != nil {
		return "", "", err
	}
	shareURL = rest[i+len("&url="):]
	u, err := url.Parse(shareURL)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return "", "", fmt.Errorf("share URL %q is not an absolute https URL", shareURL)
	}
	return file, shareURL, nil
}

func dropshareID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// waitDropshareCallback polls the inbox for the request's callback record and
// removes it once read.
func waitDropshareCallback(ctx context.Context, inbox, id string) (string, error) {
	path := filepath.Join(inbox, id)
	t := time.NewTicker(dropsharePoll)
	defer t.Stop()
	for {
		if b, err := os.ReadFile(path); err == nil {
			os.Remove(path)
			return strings.TrimSpace(string(b)), nil
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-t.C:
		}
	}
}

// pruneOlderThan removes entries of dir older than maxAge: callback records
// no request consumed, and staged copies of uploads that never called back.
func pruneOlderThan(dir string, maxAge time.Duration) {
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if fi, err := e.Info(); err == nil && time.Since(fi.ModTime()) > maxAge {
			os.RemoveAll(filepath.Join(dir, e.Name()))
		}
	}
}

// prepareDropshareHost makes sure the callback applet is current and Dropshare
// is running, and returns the directory holding the inbox, staging area, and
// upload lock.
func prepareDropshareHost() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".asylum", "dropshare")
	if err := ensureDropshareApplet(dir); err != nil {
		return "", fmt.Errorf("install Dropshare callback applet: %w", err)
	}
	for _, sub := range []string{"inbox", "staging"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0700); err != nil {
			return "", err
		}
		pruneOlderThan(filepath.Join(dir, sub), time.Hour)
	}
	return dir, ensureDropshareRunning()
}

// dropshareCallbackScript runs for each callback, with the URL as $1. It
// accepts only well-formed request IDs, so a crafted URL cannot write outside
// the inbox, and renames into place so readers never see a partial record.
const dropshareCallbackScript = `inbox="$(dirname "$0")/inbox"
case "$1" in asylum-dropshare:///done/*) ;; *) exit 0 ;; esac
id=${1#asylum-dropshare:///done/}
id=${id%%[?]*}
case "$id" in *[!0-9a-f]* | '') exit 0 ;; esac
[ ${#id} -eq 32 ] || exit 0
printf '%s\n' "$1" > "$inbox/.$id.tmp" && mv "$inbox/.$id.tmp" "$inbox/$id"
`

// dropshareApplet returns the applet's AppleScript source and Info.plist
// overrides for an install under dir.
func dropshareApplet(dir string) (source string, plist map[string]any) {
	script := filepath.Join(dir, "callback.sh")
	quoted := strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(script)
	source = "on open location theURL\n" +
		"\tdo shell script \"/bin/sh \" & quoted form of \"" + quoted + "\" & \" \" & quoted form of theURL\n" +
		"end open location\n"
	plist = map[string]any{
		"CFBundleIdentifier": dropshareBundleID,
		"LSUIElement":        true,
		"CFBundleURLTypes": []map[string]any{{
			"CFBundleURLName":    "Asylum Dropshare callback",
			"CFBundleURLSchemes": []string{dropshareScheme},
		}},
	}
	return source, plist
}

// dropshareAppletHash identifies the applet definition, so a changed
// definition triggers a rebuild.
func dropshareAppletHash(dir string) string {
	source, plist := dropshareApplet(dir)
	p, _ := json.Marshal(plist) // map keys marshal sorted
	h := sha256.New()
	for _, part := range []string{source, string(p), dropshareCallbackScript} {
		fmt.Fprintf(h, "%d:%s", len(part), part)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// ensureDropshareApplet builds and registers the applet unless the installed
// one matches the current definition. Brokers of several projects can race on
// the first upload, so the build runs under a lock and swaps the finished
// bundle into place.
func ensureDropshareApplet(dir string) error {
	stamp := filepath.Join(dir, ".applet-sha256")
	app := filepath.Join(dir, dropshareAppName)
	want := dropshareAppletHash(dir)
	current := func() bool {
		b, err := os.ReadFile(stamp)
		return err == nil && string(b) == want && dirExists(app)
	}
	if current() {
		return nil
	}
	if err := os.MkdirAll(filepath.Join(dir, "inbox"), 0700); err != nil {
		return err
	}
	lock, err := os.OpenFile(filepath.Join(dir, ".lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer lock.Close()
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	if current() {
		return nil
	}

	if err := os.WriteFile(filepath.Join(dir, "callback.sh"), []byte(dropshareCallbackScript), 0644); err != nil {
		return err
	}
	tmp, err := os.MkdirTemp(dir, ".build-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	source, plist := dropshareApplet(dir)
	src := filepath.Join(tmp, "applet.applescript")
	if err := os.WriteFile(src, []byte(source), 0644); err != nil {
		return err
	}
	built := filepath.Join(tmp, dropshareAppName)
	info := filepath.Join(built, "Contents", "Info.plist")
	steps := [][]string{{"osacompile", "-o", built, src}}
	for _, key := range []string{"CFBundleIdentifier", "LSUIElement", "CFBundleURLTypes"} {
		v, _ := json.Marshal(plist[key])
		steps = append(steps, []string{"plutil", "-replace", key, "-json", string(v), info})
	}
	// Editing Info.plist invalidates osacompile's ad-hoc signature.
	steps = append(steps, []string{"codesign", "--force", "--sign", "-", built})
	for _, s := range steps {
		if out, err := exec.Command(s[0], s[1:]...).CombinedOutput(); err != nil {
			return fmt.Errorf("%s: %v: %s", s[0], err, out)
		}
	}
	if err := os.RemoveAll(app); err != nil {
		return err
	}
	if err := os.Rename(built, app); err != nil {
		return err
	}
	lsregister := "/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister"
	if out, err := exec.Command(lsregister, "-f", app).CombinedOutput(); err != nil {
		return fmt.Errorf("lsregister: %v: %s", err, out)
	}
	return os.WriteFile(stamp, []byte(want), 0600)
}

// openDropshareURL hands a URL action to Dropshare. Dropshare names the caller
// by walking up the sender's process tree, and asks the user to confirm an
// unknown caller. A plain `open` exits before Dropshare looks, so -W keeps it
// alive while the upload runs. For a URL, -W only returns once Dropshare
// quits, so stop, or at the latest a 5 s timer, kills it.
func openDropshareURL(u string) (stop func(), err error) {
	cmd := exec.Command("open", "-W", "-g", u)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	exited := make(chan struct{})
	go func() {
		cmd.Wait()
		close(exited)
	}()
	kill := time.AfterFunc(5*time.Second, func() { cmd.Process.Kill() })
	return func() {
		kill.Stop()
		cmd.Process.Kill()
		<-exited
	}, nil
}

// ensureDropshareRunning launches Dropshare in the background if it is not
// running. A freshly launched Dropshare drops URL actions until it has
// finished starting, hence the wait.
func ensureDropshareRunning() error {
	for _, id := range dropshareBundleIDs {
		if out, _ := exec.Command("lsappinfo", "find", "bundleid="+id).Output(); len(strings.TrimSpace(string(out))) > 0 {
			return nil
		}
	}
	for _, id := range dropshareBundleIDs {
		if exec.Command("open", "-g", "-b", id).Run() == nil {
			time.Sleep(3 * time.Second)
			return nil
		}
	}
	return errors.New("Dropshare is not installed")
}
