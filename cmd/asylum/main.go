package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/inventage-ai/asylum/internal/agent"
	"github.com/inventage-ai/asylum/internal/config"
	"github.com/inventage-ai/asylum/internal/firstrun"
	"github.com/inventage-ai/asylum/internal/container"
	"github.com/inventage-ai/asylum/internal/docker"
	"github.com/inventage-ai/asylum/internal/image"
	"github.com/inventage-ai/asylum/internal/log"
	"github.com/inventage-ai/asylum/internal/onboarding"
	"github.com/inventage-ai/asylum/internal/profile"
	"github.com/inventage-ai/asylum/internal/selfupdate"
	"github.com/inventage-ai/asylum/internal/ssh"
)

var version = "dev"
var commit = ""

func die(format string, args ...any) {
	log.Error(format, args...)
	os.Exit(1)
}

func main() {
	flags, subcommand, extraArgs, err := parseArgs(os.Args[1:])
	if err != nil {
		die("%v", err)
	}

	if flags.Help {
		printUsage()
		return
	}

	if flags.Version {
		fmt.Print(versionString(flags.Short))
		return
	}

	if flags.Cleanup {
		runCleanup()
		return
	}

	if subcommand == "ssh-init" {
		if err := ssh.Init(); err != nil {
			die("%v", err)
		}
		return
	}

	projectDir, err := filepath.Abs(".")
	if err != nil {
		die("resolve project dir: %v", err)
	}

	if subcommand == "self-update" {
		execPath, err := os.Executable()
		if err != nil {
			die("resolve executable: %v", err)
		}
		if flags.Safe {
			if err := selfupdate.SafeRun(execPath); err != nil {
				die("%v", err)
			}
			return
		}
		cfg, err := config.Load(projectDir, config.CLIFlags{})
		if err != nil {
			die("load config: %v", err)
		}
		channel := selfupdate.ResolveChannel(flags.Dev, cfg.ReleaseChannel)
		if err := selfupdate.Run(version, commit, channel, execPath); err != nil {
			die("%v", err)
		}
		return
	}

	containerMode := resolveMode(subcommand, flags.Admin)

	home, err := os.UserHomeDir()
	if err != nil {
		die("home dir: %v", err)
	}

	if err := firstrun.Run(home); err != nil {
		die("first-run setup: %v", err)
	}

	cfg, err := config.Load(projectDir, config.CLIFlags{
		Agent:    flags.Agent,
		Profiles: flags.Profiles,
		Ports:    flags.Ports,
		Volumes:  flags.Volumes,
		Env:      flags.Env,
		Java:     flags.Java,
	})
	if err != nil {
		die("load config: %v", err)
	}

	// Resolve all active profiles from final merged config
	allProfiles, err := profile.Resolve(cfg.Profiles)
	if err != nil {
		die("%v", err)
	}

	// Apply profile config defaults UNDER user config (user wins).
	var profileDefaults config.Config
	for _, p := range allProfiles {
		profileDefaults = config.Merge(profileDefaults, p.Config)
	}
	cfg = config.Merge(profileDefaults, cfg)

	// Resolve global-tier profiles (from ~/.asylum/config.yaml only) for base image.
	// Project-only profiles (in final set but not global) go into the project image.
	globalProfiles, projectProfiles := resolveProfileTiers(projectDir, allProfiles)

	cacheDirs := profile.AggregateCacheDirs(allProfiles)

	agentName := cfg.Agent
	if agentName == "" {
		agentName = "claude"
	}

	a, err := agent.Get(agentName)
	if err != nil {
		die("%v", err)
	}

	if err := docker.DockerAvailable(); err != nil {
		die("%v", err)
	}

	cname := container.ContainerName(projectDir)
	newSession := flags.New
	freshContainer := false

	// If no container running, build images and start one detached
	if !docker.IsRunning(cname) {
		seeded, err := container.EnsureAgentConfig(home, a)
		if err != nil {
			die("%v", err)
		}
		if seeded {
			newSession = true
		}

		baseRebuilt, err := image.EnsureBase(globalProfiles, version, flags.Rebuild)
		if err != nil {
			die("%v", err)
		}

		imageTag, err := image.EnsureProject(projectProfiles, cfg.Packages, cfg.Versions["java"], version, baseRebuilt, flags.Rebuild)
		if err != nil {
			die("%v", err)
		}

		runArgs, err := container.RunArgs(container.RunOpts{
			Config:     cfg,
			Agent:      a,
			ImageTag:   imageTag,
			ProjectDir: projectDir,
			CacheDirs:  cacheDirs,
		})
		if err != nil {
			die("%v", err)
		}

		if err := docker.RunDetached(runArgs); err != nil {
			die("start container: %v", err)
		}

		// Wait for the entrypoint to finish (sleep becomes the main process)
		if !docker.WaitReady(cname, 60) {
			docker.ShowLogs(cname)
			docker.RemoveContainer(cname)
			die("container failed to start")
		}
		freshContainer = true

		// Fix ownership of shadow node_modules volumes (Docker creates them as root)
		if !cfg.FeatureOff("shadow-node-modules") {
			for _, nm := range container.FindNodeModulesDirs(projectDir) {
				docker.Exec(cname, "root", "chown", "claude", nm)
			}
		}

		// Migrate old bind-mounted caches to named volumes (temporary)
		oldCacheBase := filepath.Join(home, ".asylum", "cache", cname)
		for tool, dst := range cacheDirs {
			oldDir := filepath.Join(oldCacheBase, tool)
			if info, err := os.Stat(oldDir); err == nil && info.IsDir() {
				if err := docker.CopyTo(cname, oldDir, dst); err != nil {
					log.Error("migrate %s cache: %v", tool, err)
				}
			}
		}
		os.RemoveAll(oldCacheBase)
	}

	// Write session marker for agent mode
	if containerMode == container.ModeAgent {
		if c, ok := a.(interface{ WriteMarker(string) error }); ok {
			if err := c.WriteMarker(projectDir); err != nil {
				log.Error("write session marker: %v", err)
			}
		}
	}

	// Run onboarding tasks (agent mode, first container start only)
	if containerMode == container.ModeAgent && freshContainer && !flags.SkipOnboarding && !cfg.FeatureOff("onboarding") {
		containerPath, _ := docker.ReadFile(cname, "/tmp/asylum-path")
		onboarding.Run(onboarding.Opts{
			ProjectDir:    projectDir,
			ContainerName: cname,
			ContainerPath: containerPath,
			Tasks:         profile.AggregateOnboardingTasks(allProfiles),
			Onboarding:    cfg.Onboarding,
		})
	}

	// Exec session into the running container
	execArgs := container.ExecArgs(container.ExecOpts{
		ContainerName: cname,
		Mode:          containerMode,
		Agent:         a,
		ProjectDir:    projectDir,
		ExtraArgs:     extraArgs,
		NewSession:    newSession,
		Config:        cfg,
	})

	if _, err := container.IncrementSessions(cname); err != nil {
		log.Error("track session: %v", err)
	}

	setTabTitle(cfg.TabTitle, projectDir, agentName, containerMode)
	exitCode := runDocker(execArgs)

	// Cleanup: remove container if this was the last session
	remaining, err := container.DecrementSessions(cname)
	if err != nil {
		log.Error("track session: %v", err)
	}
	if remaining <= 0 {
		docker.RemoveContainer(cname)
	}

	os.Exit(exitCode)
}

type cliFlags struct {
	Agent    string
	Profiles *[]string
	Ports    []string
	Volumes  []string
	Env      map[string]string
	Java     string
	New      bool
	Rebuild bool
	Cleanup bool
	Help    bool
	Version bool
	Short   bool
	Admin          bool
	Dev            bool
	SkipOnboarding bool
	Safe    bool
}

func parseArgs(args []string) (cliFlags, string, []string, error) {
	var flags cliFlags
	var subcommand string
	var extraArgs []string

	i := 0
	next := func(flag string) (string, error) {
		if i+1 < len(args) {
			val := args[i+1]
			if strings.HasPrefix(val, "-") {
				return "", fmt.Errorf("flag %q requires a value, got %q", flag, val)
			}
			i += 2
			return val, nil
		}
		return "", fmt.Errorf("flag %q requires a value", flag)
	}

	for i < len(args) {
		arg := args[i]

		if arg == "--" {
			extraArgs = append(extraArgs, args[i+1:]...)
			break
		}

		var err error
		switch {
		case arg == "-a" || arg == "--agent":
			flags.Agent, err = next(arg)
		case strings.HasPrefix(arg, "-a") && len(arg) > 2 && arg[2] != '-':
			flags.Agent = arg[2:]
			i++
		case arg == "-p":
			var p string
			if p, err = next(arg); err == nil {
				flags.Ports = append(flags.Ports, p)
			}
		case arg == "-v":
			var v string
			if v, err = next(arg); err == nil {
				flags.Volumes = append(flags.Volumes, v)
			}
		case arg == "-e":
			var val string
			if val, err = next(arg); err == nil {
				k, v, ok := strings.Cut(val, "=")
				if !ok || k == "" {
					err = fmt.Errorf("invalid env var %q: must be KEY=VALUE", val)
				} else {
					if flags.Env == nil {
						flags.Env = make(map[string]string)
					}
					flags.Env[k] = v
				}
			}
		case arg == "--java":
			flags.Java, err = next(arg)
		case arg == "--profiles":
			var val string
			if val, err = next(arg); err == nil {
				p := strings.Split(val, ",")
				flags.Profiles = &p
			}
		case arg == "-n" || arg == "--new":
			flags.New = true
			i++
		case arg == "--rebuild":
			flags.Rebuild = true
			i++
		case arg == "--cleanup":
			flags.Cleanup = true
			i++
		case arg == "--skip-onboarding":
			flags.SkipOnboarding = true
			i++
		case arg == "--version":
			flags.Version = true
			i++
			if i < len(args) && args[i] == "--short" {
				flags.Short = true
				i++
			}
		case arg == "-h" || arg == "--help":
			flags.Help = true
			i++
		case arg == "shell":
			subcommand = "shell"
			i++
			for i < len(args) {
				if args[i] == "--admin" {
					flags.Admin = true
					i++
				} else {
					return cliFlags{}, "", nil, fmt.Errorf("unknown flag %q for shell (only --admin is supported)", args[i])
				}
			}
		case arg == "ssh-init":
			subcommand = "ssh-init"
			i++
			if i < len(args) {
				return cliFlags{}, "", nil, fmt.Errorf("unexpected argument %q after ssh-init", args[i])
			}
		case arg == "self-update":
			subcommand = "self-update"
			i++
			for i < len(args) {
				switch args[i] {
				case "--dev":
					flags.Dev = true
				case "--safe":
					flags.Safe = true
				default:
					return cliFlags{}, "", nil, fmt.Errorf("unknown flag %q for self-update", args[i])
				}
				i++
			}
		case arg == "run":
			subcommand = "run"
			rest := args[i+1:]
			if len(rest) > 0 && rest[0] == "--" {
				rest = rest[1:]
			}
			if len(rest) == 0 {
				return cliFlags{}, "", nil, fmt.Errorf("'run' requires a command (e.g., asylum run ls -la)")
			}
			extraArgs = append(extraArgs, rest...)
			i = len(args)
		case strings.HasPrefix(arg, "-"):
			return cliFlags{}, "", nil, fmt.Errorf("unknown flag %q", arg)
		default:
			return cliFlags{}, "", nil, fmt.Errorf("unexpected argument %q (use 'run' to execute commands in the container)", arg)
		}
		if err != nil {
			return cliFlags{}, "", nil, err
		}
	}

	return flags, subcommand, extraArgs, nil
}

func versionString(short bool) string {
	v := version
	if commit != "" {
		c := commit
		if len(c) > 7 {
			c = c[:7]
		}
		v += " (" + c + ")"
	}
	if short {
		return v + "\n"
	}
	return "asylum " + v + "\n"
}

const defaultTabTitle = "🤖 {project}"

func setTabTitle(template, projectDir, agent string, mode container.Mode) {
	if template == "" {
		template = defaultTabTitle
	}
	modeName := "agent"
	switch mode {
	case container.ModeShell:
		modeName = "shell"
	case container.ModeAdminShell:
		modeName = "admin"
	case container.ModeCommand:
		modeName = "run"
	}
	r := strings.NewReplacer(
		"{project}", filepath.Base(projectDir),
		"{agent}", agent,
		"{mode}", modeName,
	)
	fmt.Printf("\033]0;%s\007", r.Replace(template))
}

func runDocker(args []string) int {
	cmd := exec.Command("docker", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Forward signals to the docker process
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for sig := range sigCh {
			if cmd.Process != nil {
				cmd.Process.Signal(sig)
			}
		}
	}()

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		die("docker exec: %v", err)
	}
	return 0
}

func resolveMode(subcommand string, admin bool) container.Mode {
	switch subcommand {
	case "shell":
		if admin {
			return container.ModeAdminShell
		}
		return container.ModeShell
	case "run":
		return container.ModeCommand
	default:
		return container.ModeAgent
	}
}

func runCleanup() {
	log.Info("removing asylum images and volumes...")

	var errs int
	if err := docker.RemoveImages("asylum:latest"); err != nil {
		log.Error("remove asylum:latest: %v", err)
		errs++
	}

	if imgs, err := docker.ListImages("asylum:proj-*"); err != nil {
		log.Error("list project images: %v", err)
		errs++
	} else if len(imgs) > 0 {
		if err := docker.RemoveImages(imgs...); err != nil {
			log.Error("remove project images: %v", err)
			errs++
		}
	}

	if vols, err := docker.ListVolumes("asylum-"); err != nil {
		log.Error("list volumes: %v", err)
		errs++
	} else if len(vols) > 0 {
		if err := docker.RemoveVolumes(vols...); err != nil {
			log.Error("remove volumes: %v", err)
			errs++
		}
	}

	switch {
	case errs == 0:
		log.Success("images and volumes removed")
	case errs < 3:
		log.Warn("partially cleaned up (see errors above)")
	default:
		log.Warn("cleanup failed (see errors above)")
	}

	fmt.Print("Remove cached data (~/.asylum/cache/ and ~/.asylum/projects/)? (y/N) ")
	var answer string
	fmt.Scanln(&answer)

	if strings.ToLower(strings.TrimSpace(answer)) == "y" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Error("home dir: %v", err)
			return
		}
		if err := os.RemoveAll(filepath.Join(home, ".asylum", "cache")); err != nil {
			log.Error("remove cache: %v", err)
		}
		if err := os.RemoveAll(filepath.Join(home, ".asylum", "projects")); err != nil {
			log.Error("remove projects: %v", err)
		}
		log.Success("cached data removed")
	}

	log.Info("agent config (~/.asylum/agents/) preserved — delete manually if needed")
}

// resolveProfileTiers splits allProfiles into global (for base image) and
// project-only (for project image). Global profiles come from ~/.asylum/config.yaml;
// project-only profiles are those in allProfiles but not in the global set.
func resolveProfileTiers(projectDir string, allProfiles []*profile.Profile) (global, projectOnly []*profile.Profile) {
	home, err := os.UserHomeDir()
	if err != nil {
		return allProfiles, nil
	}

	globalCfg, err := config.LoadFile(filepath.Join(home, ".asylum", "config.yaml"))
	if err != nil {
		return allProfiles, nil
	}

	globalResolved, err := profile.Resolve(globalCfg.Profiles)
	if err != nil {
		return allProfiles, nil
	}

	globalSet := map[string]bool{}
	for _, p := range globalResolved {
		globalSet[p.Name] = true
	}

	for _, p := range allProfiles {
		if globalSet[p.Name] {
			global = append(global, p)
		} else {
			projectOnly = append(projectOnly, p)
		}
	}
	return global, projectOnly
}

func printUsage() {
	fmt.Printf(`asylum %s — Docker sandbox for AI coding agents

Usage:
  asylum [flags]                Start default agent
  asylum [flags] -- [args]      Start agent with extra args
  asylum [flags] shell          Interactive zsh shell
  asylum [flags] shell --admin  Admin shell with sudo notice
  asylum [flags] run <cmd>      Run command in container
  asylum ssh-init               Initialize SSH directory
  asylum self-update [--dev]    Update to latest version
  asylum self-update --safe     Emergency update (always dev, no checks)

Flags:
  -a, --agent <name>   Agent: claude, gemini, codex (default: claude)
  -p <port>            Port forwarding (repeatable)
  -v <volume>          Additional volume mount (repeatable)
  -e KEY=VALUE         Environment variable (repeatable, last wins)
  --java <version>     Java version in container
  --profiles <list>    Comma-separated profiles (default: all)
  -n, --new            Start new session (skip resume)
  --rebuild            Force rebuild Docker image
  --skip-onboarding    Skip project onboarding tasks
  --cleanup            Remove Asylum images and cached data
  --version            Show version
  -h, --help           Show this help
`, version)
}
