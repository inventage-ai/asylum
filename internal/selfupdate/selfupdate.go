package selfupdate

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/inventage-ai/asylum/internal/log"
)

const repo = "inventage-ai/asylum"

type release struct {
	TagName         string  `json:"tag_name"`
	TargetCommitish string  `json:"target_commitish"`
	Assets          []asset `json:"assets"`
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
	if devFlag {
		return "dev"
	}
	if configChannel == "dev" {
		return "dev"
	}
	return "stable"
}

// Run performs the self-update. It resolves the target binary path by following
// symlinks from execPath, fetches the appropriate release, and atomically
// replaces the binary.
func Run(currentVersion, channel, execPath string) error {
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

	name := AssetName()
	var downloadURL string
	for _, a := range rel.Assets {
		if a.Name == name {
			downloadURL = a.BrowserDownloadURL
			break
		}
	}
	if downloadURL == "" {
		return fmt.Errorf("no asset %q in release %s", name, version)
	}

	log.Info("downloading %s...", version)
	if err := downloadAndReplace(downloadURL, binPath); err != nil {
		return err
	}

	if c := shortCommit(rel.TargetCommitish); c != "" {
		log.Success("updated to %s (%s)", version, c)
	} else {
		log.Success("updated to %s", version)
	}
	return nil
}

func shortCommit(commitish string) string {
	if len(commitish) >= 7 {
		return commitish[:7]
	}
	return commitish
}

func fetchRelease(channel string) (release, error) {
	var url string
	if channel == "dev" {
		url = fmt.Sprintf("https://api.github.com/repos/%s/releases/tags/dev", repo)
	} else {
		url = fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	}

	resp, err := http.Get(url)
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

	resp, err := http.Get(url)
	if err != nil {
		tmp.Close()
		return fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		tmp.Close()
		return fmt.Errorf("download: HTTP %d", resp.StatusCode)
	}

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		tmp.Close()
		return fmt.Errorf("download: %w", err)
	}
	tmp.Close()

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
