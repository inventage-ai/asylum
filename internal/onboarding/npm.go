package onboarding

import (
	"os"
	"path/filepath"

	"github.com/inventage-ai/asylum/internal/container"
)

var lockfiles = []struct {
	file string
	cmd  []string
}{
	{"package-lock.json", []string{"npm", "ci"}},
	{"pnpm-lock.yaml", []string{"pnpm", "install", "--frozen-lockfile"}},
	{"yarn.lock", []string{"yarn", "install", "--frozen-lockfile"}},
	{"bun.lock", []string{"bun", "install", "--frozen-lockfile"}},
	{"bun.lockb", []string{"bun", "install", "--frozen-lockfile"}},
}

// NPMTask detects Node.js projects with lockfiles.
type NPMTask struct{}

func (NPMTask) Name() string { return "npm" }

func (NPMTask) Detect(projectDir string) []Workload {
	var workloads []Workload
	for _, nm := range container.FindNodeModulesDirs(projectDir) {
		dir := filepath.Dir(nm)
		for _, lf := range lockfiles {
			lfPath := filepath.Join(dir, lf.file)
			if _, err := os.Stat(lfPath); err == nil {
				rel, _ := filepath.Rel(projectDir, dir)
				if rel == "." {
					rel = filepath.Base(projectDir)
				}
				workloads = append(workloads, Workload{
					Label:      rel,
					Command:    lf.cmd,
					Dir:        dir,
					HashInputs: []string{lfPath},
					Phase:      PostContainer,
				})
				break
			}
		}
	}
	return workloads
}
