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
	"strconv"
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
	args, err = appendPorts(args, opts.Config.Ports)
	if err != nil {
		return nil, err
	}

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

	// Git worktree: mount both the worktree gitdir and main repo's .git
	if wtDir, commonDir := resolveGitWorktree(opts.ProjectDir); wtDir != "" {
		vol(wtDir, wtDir, "z")
		if commonDir != wtDir {
			vol(commonDir, commonDir, "z")
		}
	}

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
	agentDir := config.ExpandTilde(opts.Agent.AsylumConfigDir(), home)
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

	for _, k := range slices.Sorted(maps.Keys(opts.Config.Env)) {
		env(k, opts.Config.Env[k])
	}

	env("ASYLUM_DOCKER", "1")
	env("HISTFILE", "/home/claude/.shell_history/zsh_history")
	env("HOST_PROJECT_DIR", opts.ProjectDir)

	if java := opts.Config.Versions["java"]; java != "" {
		env("ASYLUM_JAVA_VERSION", java)
	}

	agentEnv := opts.Agent.EnvVars()
	for _, k := range slices.Sorted(maps.Keys(agentEnv)) {
		env(k, agentEnv[k])
	}

	return args
}

func validPort(s string) bool {
	n, err := strconv.Atoi(s)
	return err == nil && n > 0 && n <= 65535
}

func appendPorts(args []string, ports []string) ([]string, error) {
	for _, p := range ports {
		if strings.Contains(p, ":") {
			parts := strings.SplitN(p, ":", 2)
			if !validPort(parts[0]) || !validPort(parts[1]) {
				return nil, fmt.Errorf("invalid port mapping %q: ports must be between 1 and 65535", p)
			}
			args = append(args, "-p", p)
		} else {
			if !validPort(p) {
				return nil, fmt.Errorf("invalid port %q: must be between 1 and 65535", p)
			}
			args = append(args, "-p", p+":"+p)
		}
	}
	return args, nil
}

func ExecArgs(containerName string, mode Mode, extraArgs []string) ([]string, error) {
	if mode == ModeAgent {
		return nil, fmt.Errorf("exec into running container is not supported for agent mode")
	}
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
	return args, nil
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
	agentDir := config.ExpandTilde(a.AsylumConfigDir(), home)

	if dirExists(agentDir) {
		return false, nil
	}

	nativeDir := config.ExpandTilde(a.NativeConfigDir(), home)
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
			// File may have been deleted between WalkDir and Info; skip it.
			return nil
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

// resolveGitWorktree detects if projectDir is a git worktree and returns
// the worktree-specific gitdir and the common (main repo) gitdir.
// Returns empty strings if the project is not a worktree.
func resolveGitWorktree(projectDir string) (worktreeDir, commonDir string) {
	dotGit := filepath.Join(projectDir, ".git")
	info, err := os.Lstat(dotGit)
	if err != nil || info.IsDir() {
		return "", ""
	}

	data, err := os.ReadFile(dotGit)
	if err != nil {
		return "", ""
	}
	line := strings.TrimSpace(string(data))
	if !strings.HasPrefix(line, "gitdir: ") {
		return "", ""
	}

	gitdir := line[len("gitdir: "):]
	if !filepath.IsAbs(gitdir) {
		gitdir = filepath.Join(projectDir, gitdir)
	}
	gitdir = filepath.Clean(gitdir)

	commonFile := filepath.Join(gitdir, "commondir")
	commonData, err := os.ReadFile(commonFile)
	if err != nil {
		return "", ""
	}
	common := strings.TrimSpace(string(commonData))
	if !filepath.IsAbs(common) {
		common = filepath.Join(gitdir, common)
	}
	common = filepath.Clean(common)

	return gitdir, common
}


func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
