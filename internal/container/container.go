package container

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"syscall"

	"github.com/inventage-ai/asylum/internal/agent"
	"github.com/inventage-ai/asylum/internal/config"
	"github.com/inventage-ai/asylum/internal/log"
	"github.com/inventage-ai/asylum/internal/term"
)

var invalidHostnameChars = regexp.MustCompile(`[^a-z0-9-]`)

// defaultCacheDirs is the fallback cache dir map when no profiles are configured.
var defaultCacheDirs = map[string]string{
	"npm":    "/home/claude/.npm",
	"pip":    "/home/claude/.cache/pip",
	"maven":  "/home/claude/.m2",
	"gradle": "/home/claude/.gradle",
}

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
	CacheDirs  map[string]string // tool name → container path
}

func RunArgs(opts RunOpts) ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("home dir: %w", err)
	}

	containerName := ContainerName(opts.ProjectDir)
	hostname := safeHostname(opts.ProjectDir)

	args := []string{
		"run", "-d", "--rm", "--init",
		"--name", containerName,
		"--hostname", hostname,
		"-w", opts.ProjectDir,
	}
	if opts.Config.KitActive("docker") {
		args = append(args, "--privileged")
	}

	args, err = appendVolumes(args, home, containerName, opts)
	if err != nil {
		return nil, err
	}
	args, err = appendEnvVars(args, opts)
	if err != nil {
		return nil, err
	}
	args, err = appendPorts(args, opts.Config.Ports)
	if err != nil {
		return nil, err
	}

	if envFile := filepath.Join(opts.ProjectDir, ".env"); fileExists(envFile) {
		args = append(args, "--env-file", envFile)
	}

	args = append(args, opts.ImageTag)
	args = append(args, "sleep", "infinity")

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

	// Shadow node_modules with named volumes so host OS-specific
	// binaries aren't visible inside the container. Named volumes
	// persist across container restarts so npm install isn't lost.
	if !opts.Config.ShadowNodeModulesOff() {
		for _, nm := range FindNodeModulesDirs(opts.ProjectDir) {
			rel, _ := filepath.Rel(opts.ProjectDir, nm)
			hash := fmt.Sprintf("%x", sha256.Sum256([]byte(rel)))[:11]
			volName := cname + "-npm-" + hash
			args = append(args, "--mount", "type=volume,src="+volName+",dst="+nm)
		}
	}

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

	// Caches (named volumes for better IO on macOS)
	cacheDirs := opts.CacheDirs
	if cacheDirs == nil {
		cacheDirs = defaultCacheDirs
	}
	for _, name := range slices.Sorted(maps.Keys(cacheDirs)) {
		volName := cname + "-cache-" + name
		args = append(args, "--mount", "type=volume,src="+volName+",dst="+cacheDirs[name])
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
		parsed, err := config.ParseVolume(v, home)
		if err != nil {
			return nil, fmt.Errorf("invalid volume %q: %w", v, err)
		}
		vol(parsed.Host, parsed.Container, parsed.Options)
	}

	return args, nil
}

func appendEnvVars(args []string, opts RunOpts) ([]string, error) {
	env := func(k, v string) {
		args = append(args, "-e", k+"="+v)
	}

	for _, k := range slices.Sorted(maps.Keys(opts.Config.Env)) {
		if strings.ContainsAny(opts.Config.Env[k], "\n\r") {
			return nil, fmt.Errorf("env var %q contains newlines, which Docker does not support", k)
		}
		env(k, opts.Config.Env[k])
	}

	if opts.Config.KitActive("docker") {
		env("ASYLUM_DOCKER", "1")
	}
	env("COLORTERM", "truecolor")
	env("TERM", "xterm-256color")
	if !opts.Config.AllowAgentTermTitle() {
		env("CLAUDE_CODE_DISABLE_TERMINAL_TITLE", "1")
	}
	env("HISTFILE", "/home/claude/.shell_history/zsh_history")
	env("HOST_PROJECT_DIR", opts.ProjectDir)

	if java := opts.Config.JavaVersion(); java != "" {
		env("ASYLUM_JAVA_VERSION", java)
	}

	agentEnv := opts.Agent.EnvVars()
	for _, k := range slices.Sorted(maps.Keys(agentEnv)) {
		env(k, agentEnv[k])
	}

	return args, nil
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

type ExecOpts struct {
	ContainerName string
	Mode          Mode
	Agent         agent.Agent
	ProjectDir    string
	ExtraArgs     []string
	NewSession    bool
	Config        config.Config
}

func ExecArgs(opts ExecOpts) []string {
	args := []string{"exec"}
	if term.IsTerminal() {
		args = append(args, "-it")
	} else {
		args = append(args, "-i")
	}
	if opts.Mode == ModeAdminShell {
		args = append(args, "-u", "root")
	}
	args = append(args, opts.ContainerName)

	switch opts.Mode {
	case ModeShell, ModeAdminShell:
		args = append(args, "/bin/zsh")
	case ModeCommand:
		args = append(args, opts.ExtraArgs...)
	case ModeAgent:
		args = append(args, agentCommand(opts)...)
	}
	return args
}

func agentCommand(opts ExecOpts) []string {
	resume := !opts.NewSession && opts.Agent.HasSession(opts.ProjectDir)
	extra := opts.ExtraArgs
	if opts.Config.KitActive("title") && opts.Agent.Name() == "claude" && !resume {
		extra = append([]string{"--name", filepath.Base(opts.ProjectDir)}, extra...)
	}
	return opts.Agent.Command(resume, extra)
}

// EnsureAgentConfig returns true if the config was freshly created (first run).
func EnsureAgentConfig(home string, a agent.Agent) (bool, error) {
	agentDir := config.ExpandTilde(a.AsylumConfigDir(), home)

	if dirExists(agentDir) {
		return false, nil
	}

	nativeDir := a.NativeConfigDir()
	if nativeDir == "" {
		log.Info("creating %s config directory", a.Name())
		return true, os.MkdirAll(agentDir, 0755)
	}
	nativeDir = config.ExpandTilde(nativeDir, home)
	if dirExists(nativeDir) {
		log.Info("seeding %s config from %s", a.Name(), nativeDir)
		return true, copyDir(nativeDir, agentDir)
	}

	log.Info("creating %s config directory", a.Name())
	return true, os.MkdirAll(agentDir, 0755)
}

func copyDir(src, dst string) error {
	return copyDirVisited(src, dst, map[string]bool{})
}

func copyDirVisited(src, dst string, visited map[string]bool) error {
	realSrc, err := filepath.EvalSymlinks(src)
	if err != nil {
		return err
	}
	if visited[realSrc] {
		return nil
	}
	visited[realSrc] = true

	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)

		if d.Type()&fs.ModeSymlink != 0 {
			// Resolve the symlink and copy the target contents instead of
			// recreating the symlink, which may dangle in the destination.
			resolved, err := filepath.EvalSymlinks(path)
			if err != nil {
				// Dangling symlink — skip it.
				return nil
			}
			ri, err := os.Stat(resolved)
			if err != nil {
				return nil
			}
			if ri.IsDir() {
				return copyDirVisited(resolved, target, visited)
			}
			data, err := os.ReadFile(resolved)
			if err != nil {
				return err
			}
			return os.WriteFile(target, data, ri.Mode())
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

var nodeModulesCache struct {
	dir     string
	results []string
}

// FindNodeModulesDirs returns absolute paths to node_modules directories
// that should be shadowed. It finds every directory containing a
// package.json and returns the node_modules path next to it, whether or
// not node_modules exists yet. This ensures fresh clones get shadow
// volumes before npm install runs.
//
// Results are cached per projectDir for the lifetime of the process.
func FindNodeModulesDirs(projectDir string) []string {
	if nodeModulesCache.dir == projectDir {
		return nodeModulesCache.results
	}
	// Directories that never contain relevant package.json files
	skip := map[string]bool{
		".git": true, ".venv": true, "__pycache__": true,
		"vendor": true, "target": true, "dist": true,
	}

	var results []string
	filepath.WalkDir(projectDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		name := d.Name()
		if name == "node_modules" {
			return filepath.SkipDir
		}
		if path != projectDir && skip[name] {
			return filepath.SkipDir
		}
		if fileExists(filepath.Join(path, "package.json")) {
			results = append(results, filepath.Join(path, "node_modules"))
		}
		return nil
	})
	slices.Sort(results)
	nodeModulesCache.dir = projectDir
	nodeModulesCache.results = results
	return results
}

// sessionCounterPath returns the path to the session counter file for a container.
func sessionCounterPath(containerName string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".asylum", "projects", containerName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "sessions"), nil
}

// IncrementSessions atomically increments the session counter and returns the new value.
func IncrementSessions(containerName string) (int, error) {
	path, err := sessionCounterPath(containerName)
	if err != nil {
		return 0, err
	}
	return adjustCounter(path, 1)
}

// DecrementSessions atomically decrements the session counter and returns the new value.
func DecrementSessions(containerName string) (int, error) {
	path, err := sessionCounterPath(containerName)
	if err != nil {
		return 0, err
	}
	n, err := adjustCounter(path, -1)
	if err != nil {
		return 0, err
	}
	if n <= 0 {
		os.Remove(path)
	}
	return n, nil
}

func adjustCounter(path string, delta int) (int, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return 0, fmt.Errorf("open counter: %w", err)
	}
	defer f.Close()

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return 0, fmt.Errorf("lock counter: %w", err)
	}

	data, _ := io.ReadAll(f)
	n, _ := strconv.Atoi(strings.TrimSpace(string(data)))
	n += delta
	if n < 0 {
		n = 0
	}

	if err := f.Truncate(0); err != nil {
		return n, fmt.Errorf("truncate counter: %w", err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		return n, fmt.Errorf("seek counter: %w", err)
	}
	if _, err := f.WriteString(strconv.Itoa(n)); err != nil {
		return n, fmt.Errorf("write counter: %w", err)
	}
	return n, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
