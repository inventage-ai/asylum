package container

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/inventage-ai/asylum/internal/agent"
	"github.com/inventage-ai/asylum/internal/config"
	"github.com/inventage-ai/asylum/internal/log"
)

var invalidHostnameChars = regexp.MustCompile(`[^a-z0-9-]`)

type Mode int

const (
	ModeAgent Mode = iota
	ModeShell
	ModeAdminShell
	ModeCommand
)

type RunOpts struct {
	Config     config.Config
	Agent      agent.Agent
	ImageTag   string
	ProjectDir string
	Mode       Mode
	NewSession bool
	ExtraArgs  []string
}

func RunArgs(opts RunOpts) ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("home dir: %w", err)
	}

	containerName := ContainerName(opts.ProjectDir)
	hostname := safeHostname(opts.ProjectDir)

	seeded, err := ensureAgentConfig(home, opts.Agent)
	if err != nil {
		return nil, err
	}
	if seeded {
		opts.NewSession = true
	}

	args := []string{
		"run", "--rm", "-it", "--privileged", "--init",
		"--name", containerName,
		"--hostname", hostname,
		"-w", opts.ProjectDir,
	}

	args, err = appendVolumes(args, home, containerName, opts)
	if err != nil {
		return nil, err
	}
	args = appendEnvVars(args, opts)
	args = appendPorts(args, opts.Config.Ports)

	if envFile := filepath.Join(opts.ProjectDir, ".env"); fileExists(envFile) {
		args = append(args, "--env-file", envFile)
	}

	args = append(args, opts.ImageTag)
	args = append(args, containerCommand(opts)...)

	return args, nil
}

func ContainerName(projectDir string) string {
	h := sha256.Sum256([]byte(projectDir))
	return fmt.Sprintf("asylum-%x", h[:6])
}

func safeHostname(projectDir string) string {
	base := strings.ToLower(filepath.Base(projectDir))
	safe := invalidHostnameChars.ReplaceAllString(base, "-")
	if len(safe) > 56 { // leave room for "asylum-" prefix (7 chars) within 63 total
		safe = safe[:56]
	}
	safe = strings.Trim(safe, "-")
	if safe == "" {
		safe = "project"
	}
	return "asylum-" + safe
}

func appendVolumes(args []string, home, cname string, opts RunOpts) ([]string, error) {
	vol := func(host, container, mode string) {
		mount := host + ":" + container
		if mode != "" {
			mount += ":" + mode
		}
		args = append(args, "-v", mount)
	}

	// Project directory at real path
	vol(opts.ProjectDir, opts.ProjectDir, "z")

	// Gitconfig
	gitconfig := filepath.Join(home, ".gitconfig")
	if fileExists(gitconfig) {
		vol(gitconfig, "/tmp/host_gitconfig", "ro")
	}

	// SSH
	sshDir := filepath.Join(home, ".asylum", "ssh")
	if dirExists(sshDir) {
		vol(sshDir, "/home/claude/.ssh", "rw")
	}

	// Caches
	cacheBase := filepath.Join(home, ".asylum", "cache", cname)
	caches := map[string]string{
		"npm":    "/home/claude/.npm",
		"pip":    "/home/claude/.cache/pip",
		"maven":  "/home/claude/.m2",
		"gradle": "/home/claude/.gradle",
	}
	for _, name := range slices.Sorted(maps.Keys(caches)) {
		hostPath := filepath.Join(cacheBase, name)
		if err := os.MkdirAll(hostPath, 0755); err != nil {
			return nil, fmt.Errorf("create cache dir %s: %w", hostPath, err)
		}
		vol(hostPath, caches[name], "rw")
	}

	// Shell history
	histDir := filepath.Join(home, ".asylum", "projects", cname, "history")
	if err := os.MkdirAll(histDir, 0755); err != nil {
		return nil, fmt.Errorf("create history dir: %w", err)
	}
	vol(histDir, "/home/claude/.shell_history", "rw")

	// Agent config
	agentDir := expandTilde(opts.Agent.AsylumConfigDir(), home)
	vol(agentDir, opts.Agent.ContainerConfigDir(), "")

	// Direnv
	envrc := filepath.Join(opts.ProjectDir, ".envrc")
	if fileExists(envrc) {
		direnvAllow := filepath.Join(home, ".local", "share", "direnv", "allow")
		if dirExists(direnvAllow) {
			vol(direnvAllow, "/tmp/host_direnv_allow", "ro")
		}
	}

	for _, v := range opts.Config.Volumes {
		parsed := config.ParseVolume(v, home)
		vol(parsed.Host, parsed.Container, parsed.Options)
	}

	return args, nil
}

func appendEnvVars(args []string, opts RunOpts) []string {
	env := func(k, v string) {
		args = append(args, "-e", k+"="+v)
	}

	env("ASYLUM_DOCKER", "1")
	env("HISTFILE", "/home/claude/.shell_history/zsh_history")
	env("HOST_PROJECT_DIR", opts.ProjectDir)

	if java := opts.Config.Versions["java"]; java != "" {
		env("ASYLUM_JAVA_VERSION", java)
	}

	for k, v := range opts.Agent.EnvVars() {
		env(k, v)
	}

	return args
}

func appendPorts(args []string, ports []string) []string {
	for _, p := range ports {
		if strings.Contains(p, ":") {
			args = append(args, "-p", p)
		} else {
			args = append(args, "-p", p+":"+p)
		}
	}
	return args
}

func ExecArgs(containerName string, mode Mode, extraArgs []string) []string {
	args := []string{"exec", "-it"}
	if mode == ModeAdminShell {
		args = append(args, "-u", "root")
	}
	args = append(args, containerName)
	switch mode {
	case ModeShell, ModeAdminShell:
		args = append(args, "/bin/zsh")
	case ModeCommand:
		args = append(args, extraArgs...)
	}
	return args
}

func containerCommand(opts RunOpts) []string {
	switch opts.Mode {
	case ModeShell:
		return []string{"/bin/zsh"}
	case ModeAdminShell:
		return []string{"bash", "-c", "echo 'Admin shell - sudo access enabled' && exec /bin/zsh"}
	case ModeCommand:
		return opts.ExtraArgs
	default:
		resume := !opts.NewSession && opts.Agent.HasSession(opts.ProjectDir)
		return opts.Agent.Command(resume, opts.ExtraArgs)
	}
}

// ensureAgentConfig returns true if the config was freshly created (first run).
func ensureAgentConfig(home string, a agent.Agent) (bool, error) {
	agentDir := expandTilde(a.AsylumConfigDir(), home)

	if dirExists(agentDir) {
		return false, nil
	}

	nativeDir := expandTilde(a.NativeConfigDir(), home)
	if dirExists(nativeDir) {
		log.Info("seeding %s config from %s", a.Name(), nativeDir)
		return true, copyDir(nativeDir, agentDir)
	}

	log.Info("creating %s config directory", a.Name())
	return true, os.MkdirAll(agentDir, 0755)
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)

		if d.Type()&fs.ModeSymlink != 0 {
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(link, target)
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		if d.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
}

func expandTilde(path, home string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}
	return path
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
