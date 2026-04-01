package container

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"

	"github.com/inventage-ai/asylum/assets"
	"github.com/inventage-ai/asylum/internal/agent"
	"github.com/inventage-ai/asylum/internal/config"
	"github.com/inventage-ai/asylum/internal/kit"
	"github.com/inventage-ai/asylum/internal/log"
	"github.com/inventage-ai/asylum/internal/ports"
	"github.com/inventage-ai/asylum/internal/term"
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
	CacheDirs  map[string]string // tool name → container path
	Kits       []*kit.Kit
	Version    string
}

// RunArgs assembles docker run arguments via a unified RunArg pipeline.
// All sources (core, kits, user config) produce typed RunArgs that are
// collected, deduplicated, and validated before being flattened to []string.
// Returns the flat args for docker run, plus the resolved RunArgs and any
// overrides for debug output.
func RunArgs(opts RunOpts) ([]string, []kit.RunArg, []kit.Override, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("home dir: %w", err)
	}

	containerName := ContainerName(opts.ProjectDir)
	hostname := safeHostname(opts.ProjectDir)

	var all []kit.RunArg

	// Core structural args
	core := func(flag, value string) {
		all = append(all, kit.RunArg{Flag: flag, Value: value, Source: "core", Priority: kit.PriorityCore})
	}
	core("run", "")
	core("-d", "")
	core("--rm", "")
	core("--init", "")
	core("--name", containerName)
	core("--hostname", hostname)
	core("--add-host", "host.docker.internal:host-gateway")
	core("-w", opts.ProjectDir)

	if kit.AnyNeedsMount(opts.Kits) {
		core("--cap-add", "SYS_ADMIN")
	}

	// Core volume mounts
	coreVols, err := coreVolumes(home, containerName, opts)
	if err != nil {
		return nil, nil, nil, err
	}
	all = append(all, coreVols...)

	// Core env vars
	coreEnvs, err := coreEnvVars(home, opts)
	if err != nil {
		return nil, nil, nil, err
	}
	all = append(all, coreEnvs...)

	// .env file
	if envFile := filepath.Join(opts.ProjectDir, ".env"); fileExists(envFile) {
		core("--env-file", envFile)
	}

	// Kit ContainerFunc args (includes ports, docker --privileged, etc.)
	kitArgs := kit.AggregateContainerArgs(opts.Kits, kit.ContainerOpts{
		ProjectDir:    opts.ProjectDir,
		ContainerName: containerName,
		HomeDir:       home,
		Config:        opts.Config,
	})
	all = append(all, kitArgs...)

	// Kit credential mounts
	credArgs, err := kitCredentialArgs(home, containerName, opts)
	if err != nil {
		return nil, nil, nil, err
	}
	all = append(all, credArgs...)

	// Kit volume mounts (non-credential)
	mountArgs := kitMountArgs(home, containerName, opts)
	all = append(all, mountArgs...)

	// User config: ports
	cfgPortArgs, err := configPortArgs(opts.Config.Ports)
	if err != nil {
		return nil, nil, nil, err
	}
	all = append(all, cfgPortArgs...)

	// User config: volumes
	cfgVolArgs, err := configVolumeArgs(opts.Config.Volumes, home)
	if err != nil {
		return nil, nil, nil, err
	}
	all = append(all, cfgVolArgs...)

	// User config: env vars
	for _, k := range slices.Sorted(maps.Keys(opts.Config.Env)) {
		if strings.ContainsAny(opts.Config.Env[k], "\n\r") {
			return nil, nil, nil, fmt.Errorf("env var %q contains newlines, which Docker does not support", k)
		}
		all = append(all, kit.RunArg{Flag: "-e", Value: k + "=" + opts.Config.Env[k], Source: "user config (env)", Priority: kit.PriorityConfig})
	}

	// Sandbox rules for Claude
	if opts.Agent.Name() == "claude" {
		rulesArgs := claudeSandboxRulesArgs(home, containerName, opts, all)
		all = append(all, rulesArgs...)
	}

	// Resolve: dedup, conflict detection
	resolved, overrides, err := ResolveArgs(all)
	if err != nil {
		return nil, nil, nil, err
	}

	// Flatten to []string, then append image + command (not subject to dedup)
	flat := FlattenArgs(resolved)
	flat = append(flat, opts.ImageTag)
	flat = append(flat, "sleep", "infinity")

	return flat, resolved, overrides, nil
}

func ContainerName(projectDir string) string {
	h := sha256.Sum256([]byte(projectDir))
	return fmt.Sprintf("asylum-%x-%s", h[:6], sanitizeProject(projectDir))
}

// OldContainerName returns the pre-migration container name format (hash only,
// no project suffix). Used during migration to find old project directories.
func OldContainerName(projectDir string) string {
	h := sha256.Sum256([]byte(projectDir))
	return fmt.Sprintf("asylum-%x", h[:6])
}

func sanitizeProject(projectDir string) string {
	base := strings.ToLower(filepath.Base(projectDir))
	safe := invalidHostnameChars.ReplaceAllString(base, "-")
	if len(safe) > 56 {
		safe = safe[:56]
	}
	safe = strings.Trim(safe, "-")
	if safe == "" {
		safe = "project"
	}
	return safe
}

func safeHostname(projectDir string) string {
	safe := sanitizeProject(projectDir)
	return "asylum-" + safe
}

// MigrateProjectDir renames old-format project directories
// (asylum-<hash>) to the new format (asylum-<hash>-<project>).
func MigrateProjectDir(projectDir string) error {
	oldName := OldContainerName(projectDir)
	newName := ContainerName(projectDir)
	if oldName == newName {
		return nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	projectsDir := filepath.Join(home, ".asylum", "projects")
	oldDir := filepath.Join(projectsDir, oldName)
	newDir := filepath.Join(projectsDir, newName)

	if _, err := os.Stat(oldDir); err != nil {
		return nil // no old directory to migrate
	}
	if _, err := os.Stat(newDir); err == nil {
		return nil // new directory already exists
	}

	if err := os.Rename(oldDir, newDir); err != nil {
		return fmt.Errorf("migrate project dir: %w", err)
	}

	if err := ports.RenameContainer(oldName, newName); err != nil {
		log.Warn("migrate port allocation: %v", err)
	}

	log.Info("migrated project data from %s to %s", oldName, newName)
	return nil
}

// coreVolumes produces RunArgs for all core volume mounts.
func coreVolumes(home, cname string, opts RunOpts) ([]kit.RunArg, error) {
	var args []kit.RunArg
	vol := func(host, container, mode string) {
		mount := host + ":" + container
		if mode != "" {
			mount += ":" + mode
		}
		args = append(args, kit.RunArg{Flag: "-v", Value: mount, Source: "core", Priority: kit.PriorityCore})
	}
	mnt := func(value string) {
		args = append(args, kit.RunArg{Flag: "--mount", Value: value, Source: "core", Priority: kit.PriorityCore})
	}

	// Project directory at real path
	vol(opts.ProjectDir, opts.ProjectDir, "z")

	// Shadow node_modules with named volumes
	if !opts.Config.ShadowNodeModulesOff() {
		for _, nm := range FindNodeModulesDirs(opts.ProjectDir) {
			rel, _ := filepath.Rel(opts.ProjectDir, nm)
			hash := fmt.Sprintf("%x", sha256.Sum256([]byte(rel)))[:11]
			volName := cname + "-npm-" + hash
			mnt("type=volume,src=" + volName + ",dst=" + nm)
		}
	}

	// Git worktree
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

	// Caches (named volumes for better IO on macOS)
	for _, name := range slices.Sorted(maps.Keys(opts.CacheDirs)) {
		volName := cname + "-cache-" + name
		dst := config.ExpandTilde(opts.CacheDirs[name], home)
		mnt("type=volume,src=" + volName + ",dst=" + dst)
	}

	// Shell history
	histDir := filepath.Join(home, ".asylum", "projects", cname, "history")
	if err := os.MkdirAll(histDir, 0755); err != nil {
		return nil, fmt.Errorf("create history dir: %w", err)
	}
	vol(histDir, filepath.Join(home, ".shell_history"), "rw")

	// Agent config — mount depends on isolation level
	containerConfigDir := config.ExpandTilde(opts.Agent.ContainerConfigDir(), home)
	hostConfigDir, err := agent.ResolveConfigDir(
		opts.Agent,
		opts.Config.AgentIsolation(opts.Agent.Name()),
		cname,
	)
	if err != nil {
		return nil, fmt.Errorf("resolve agent config dir: %w", err)
	}
	os.MkdirAll(hostConfigDir, 0755)
	if resolved, err := filepath.EvalSymlinks(hostConfigDir); err == nil {
		hostConfigDir = resolved
	}
	vol(hostConfigDir, containerConfigDir, "")

	// Direnv
	envrc := filepath.Join(opts.ProjectDir, ".envrc")
	if fileExists(envrc) {
		direnvAllow := filepath.Join(home, ".local", "share", "direnv", "allow")
		if dirExists(direnvAllow) {
			vol(direnvAllow, "/tmp/host_direnv_allow", "ro")
		}
	}

	return args, nil
}

// coreEnvVars produces RunArgs for agent and core environment variables.
func coreEnvVars(home string, opts RunOpts) ([]kit.RunArg, error) {
	var args []kit.RunArg
	env := func(k, v string) {
		args = append(args, kit.RunArg{Flag: "-e", Value: k + "=" + v, Source: "core", Priority: kit.PriorityCore})
	}

	// Agent env vars
	agentEnv := opts.Agent.EnvVars()
	for _, k := range slices.Sorted(maps.Keys(agentEnv)) {
		env(k, agentEnv[k])
	}

	env("COLORTERM", "truecolor")
	env("TERM", "xterm-256color")
	if !opts.Config.AllowAgentTermTitle() {
		env("CLAUDE_CODE_DISABLE_TERMINAL_TITLE", "1")
	}
	env("HISTFILE", filepath.Join(home, ".shell_history", "zsh_history"))
	env("HOST_PROJECT_DIR", opts.ProjectDir)

	if java := opts.Config.JavaVersion(); java != "" {
		env("ASYLUM_JAVA_VERSION", java)
	}

	return args, nil
}

// kitCredentialArgs produces RunArgs for kit credential mounts.
func kitCredentialArgs(home, cname string, opts RunOpts) ([]kit.RunArg, error) {
	var args []kit.RunArg
	credDir := filepath.Join(home, ".asylum", "projects", cname, "credentials")

	for _, k := range opts.Kits {
		if k.CredentialFunc == nil {
			continue
		}
		kitName, _, _ := strings.Cut(k.Name, "/")
		mode := opts.Config.KitCredentialMode(kitName)
		if mode == "none" {
			continue
		}
		if mode == "" {
			if k.Tier == kit.TierAlwaysOn {
				mode = "auto"
			} else {
				continue
			}
		}
		credMode := kit.CredentialAuto
		var explicit []string
		if mode == "explicit" {
			credMode = kit.CredentialExplicit
			explicit = opts.Config.KitCredentialExplicit(kitName)
		}
		var isolation string
		if kc := opts.Config.KitOption(kitName); kc != nil {
			isolation = kc.Isolation
		}
		mounts, err := k.CredentialFunc(kit.CredentialOpts{
			ProjectDir:    opts.ProjectDir,
			HomeDir:       home,
			ContainerName: cname,
			Isolation:     isolation,
			Mode:          credMode,
			Explicit:      explicit,
		})
		if err != nil {
			log.Warn("credentials for %s: %v", k.Name, err)
			continue
		}
		source := k.Name + " kit (credentials)"
		for _, m := range mounts {
			dst := config.ExpandTilde(m.Destination, home)
			volMode := "ro"
			if m.Writable {
				volMode = ""
			}
			if m.HostPath != "" {
				mount := m.HostPath + ":" + dst
				if volMode != "" {
					mount += ":" + volMode
				}
				args = append(args, kit.RunArg{Flag: "-v", Value: mount, Source: source, Priority: kit.PriorityKit})
				continue
			}
			if err := os.MkdirAll(credDir, 0755); err != nil {
				return nil, fmt.Errorf("create credentials dir: %w", err)
			}
			if m.FileName != "" {
				subDir := filepath.Join(credDir, filepath.Base(dst))
				if err := os.MkdirAll(subDir, 0755); err != nil {
					return nil, fmt.Errorf("create credential subdir: %w", err)
				}
				if err := os.WriteFile(filepath.Join(subDir, m.FileName), m.Content, 0600); err != nil {
					return nil, fmt.Errorf("write credential file: %w", err)
				}
				mount := subDir + ":" + dst
				if volMode != "" {
					mount += ":" + volMode
				}
				args = append(args, kit.RunArg{Flag: "-v", Value: mount, Source: source, Priority: kit.PriorityKit})
			} else {
				filename := filepath.Base(dst)
				hostPath := filepath.Join(credDir, filename)
				if err := os.WriteFile(hostPath, m.Content, 0600); err != nil {
					return nil, fmt.Errorf("write credential file: %w", err)
				}
				mount := hostPath + ":" + dst
				if volMode != "" {
					mount += ":" + volMode
				}
				args = append(args, kit.RunArg{Flag: "-v", Value: mount, Source: source, Priority: kit.PriorityKit})
			}
		}
	}
	return args, nil
}

// kitMountArgs produces RunArgs for kit volume mounts (non-credential).
func kitMountArgs(home, cname string, opts RunOpts) []kit.RunArg {
	var args []kit.RunArg
	for _, k := range opts.Kits {
		if k.MountFunc == nil {
			continue
		}
		kitName, _, _ := strings.Cut(k.Name, "/")
		var isolation string
		if kc := opts.Config.KitOption(kitName); kc != nil {
			isolation = kc.Isolation
		}
		mounts, err := k.MountFunc(kit.CredentialOpts{
			HomeDir:       home,
			ContainerName: cname,
			Isolation:     isolation,
		})
		if err != nil {
			log.Warn("mounts for %s: %v", k.Name, err)
			continue
		}
		source := k.Name + " kit (mounts)"
		for _, m := range mounts {
			dst := config.ExpandTilde(m.Destination, home)
			mode := "ro"
			if m.Writable {
				mode = ""
			}
			mount := m.HostPath + ":" + dst
			if mode != "" {
				mount += ":" + mode
			}
			args = append(args, kit.RunArg{Flag: "-v", Value: mount, Source: source, Priority: kit.PriorityKit})
		}
	}
	return args
}

// configPortArgs produces RunArgs for user-configured port mappings.
func configPortArgs(cfgPorts []string) ([]kit.RunArg, error) {
	var args []kit.RunArg
	for _, p := range cfgPorts {
		if strings.Contains(p, ":") {
			parts := strings.SplitN(p, ":", 2)
			if !validPort(parts[0]) || !validPort(parts[1]) {
				return nil, fmt.Errorf("invalid port mapping %q: ports must be between 1 and 65535", p)
			}
			args = append(args, kit.RunArg{Flag: "-p", Value: p, Source: "user config (ports)", Priority: kit.PriorityConfig})
		} else {
			if !validPort(p) {
				return nil, fmt.Errorf("invalid port %q: must be between 1 and 65535", p)
			}
			args = append(args, kit.RunArg{Flag: "-p", Value: p + ":" + p, Source: "user config (ports)", Priority: kit.PriorityConfig})
		}
	}
	return args, nil
}

// configVolumeArgs produces RunArgs for user-configured volume mounts.
func configVolumeArgs(volumes []string, home string) ([]kit.RunArg, error) {
	var args []kit.RunArg
	for _, v := range volumes {
		parsed, err := config.ParseVolume(v, home)
		if err != nil {
			return nil, fmt.Errorf("invalid volume %q: %w", v, err)
		}
		mount := parsed.Host + ":" + parsed.Container
		if parsed.Options != "" {
			mount += ":" + parsed.Options
		}
		args = append(args, kit.RunArg{Flag: "-v", Value: mount, Source: "user config (volumes)", Priority: kit.PriorityConfig})
	}
	return args, nil
}

// claudeSandboxRulesArgs generates sandbox rules and returns the volume mount RunArgs.
// It extracts allocated ports from the collected RunArgs to include in the rules.
func claudeSandboxRulesArgs(home, containerName string, opts RunOpts, collected []kit.RunArg) []kit.RunArg {
	// Pre-create mountpoints for Docker/runc compatibility
	hostConfigDir, err := agent.ResolveConfigDir(
		opts.Agent,
		opts.Config.AgentIsolation(opts.Agent.Name()),
		containerName,
	)
	if err == nil {
		rulesSubdir := filepath.Join(hostConfigDir, "rules")
		os.MkdirAll(rulesSubdir, 0755)
		ensureMountpoint(filepath.Join(rulesSubdir, "asylum-sandbox.md"))
		ensureMountpoint(filepath.Join(hostConfigDir, "asylum-reference.md"))
	}

	// Extract allocated ports from kit-produced -p args
	var allocatedPorts []int
	for _, a := range collected {
		if a.Flag == "-p" && a.Source == "ports kit" {
			// Value is "port:port", extract the container port
			if i := strings.LastIndex(a.Value, ":"); i >= 0 {
				if p, err := strconv.Atoi(a.Value[i+1:]); err == nil {
					allocatedPorts = append(allocatedPorts, p)
				}
			}
		}
	}

	rulesDir, err := generateSandboxRules(home, containerName, opts.Kits, opts.Version, allocatedPorts)
	if err != nil {
		log.Warn("could not generate sandbox rules: %v", err)
		return nil
	}

	containerClaude := config.ExpandTilde(opts.Agent.ContainerConfigDir(), home)
	return []kit.RunArg{
		{Flag: "-v", Value: filepath.Join(rulesDir, "asylum-sandbox.md") + ":" + filepath.Join(containerClaude, "rules", "asylum-sandbox.md") + ":ro", Source: "core", Priority: kit.PriorityCore},
		{Flag: "-v", Value: filepath.Join(rulesDir, "asylum-reference.md") + ":" + filepath.Join(containerClaude, "asylum-reference.md") + ":ro", Source: "core", Priority: kit.PriorityCore},
	}
}

func validPort(s string) bool {
	n, err := strconv.Atoi(s)
	return err == nil && n > 0 && n <= 65535
}

const sandboxRulesTemplate = `# Asylum Sandbox (v%s)

You are running inside an Asylum Docker container (Debian bookworm, %s). Do not install system packages via apt — they won't survive container rebuilds. If a tool is missing, tell the user to add it as a kit or custom build step in ` + "`.asylum`" + `.

For detailed documentation, troubleshooting, and config reference, read ~/.claude/asylum-reference.md

## Environment
- User: %s (with passwordless sudo)
- Host machine: reachable at host.docker.internal
- Default shell: zsh (oh-my-zsh)
- Project directory: mounted from the host at its real path

## Base Tools (always available)
git, docker, curl, wget, jq, yq, ripgrep (rg), fd, direnv, make, cmake, gcc/g++, vim, nano, htop, zip/unzip, ssh

## Package & Version Managers
- fnm: Node.js version manager — ` + "`fnm use <version>`" + `
- uv: Python package installer and venv manager
- mise: Runtime version manager (Java, etc.) — ` + "`mise use java@<version>`" + `
`

// generateSandboxRules writes the rules file and reference doc to
// ~/.asylum/projects/<container>/ and returns the directory path.
func generateSandboxRules(home, containerName string, kits []*kit.Kit, version string, allocatedPorts []int) (string, error) {
	dir := filepath.Join(home, ".asylum", "projects", containerName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create rules dir: %w", err)
	}

	var b strings.Builder
	username := "claude"
	if u, err := user.Current(); err == nil {
		username = u.Username
	}
	fmt.Fprintf(&b, sandboxRulesTemplate, version, runtime.GOARCH, username)

	if tools := kit.AggregateTools(kits); len(tools) > 0 {
		b.WriteString("\n## Kit Tools\n")
		b.WriteString(strings.Join(tools, ", "))
		b.WriteByte('\n')
	}

	if len(allocatedPorts) > 0 {
		b.WriteString("\n## Forwarded Ports\n")
		b.WriteString("These ports are forwarded from the container to the host. ")
		b.WriteString("Start web servers on any of these ports and the user can access them at http://localhost:<port>\n")
		for _, p := range allocatedPorts {
			fmt.Fprintf(&b, "- %d\n", p)
		}
	}

	if kitSnippets := kit.AssembleRulesSnippets(kits); kitSnippets != "" {
		b.WriteString("\n## Active Kits\n\n")
		b.WriteString(kitSnippets)
	}

	// List kits that are available but not active.
	active := map[string]bool{}
	for _, k := range kits {
		active[k.Name] = true
	}
	var disabled []string
	for _, name := range kit.All() {
		if !active[name] {
			if k := kit.Get(name); k != nil {
				disabled = append(disabled, fmt.Sprintf("- **%s** — %s", k.Name, k.Description))
			}
		}
	}
	if len(disabled) > 0 {
		b.WriteString("\n## Disabled Kits\n")
		b.WriteString("These kits are available but not active. See ~/.claude/asylum-reference.md for how to enable them.\n")
		for _, line := range disabled {
			b.WriteString(line)
			b.WriteByte('\n')
		}
	}

	if err := os.WriteFile(filepath.Join(dir, "asylum-sandbox.md"), []byte(b.String()), 0644); err != nil {
		return "", fmt.Errorf("write rules: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "asylum-reference.md"), assets.AsylumReference, 0644); err != nil {
		return "", fmt.Errorf("write reference: %w", err)
	}
	return dir, nil
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
	configDir, err := agent.ResolveConfigDir(
		opts.Agent,
		opts.Config.AgentIsolation(opts.Agent.Name()),
		opts.ContainerName,
	)
	resume := err == nil && !opts.NewSession && opts.Agent.HasSession(configDir, opts.ProjectDir)
	return opts.Agent.Command(resume, opts.ExtraArgs)
}

// EnsureAgentConfig returns true if the config was freshly created (first run).
// EnsureAgentConfigAt ensures the agent config exists at the given target directory,
// seeding from the host native config if available.
func EnsureAgentConfigAt(home string, a agent.Agent, targetDir string) (bool, error) {
	if dirExists(targetDir) {
		return false, nil
	}

	nativeDir := a.NativeConfigDir()
	if nativeDir == "" {
		log.Info("creating %s config directory", a.Name())
		return true, os.MkdirAll(targetDir, 0755)
	}
	nativeDir = config.ExpandTilde(nativeDir, home)
	if dirExists(nativeDir) {
		log.Info("seeding %s config from %s", a.Name(), nativeDir)
		return true, copyDir(nativeDir, targetDir)
	}

	log.Info("creating %s config directory", a.Name())
	return true, os.MkdirAll(targetDir, 0755)
}

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
		".git": true, ".claude": true, ".venv": true, "__pycache__": true,
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

// ensureMountpoint ensures path exists as a regular file so that Docker
// can bind-mount a file on top of it. Without this, runc must create the
// mountpoint itself, which fails on some Docker versions when the path
// resolves through a VirtioFS-backed bind mount (outside the overlay rootfs).
func ensureMountpoint(path string) {
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		os.RemoveAll(path)
	}
	if !fileExists(path) {
		os.WriteFile(path, nil, 0644)
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
