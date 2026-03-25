package selfupdate

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/inventage-ai/asylum/internal/log"
)

var httpClient = &http.Client{Timeout: 60 * time.Second}

const repo = "inventage-ai/asylum"

type release struct {
	TagName string  `json:"tag_name"`
	Assets  []asset `json:"assets"`
}

type asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// AssetName returns the expected binary asset name for the current platform.
func AssetName() string {
	return fmt.Sprintf("asylum-%s-%s", runtime.GOOS, runtime.GOARCH)
}

// ResolveChannel returns "dev" if the dev flag is set or the config channel is
// "dev", otherwise "stable".
func ResolveChannel(devFlag bool, configChannel string) string {
	if devFlag || configChannel == "dev" {
		return "dev"
	}
	return "stable"
}

// SafeRun is a stripped-down emergency updater. Always pulls the dev release,
// no version checks, no changelog — just download and replace.
func SafeRun(execPath string) error {
	binPath, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("resolve binary path: %w", err)
	}

	rel, err := fetchRelease("dev")
	if err != nil {
		return err
	}

	downloadURL, err := findAssetURL(rel)
	if err != nil {
		return err
	}

	fmt.Println("downloading dev release...")
	if err := downloadAndReplace(downloadURL, binPath); err != nil {
		return err
	}

	fmt.Println("done")
	return nil
}

// Run performs the self-update. It resolves the target binary path by following
// symlinks from execPath, fetches the appropriate release, and atomically
// replaces the binary.
func Run(currentVersion, currentCommit, channel, execPath string) error {
	binPath, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("resolve binary path: %w", err)
	}

	rel, err := fetchRelease(channel)
	if err != nil {
		return err
	}

	version := rel.TagName
	if channel == "stable" && currentVersion == strings.TrimPrefix(version, "v") {
		log.Success("already up to date (%s)", version)
		return nil
	}

	downloadURL, err := findAssetURL(rel)
	if err != nil {
		return err
	}

	log.Info("downloading %s...", version)
	if err := downloadAndReplace(downloadURL, binPath); err != nil {
		return err
	}

	// Report the version baked into the downloaded binary
	newVersion := version
	var newCommit string
	if out, err := exec.Command(binPath, "--version", "--short").Output(); err == nil {
		newVersion = strings.TrimSpace(string(out))
		// Extract commit hash from "dev (abc1234)" format
		if i := strings.Index(newVersion, "("); i != -1 {
			if j := strings.Index(newVersion[i:], ")"); j != -1 {
				newCommit = newVersion[i+1 : i+j]
			}
		}
	}
	log.Success("updated to %s", newVersion)

	if channel == "dev" && currentCommit != "" && newCommit != "" && currentCommit != newCommit {
		showChangelog(currentCommit, newCommit)
	}
	return nil
}

func showChangelog(fromCommit, toCommit string) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/compare/%s...%s", repo, fromCommit, toCommit)
	resp, err := httpClient.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		return
	}
	defer resp.Body.Close()

	var result struct {
		TotalCommits int `json:"total_commits"`
		Commits      []struct {
			Commit struct {
				Message string `json:"message"`
			} `json:"commit"`
		} `json:"commits"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || result.TotalCommits == 0 {
		return
	}

	log.Info("%d commit(s) since %s:", result.TotalCommits, fromCommit)
	// Show up to 5 most recent commits (API returns oldest first)
	commits := result.Commits
	start := 0
	if len(commits) > 5 {
		start = len(commits) - 5
	}
	for _, c := range commits[start:] {
		msg, _, _ := strings.Cut(c.Commit.Message, "\n")
		fmt.Printf("  %s\n", msg)
	}
	if start > 0 {
		fmt.Printf("  ... and %d more\n", start)
	}
}

func findAssetURL(rel release) (string, error) {
	name := AssetName()
	for _, a := range rel.Assets {
		if a.Name == name {
			return a.BrowserDownloadURL, nil
		}
	}
	return "", fmt.Errorf("no asset %q in release %s", name, rel.TagName)
}

func fetchRelease(channel string) (release, error) {
	var url string
	if channel == "dev" {
		url = fmt.Sprintf("https://api.github.com/repos/%s/releases/tags/dev", repo)
	} else {
		url = fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	}

	resp, err := httpClient.Get(url)
	if err != nil {
		return release{}, fmt.Errorf("fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return release{}, fmt.Errorf("fetch release: HTTP %d", resp.StatusCode)
	}

	var rel release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return release{}, fmt.Errorf("parse release: %w", err)
	}
	return rel, nil
}

func downloadAndReplace(url, binPath string) error {
	dir := filepath.Dir(binPath)

	tmp, err := os.CreateTemp(dir, ".asylum-update-*")
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			return fmt.Errorf("no write permission to %s — try: sudo asylum self-update", dir)
		}
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	written, err := downloadToFile(tmp, url)
	tmp.Close()
	if err != nil {
		return err
	}

	if written.contentLength > 0 && written.bytes != written.contentLength {
		return fmt.Errorf("download incomplete: got %d bytes, expected %d", written.bytes, written.contentLength)
	}

	if err := os.Chmod(tmpPath, 0755); err != nil {
		return fmt.Errorf("chmod: %w", err)
	}

	if err := os.Rename(tmpPath, binPath); err != nil {
		if errors.Is(err, os.ErrPermission) {
			return fmt.Errorf("no write permission to %s — try: sudo asylum self-update", binPath)
		}
		return fmt.Errorf("replace binary: %w", err)
	}

	return nil
}

type downloadResult struct {
	bytes         int64
	contentLength int64
}

// maxDownloadSize is 512 MB, well above any expected binary size.
const maxDownloadSize = 512 << 20

func downloadToFile(w io.Writer, url string) (downloadResult, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return downloadResult{}, fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return downloadResult{}, fmt.Errorf("download: HTTP %d", resp.StatusCode)
	}

	if resp.ContentLength <= 0 {
		return downloadResult{}, fmt.Errorf("download: missing or invalid Content-Length")
	}

	if resp.ContentLength > maxDownloadSize {
		return downloadResult{}, fmt.Errorf("download: Content-Length %d exceeds limit of %d", resp.ContentLength, maxDownloadSize)
	}

	reader := io.LimitReader(resp.Body, maxDownloadSize)
	n, err := io.Copy(w, reader)
	if err != nil {
		return downloadResult{}, fmt.Errorf("download: %w", err)
	}
	return downloadResult{bytes: n, contentLength: resp.ContentLength}, nil
}
