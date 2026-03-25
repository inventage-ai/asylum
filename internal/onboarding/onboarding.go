package onboarding

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/inventage-ai/asylum/internal/log"
	"github.com/inventage-ai/asylum/internal/term"
)

// Phase determines when a workload runs relative to the container lifecycle.
type Phase int

const (
	PostContainer Phase = iota
)

// Task detects onboarding workloads for a project.
type Task interface {
	Name() string
	Detect(projectDir string) []Workload
}

// Workload is a single unit of onboarding work.
type Workload struct {
	Label      string   // display name (e.g. "eamportal-view/.../angular")
	Command    []string // e.g. ["pnpm", "install", "--frozen-lockfile"]
	Dir        string   // working directory inside container
	HashInputs []string // paths to files whose hash determines re-run
	Phase      Phase
}

// Opts configures an onboarding run.
type Opts struct {
	ProjectDir    string
	ContainerName string
	ContainerPath string            // resolved PATH from container
	Tasks         []Task
	Onboarding    map[string]bool   // per-task config (e.g. {"npm": false})
}

// State tracks completed onboarding workloads.
type State map[string]map[string]string // task name → label → hash

func statePath(containerName string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".asylum", "projects", containerName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "onboarding.json"), nil
}

func loadState(containerName string) State {
	path, err := statePath(containerName)
	if err != nil {
		return State{}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return State{}
	}
	var s State
	if json.Unmarshal(data, &s) != nil {
		return State{}
	}
	return s
}

func saveState(containerName string, s State) {
	path, err := statePath(containerName)
	if err != nil {
		return
	}
	data, _ := json.MarshalIndent(s, "", "  ")
	os.WriteFile(path, data, 0644)
}

func hashFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h)
}

func hashInputs(paths []string) string {
	h := sha256.New()
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		h.Write(data)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Run orchestrates onboarding: detect → prompt → execute → save state.
func Run(opts Opts) {
	state := loadState(opts.ContainerName)

	var pending []pendingWorkload
	for _, task := range opts.Tasks {
		if !opts.Onboarding[task.Name()] {
			continue
		}
		taskState := state[task.Name()]
		if taskState == nil {
			taskState = map[string]string{}
		}

		for _, w := range task.Detect(opts.ProjectDir) {
			hash := hashInputs(w.HashInputs)
			if taskState[w.Label] == hash {
				continue
			}
			pending = append(pending, pendingWorkload{
				task:     task.Name(),
				workload: w,
				hash:     hash,
			})
		}
	}

	if len(pending) == 0 {
		return
	}

	// Consolidated prompt
	fmt.Println()
	log.Info("Project onboarding:")
	for _, p := range pending {
		fmt.Printf("  - %s (%s)\n", p.workload.Label, strings.Join(p.workload.Command, " "))
	}
	if !term.IsTerminal() {
		log.Warn("skipping onboarding (not a terminal)")
		return
	}
	fmt.Print("Run setup tasks? [Y/n] ")
	var answer string
	fmt.Scanln(&answer)
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(answer)), "n") {
		return
	}

	// Execute
	for _, p := range pending {
		log.Info("running: %s in %s", strings.Join(p.workload.Command, " "), p.workload.Label)
		if err := execInContainer(opts, p.workload); err != nil {
			log.Error("%s failed: %v", p.workload.Label, err)
			continue
		}
		// Save state for this workload
		if state[p.task] == nil {
			state[p.task] = map[string]string{}
		}
		state[p.task][p.workload.Label] = p.hash
		saveState(opts.ContainerName, state)
	}
}

type pendingWorkload struct {
	task     string
	workload Workload
	hash     string
}

func execInContainer(opts Opts, w Workload) error {
	args := []string{"exec"}
	if term.IsTerminal() {
		args = append(args, "-it")
	}
	args = append(args, "-w", w.Dir, opts.ContainerName)
	if opts.ContainerPath != "" {
		// Run through bash so PATH is applied before command lookup
		quoted := term.ShellQuoteArgs(w.Command)
		script := "export PATH=" + term.ShellQuote(opts.ContainerPath) + "; " + strings.Join(quoted, " ")
		args = append(args, "bash", "-c", script)
	} else {
		args = append(args, w.Command...)
	}

	cmd := exec.Command("docker", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// shellQuote is a package-level alias for term.ShellQuote,
// used by tests in this package.
var shellQuote = term.ShellQuote
