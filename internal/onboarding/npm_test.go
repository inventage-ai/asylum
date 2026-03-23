package onboarding

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNPMTaskDetect(t *testing.T) {
	tests := []struct {
		name     string
		lockfile string
		wantCmd  string
	}{
		{"npm", "package-lock.json", "npm"},
		{"pnpm", "pnpm-lock.yaml", "pnpm"},
		{"yarn", "yarn.lock", "yarn"},
		{"bun", "bun.lock", "bun"},
		{"bun binary", "bun.lockb", "bun"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0644)
			os.WriteFile(filepath.Join(dir, tt.lockfile), []byte("lock"), 0644)

			task := NPMTask{}
			workloads := task.Detect(dir)
			if len(workloads) != 1 {
				t.Fatalf("got %d workloads, want 1", len(workloads))
			}
			if workloads[0].Command[0] != tt.wantCmd {
				t.Errorf("command = %q, want %q", workloads[0].Command[0], tt.wantCmd)
			}
		})
	}
}

func TestNPMTaskDetectNoLockfile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0644)

	task := NPMTask{}
	workloads := task.Detect(dir)
	if len(workloads) != 0 {
		t.Errorf("got %d workloads, want 0 (no lockfile)", len(workloads))
	}
}

func TestNPMTaskDetectNoPackageJSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package-lock.json"), []byte("lock"), 0644)

	task := NPMTask{}
	workloads := task.Detect(dir)
	if len(workloads) != 0 {
		t.Errorf("got %d workloads, want 0 (no package.json)", len(workloads))
	}
}

func TestNPMTaskDetectMonorepo(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(dir, "package-lock.json"), []byte("root-lock"), 0644)

	pkgDir := filepath.Join(dir, "packages", "app")
	os.MkdirAll(pkgDir, 0755)
	os.WriteFile(filepath.Join(pkgDir, "package.json"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(pkgDir, "package-lock.json"), []byte("app-lock"), 0644)

	task := NPMTask{}
	workloads := task.Detect(dir)
	if len(workloads) != 2 {
		t.Fatalf("got %d workloads, want 2", len(workloads))
	}
}

func TestNPMTaskDetectHashInput(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(dir, "pnpm-lock.yaml"), []byte("lock-content"), 0644)

	task := NPMTask{}
	workloads := task.Detect(dir)
	if len(workloads) != 1 {
		t.Fatalf("got %d workloads, want 1", len(workloads))
	}
	if len(workloads[0].HashInputs) != 1 {
		t.Fatalf("expected 1 hash input, got %d", len(workloads[0].HashInputs))
	}
	if filepath.Base(workloads[0].HashInputs[0]) != "pnpm-lock.yaml" {
		t.Errorf("hash input = %q, want pnpm-lock.yaml", workloads[0].HashInputs[0])
	}
}

func TestNPMTaskName(t *testing.T) {
	task := NPMTask{}
	if task.Name() != "npm" {
		t.Errorf("name = %q, want %q", task.Name(), "npm")
	}
}
