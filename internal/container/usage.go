package container

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/inventage-ai/asylum/internal/docker"
)

// Usage is what each session start records about its project in the
// project data directory.
type Usage struct {
	Project  string    `json:"project"`
	LastUsed time.Time `json:"lastUsed"`
	Image    string    `json:"image"`
}

const usageFile = "project.json"

func WriteUsage(dataDir string, u Usage) error {
	data, err := json.Marshal(u)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dataDir, usageFile), data, 0644)
}

// ReadUsage returns the usage record in dataDir, or false when there is none.
func ReadUsage(dataDir string) (Usage, bool) {
	data, err := os.ReadFile(filepath.Join(dataDir, usageFile))
	if err != nil {
		return Usage{}, false
	}
	var u Usage
	if json.Unmarshal(data, &u) != nil {
		return Usage{}, false
	}
	return u, true
}

// ReadAllUsage maps each container name under projectsDir to its usage record.
func ReadAllUsage(projectsDir string) (map[string]Usage, error) {
	entries, err := readDirs(projectsDir)
	if err != nil {
		return nil, err
	}
	records := map[string]Usage{}
	for _, e := range entries {
		if u, ok := ReadUsage(filepath.Join(projectsDir, e.Name())); ok {
			records[e.Name()] = u
		}
	}
	return records, nil
}

// DataDir describes one project data directory under ~/.asylum/projects/.
type DataDir struct {
	ModTime time.Time
	Size    int64
}

// ScanDataDirs maps each container name under projectsDir to the modification
// time and total size of its data directory.
func ScanDataDirs(projectsDir string) (map[string]DataDir, error) {
	entries, err := readDirs(projectsDir)
	if err != nil {
		return nil, err
	}
	dirs := map[string]DataDir{}
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		dirs[e.Name()] = DataDir{ModTime: info.ModTime(), Size: treeSize(filepath.Join(projectsDir, e.Name()))}
	}
	return dirs, nil
}

func readDirs(dir string) ([]fs.DirEntry, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var dirs []fs.DirEntry
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e)
		}
	}
	return dirs, nil
}

func treeSize(dir string) int64 {
	var size int64
	filepath.WalkDir(dir, func(_ string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if info, err := d.Info(); err == nil && info.Mode().IsRegular() {
			size += info.Size()
		}
		return nil
	})
	return size
}

// UnusedProjectImages returns the project images among candidates that no
// remaining usage record and no container refers to. Only project images are
// considered, since the base image is shared by every project.
func UnusedProjectImages(candidates []string, remaining map[string]Usage, containers []docker.ContainerMounts) []string {
	used := map[string]bool{}
	for _, u := range remaining {
		used[u.Image] = true
	}
	for _, c := range containers {
		used[c.Image] = true
	}
	var unused []string
	for _, img := range candidates {
		if strings.HasPrefix(img, "asylum:proj-") && !used[img] && !slices.Contains(unused, img) {
			unused = append(unused, img)
		}
	}
	return unused
}
