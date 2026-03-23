package onboarding

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestHashInputs(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "lockfile")
	os.WriteFile(f, []byte("content-v1"), 0644)

	h1 := hashInputs([]string{f})
	if h1 == "" {
		t.Fatal("expected non-empty hash")
	}

	os.WriteFile(f, []byte("content-v2"), 0644)
	h2 := hashInputs([]string{f})
	if h1 == h2 {
		t.Error("hash should change when file changes")
	}
}

func TestHashInputsMissingFile(t *testing.T) {
	h := hashInputs([]string{"/nonexistent/file"})
	if h == "" {
		t.Error("expected non-empty hash even with missing files")
	}
}

func TestStateLoadSave(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cname := "asylum-test123"

	// Initially empty
	s := loadState(cname)
	if len(s) != 0 {
		t.Errorf("expected empty state, got %v", s)
	}

	// Save and reload
	s["npm"] = map[string]string{"frontend": "abc123"}
	saveState(cname, s)

	s2 := loadState(cname)
	if s2["npm"]["frontend"] != "abc123" {
		t.Errorf("expected abc123, got %q", s2["npm"]["frontend"])
	}
}

func TestStatePathCreatesDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	path, err := statePath("asylum-test456")
	if err != nil {
		t.Fatal(err)
	}

	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("expected directory to exist: %v", err)
	}
	if filepath.Base(path) != "onboarding.json" {
		t.Errorf("expected onboarding.json, got %s", filepath.Base(path))
	}
}

func TestStatePersistsJSON(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cname := "asylum-jsontest"
	s := State{"npm": {"app": "hash1", "lib": "hash2"}}
	saveState(cname, s)

	path, _ := statePath(cname)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var parsed State
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if parsed["npm"]["app"] != "hash1" || parsed["npm"]["lib"] != "hash2" {
		t.Errorf("unexpected state: %v", parsed)
	}
}

type stubTask struct {
	name      string
	workloads []Workload
}

func (s stubTask) Name() string                    { return s.name }
func (s stubTask) Detect(projectDir string) []Workload { return s.workloads }

func TestPendingDetection(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	projectDir := t.TempDir()

	lockfile := filepath.Join(projectDir, "lock")
	os.WriteFile(lockfile, []byte("v1"), 0644)

	task := stubTask{
		name: "test",
		workloads: []Workload{{
			Label:      "app",
			Command:    []string{"install"},
			Dir:        projectDir,
			HashInputs: []string{lockfile},
		}},
	}

	// First run: should have pending workload
	state := loadState("asylum-pending-test")
	taskState := state[task.Name()]
	if taskState == nil {
		taskState = map[string]string{}
	}

	w := task.Detect(projectDir)[0]
	hash := hashInputs(w.HashInputs)
	if taskState[w.Label] == hash {
		t.Error("should be pending on first run")
	}

	// Simulate completion
	state["test"] = map[string]string{"app": hash}
	saveState("asylum-pending-test", state)

	// Second run: should skip
	state2 := loadState("asylum-pending-test")
	if state2["test"]["app"] != hash {
		t.Error("state should persist")
	}

	// Change lockfile: should be pending again
	os.WriteFile(lockfile, []byte("v2"), 0644)
	newHash := hashInputs(w.HashInputs)
	if state2["test"]["app"] == newHash {
		t.Error("should be pending after lockfile change")
	}
}
