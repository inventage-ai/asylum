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
	"github.com/inventage-ai/asylum/internal/kit"
	"github.com/inventage-ai/asylum/internal/ports"
	"github.com/inventage-ai/asylum/internal/selfupdate"
	"github.com/inventage-ai/asylum/internal/ssh"
	"github.com/inventage-ai/asylum/internal/term"
	"github.com/inventage-ai/asylum/internal/tui"
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

	// Flag aliases for subcommands
	if flags.Version {
		subcommand = "version"
	}
	if flags.Cleanup {
		subcommand = "cleanup"
	}

	switch subcommand {
	case "version":
		fmt.Print(versionString(flags.Short))
		return
	case "cleanup":
		runCleanup(flags.All)
		return
	case "ssh-init":
		if err := ssh.Init(); err != nil {
			die("%v", err)
		}
		return
	}

	projectDir, err := filepath.Abs(".")
	if err != nil {
		die("resolve project dir: %v", err)
	}

	kitSnippets := kit.AssembleConfigSnippets()

	// Sync new kits to config (detect newly registered kits, prompt, update config)
	if subcommand != "self-update" {
		home, err := os.UserHomeDir()
		if err == nil {
			asylumDir := filepath.Join(home, ".asylum")
			promptFn := func(kits []*kit.Kit) []string {
				options := make([]tui.Option, len(kits))
				var defaultSel []int
				for i, k := range kits {
					options[i] = tui.Option{Label: k.Name, Description: k.Description}
					if k.Tier == kit.TierDefault {
						defaultSel = append(defaultSel, i)
					}
				}
				selected, err := tui.MultiSelect("New kits available — select which to activate", options, defaultSel)
				if err != nil {
					return nil
				}
				names := make([]string, len(selected))
				for i, idx := range selected {
					names[i] = kits[idx].Name
				}
				return names
			}
			if _, err := config.SyncNewKits(asylumDir, term.IsTerminal(), promptFn); err != nil {
				log.Error("kit sync: %v", err)
			}
		}
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
		cfg, err := config.Load(projectDir, config.CLIFlags{}, kitSnippets)
		if err != nil {
			die("load config: %v", err)
		}
		channel := selfupdate.ResolveChannel(flags.Dev, cfg.ReleaseChannel)
		if err := selfupdate.Run(version, commit, channel, flags.TargetVersion, execPath); err != nil {
			die("%v", err)
		}
		return
	}

	// Write default config on first run
	if home, err := os.UserHomeDir(); err == nil {
		cfgPath := filepath.Join(home, ".asylum", "config.yaml")
		if err := os.MkdirAll(filepath.Dir(cfgPath), 0755); err != nil {
			log.Error("create config directory: %v", err)
		} else if err := config.WriteDefaults(cfgPath, kitSnippets); err != nil {
			log.Error("write default config: %v", err)
		}
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
		Kits: flags.Kits,
		Agents:   flags.Agents,
		Ports:    flags.Ports,
		Volumes:  flags.Volumes,
		Env:      flags.Env,
		Java:     flags.Java,
	}, kitSnippets)
	if err != nil {
		die("load config: %v", err)
	}

	// Resolve all active kits from config
	allKits, err := kit.Resolve(cfg.KitNames(), cfg.DisabledKits())
	if err != nil {
		die("%v", err)
	}

	// Resolve global-tier kits (from ~/.asylum/config.yaml only) for base image.
	globalKits, projectKits := resolveKitTiers(projectDir, allKits)

	// Resolve agent installs (nil defaults to claude-only)
	agentMap := agentConfigToMap(cfg.Agents)
	kitNames := make([]string, len(allKits))
	for i, k := range allKits {
		kitNames[i] = k.Name
	}
	agentInstalls, err := agent.ResolveInstalls(agentMap, kitNames)
	if err != nil {
		die("%v", err)
	}

	cacheDirs := kit.AggregateCacheDirs(allKits)

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

	if err := container.MigrateProjectDir(projectDir); err != nil {
		log.Warn("project migration: %v", err)
	}

	cname := container.ContainerName(projectDir)
	newSession := flags.New
	freshContainer := false

	// If --rebuild requested but container is already running, ask to kill it
	if flags.Rebuild && docker.IsRunning(cname) {
		confirmed, err := tui.Confirm("Container is running. Kill it and rebuild?", false)
		if err != nil {
			die("aborted")
		}
		if confirmed {
			docker.RemoveContainer(cname)
		}
	}

	// If no container running, build images and start one detached
	if !docker.IsRunning(cname) {
		// Onboarding wizard: prompt for any unconfigured options (isolation, credentials)
		runOnboarding(&cfg, a, allKits, home)

		// Ensure agent config exists — behavior depends on isolation level
		switch cfg.AgentIsolation(a.Name()) {
		case "shared":
			// Host dir used directly — no seeding needed
		case "project":
			// Seed per-project dir from host config
			projConfigDir := filepath.Join(home, ".asylum", "projects", cname, a.Name()+"-config")
			seeded, err := container.EnsureAgentConfigAt(home, a, projConfigDir)
			if err != nil {
				die("%v", err)
			}
			if seeded {
				newSession = true
			}
		default: // "isolated" or empty
			seeded, err := container.EnsureAgentConfig(home, a)
			if err != nil {
				die("%v", err)
			}
			if seeded {
				newSession = true
			}
		}

		baseRebuilt, err := image.EnsureBase(globalKits, agentInstalls, version, flags.Rebuild)
		if err != nil {
			die("%v", err)
		}

		imageTag, err := image.EnsureProject(projectKits, collectPackages(cfg), cfg.JavaVersion(), version, baseRebuilt, flags.Rebuild)
		if err != nil {
			die("%v", err)
		}

		var allocatedPorts []int
		if cfg.KitActive("ports") {
			pr, err := ports.Allocate(projectDir, cname, cfg.PortCount())
			if err != nil {
				log.Warn("port allocation: %v", err)
			} else {
				allocatedPorts = pr.Ports()
			}
		}

		runArgs, err := container.RunArgs(container.RunOpts{
			Config:         cfg,
			Agent:          a,
			ImageTag:       imageTag,
			ProjectDir:     projectDir,
			CacheDirs:      cacheDirs,
			Kits:           allKits,
			Version:        version,
			AllocatedPorts: allocatedPorts,
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
		if !cfg.ShadowNodeModulesOff() {
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
		if c, ok := a.(interface{ WriteMarker(string, string) error }); ok {
			configDir, err := agent.ResolveConfigDir(a, cfg.AgentIsolation(a.Name()), cname)
			if err == nil {
				if err := c.WriteMarker(configDir, projectDir); err != nil {
					log.Error("write session marker: %v", err)
				}
			}
		}
	}

	// Run onboarding tasks (agent mode, first container start only)
	if containerMode == container.ModeAgent && freshContainer && !flags.SkipOnboarding {
		containerPath, _ := docker.ReadFile(cname, "/tmp/asylum-path")
		onboarding.Run(onboarding.Opts{
			ProjectDir:    projectDir,
			ContainerName: cname,
			ContainerPath: containerPath,
			Tasks:         kit.AggregateOnboardingTasks(allKits),
			Onboarding:    collectOnboarding(cfg),
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

	setTabTitle(cfg.TabTitle(), projectDir, agentName, containerMode)
	exitCode := runDocker(execArgs)

	// Cleanup: remove container if no other sessions are active
	if !docker.HasOtherSessions(cname) {
		docker.RemoveContainer(cname)
	}

	os.Exit(exitCode)
}

type cliFlags struct {
	Agent    string
	Kits *[]string
	Agents   *[]string
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
	All            bool
	Admin          bool
	Dev            bool
	SkipOnboarding bool
	Safe           bool
	TargetVersion  string
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
		case arg == "--kits":
			var val string
			if val, err = next(arg); err == nil {
				p := strings.Split(val, ",")
				flags.Kits = &p
			}
		case arg == "--agents":
			var val string
			if val, err = next(arg); err == nil {
				a := strings.Split(val, ",")
				flags.Agents = &a
			}
		case arg == "-w" || arg == "--worktree":
			extraArgs = append(extraArgs, "--worktree")
			i++
			if i < len(args) && !strings.HasPrefix(args[i], "-") {
				extraArgs = append(extraArgs, args[i])
				i++
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
			if i < len(args) && args[i] == "--all" {
				flags.All = true
				i++
			}
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
		case arg == "version":
			subcommand = "version"
			i++
			for i < len(args) {
				if args[i] == "--short" {
					flags.Short = true
				} else {
					return cliFlags{}, "", nil, fmt.Errorf("unknown flag %q for version", args[i])
				}
				i++
			}
		case arg == "cleanup":
			subcommand = "cleanup"
			i++
			for i < len(args) {
				if args[i] == "--all" {
					flags.All = true
					i++
				} else {
					return cliFlags{}, "", nil, fmt.Errorf("unknown flag %q for cleanup (only --all is supported)", args[i])
				}
			}
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
		case arg == "self-update" || arg == "selfupdate":
			subcommand = "self-update"
			i++
			for i < len(args) {
				switch {
				case args[i] == "--dev":
					flags.Dev = true
				case args[i] == "--safe":
					flags.Safe = true
				case !strings.HasPrefix(args[i], "-") && flags.TargetVersion == "":
					flags.TargetVersion = args[i]
				case !strings.HasPrefix(args[i], "-"):
					return cliFlags{}, "", nil, fmt.Errorf("unexpected argument %q for self-update", args[i])
				default:
					return cliFlags{}, "", nil, fmt.Errorf("unknown flag %q for self-update", args[i])
				}
				i++
			}
			if flags.Dev && flags.TargetVersion != "" {
				return cliFlags{}, "", nil, fmt.Errorf("--dev and version argument cannot be combined")
			}
			if flags.Safe && flags.TargetVersion != "" {
				return cliFlags{}, "", nil, fmt.Errorf("--safe and version argument cannot be combined")
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
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		for sig := range sigCh {
			if cmd.Process != nil {
				cmd.Process.Signal(sig)
			}
		}
	}()

	err := cmd.Run()
	signal.Stop(sigCh)
	close(sigCh)

	if err != nil {
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

func runCleanup(all bool) {
	if all {
		runCleanupAll()
		return
	}
	runCleanupProject()
}

func runCleanupProject() {
	projectDir, err := filepath.Abs(".")
	if err != nil {
		log.Error("resolve project dir: %v", err)
		log.Info("use 'asylum cleanup --all' to remove all asylum resources")
		return
	}

	// Migrate old-format project dir before cleanup so we find the right resources
	container.MigrateProjectDir(projectDir)

	cname := container.ContainerName(projectDir)
	oldName := container.OldContainerName(projectDir)
	log.Info("cleaning up project %s (container %s)...", filepath.Base(projectDir), cname)

	var errs int

	// Remove container (ignore error if not exists)
	if docker.IsRunning(cname) {
		if err := docker.RemoveContainer(cname); err != nil {
			log.Error("remove container: %v", err)
			errs++
		}
	}

	// Remove volumes prefixed with container name (new and old format)
	for _, prefix := range []string{cname + "-", oldName + "-"} {
		if vols, err := docker.ListVolumes(prefix); err != nil {
			log.Error("list volumes: %v", err)
			errs++
		} else if len(vols) > 0 {
			if err := docker.RemoveVolumes(vols...); err != nil {
				log.Error("remove volumes: %v", err)
				errs++
			}
		}
	}

	// Remove project data dir
	home, err := os.UserHomeDir()
	if err != nil {
		log.Error("home dir: %v", err)
		errs++
	} else {
		projDir := filepath.Join(home, ".asylum", "projects", cname)
		if _, err := os.Stat(projDir); err == nil {
			ports.ReleaseContainer(cname)
			if err := os.RemoveAll(projDir); err != nil {
				log.Error("remove project data: %v", err)
				errs++
			}
		}
	}

	if errs == 0 {
		log.Success("project cleaned up")
	} else {
		log.Warn("partially cleaned up (see errors above)")
	}
}

func runCleanupAll() {
	if !term.IsTerminal() {
		log.Error("cleanup --all requires a terminal for confirmation")
		return
	}

	// Enumerate resources
	var images []string
	if imgs, err := docker.ListImages("asylum:proj-*"); err == nil {
		images = append(images, imgs...)
	}
	images = append(images, "asylum:latest")

	volumes, _ := docker.ListVolumes("asylum-")

	fmt.Println("The following resources will be removed:")
	fmt.Println()
	fmt.Println("  Images:")
	for _, img := range images {
		fmt.Printf("    %s\n", img)
	}
	if len(volumes) > 0 {
		fmt.Println()
		fmt.Println("  Volumes:")
		for _, vol := range volumes {
			fmt.Printf("    %s\n", vol)
		}
	}
	fmt.Println()

	fmt.Print("Proceed? (y/N) ")
	var answer string
	fmt.Scanln(&answer)
	if strings.ToLower(strings.TrimSpace(answer)) != "y" {
		log.Info("aborted")
		return
	}

	var errs int
	if err := docker.RemoveImages(images...); err != nil {
		log.Error("remove images: %v", err)
		errs++
	}

	if len(volumes) > 0 {
		if err := docker.RemoveVolumes(volumes...); err != nil {
			log.Error("remove volumes: %v", err)
			errs++
		}
	}

	switch {
	case errs == 0:
		log.Success("images and volumes removed")
	case errs < 2:
		log.Warn("partially cleaned up (see errors above)")
	default:
		log.Warn("cleanup failed (see errors above)")
	}

	fmt.Print("Remove cached data (~/.asylum/cache/ and ~/.asylum/projects/)? (y/N) ")
	answer = ""
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
		if err := removeProjectsDir(filepath.Join(home, ".asylum", "projects")); err != nil {
			log.Error("remove projects: %v", err)
		}
		log.Success("cached data removed")
	}

	log.Info("agent config (~/.asylum/agents/) preserved — delete manually if needed")
}

// removeProjectsDir removes project data but skips directories with active
// session counters to avoid killing running containers.
func removeProjectsDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var skipped int
	for _, e := range entries {
		if !e.IsDir() {
			os.Remove(filepath.Join(dir, e.Name()))
			continue
		}
		counter := filepath.Join(dir, e.Name(), "sessions")
		if data, err := os.ReadFile(counter); err == nil {
			if n := strings.TrimSpace(string(data)); n != "" && n != "0" {
				log.Warn("skipping %s (active session)", e.Name())
				skipped++
				continue
			}
		}
		ports.ReleaseContainer(e.Name())
		os.RemoveAll(filepath.Join(dir, e.Name()))
	}
	if skipped == 0 {
		return os.Remove(dir)
	}
	return nil
}

// resolveKitTiers splits allKits into global (for base image) and
// project-only (for project image). Global kits come from ~/.asylum/config.yaml;
// project-only kits are those in allKits but not in the global set.
func resolveKitTiers(projectDir string, allKits []*kit.Kit) (global, projectOnly []*kit.Kit) {
	home, err := os.UserHomeDir()
	if err != nil {
		return allKits, nil
	}

	globalCfg, err := config.LoadFile(filepath.Join(home, ".asylum", "config.yaml"))
	if err != nil {
		return allKits, nil
	}

	globalResolved, err := kit.Resolve(globalCfg.KitNames(), globalCfg.DisabledKits())
	if err != nil {
		return allKits, nil
	}

	globalSet := map[string]bool{}
	for _, k := range globalResolved {
		globalSet[k.Name] = true
	}

	for _, k := range allKits {
		if globalSet[k.Name] {
			global = append(global, k)
		} else {
			projectOnly = append(projectOnly, k)
		}
	}
	return global, projectOnly
}

// agentConfigToMap converts the agents config map to a simple name→bool map for ResolveInstalls.
func agentConfigToMap(agents map[string]*config.AgentConfig) map[string]bool {
	if agents == nil {
		return nil
	}
	m := make(map[string]bool, len(agents))
	for name := range agents {
		m[name] = true
	}
	return m
}

// collectPackages aggregates packages from kit configs into the old map format for EnsureProject.
func collectPackages(cfg config.Config) map[string][]string {
	pkgs := map[string][]string{}
	// kit name → output key
	for kit, key := range map[string]string{"apt": "apt", "node": "npm", "python": "pip", "cx": "cx-lang"} {
		if p := cfg.KitPackages(kit); len(p) > 0 {
			pkgs[key] = p
		}
	}
	if kc := cfg.KitOption("shell"); kc != nil && len(kc.Build) > 0 {
		pkgs["run"] = kc.Build
	}
	return pkgs
}

// collectOnboarding builds the onboarding enabled map from kit configs.
func collectOnboarding(cfg config.Config) map[string]bool {
	m := map[string]bool{}
	// Map kit onboarding settings to task names
	if cfg.OnboardingEnabled("node") {
		m["npm"] = true
	}
	return m
}

// runOnboarding collects all pending onboarding steps and presents them
// as a single wizard flow. Each step fires when its config is not yet set.
func runOnboarding(cfg *config.Config, a agent.Agent, allKits []*kit.Kit, home string) {
	type applyFunc func(result tui.StepResult)
	var steps []tui.WizardStep
	var appliers []applyFunc

	// Step: config isolation (if not set for Claude)
	if a.Name() == "claude" && cfg.AgentIsolation(a.Name()) == "" {
		steps = append(steps, tui.WizardStep{
			Title:       "Config Isolation",
			Description: "How should Claude's config (~/.claude) be managed inside the sandbox?",
			Kind:        tui.StepSelect,
			Options: []tui.Option{
				{Label: "Shared with host", Description: "Use your host ~/.claude directly. Changes sync both ways."},
				{Label: "Isolated (recommended)", Description: "Separate from host, shared across projects. This is the current default."},
				{Label: "Project-isolated", Description: "Separate config per project. No state shared between projects."},
			},
			DefaultIdx: 1,
		})
		appliers = append(appliers, func(result tui.StepResult) {
			levels := []string{"shared", "isolated", "project"}
			level := levels[result.SelectIdx]
			cfgPath := filepath.Join(home, ".asylum", "config.yaml")
			if err := config.SetAgentIsolation(cfgPath, a.Name(), level); err != nil {
				log.Error("save isolation config: %v", err)
			}
			if cfg.Agents == nil {
				cfg.Agents = map[string]*config.AgentConfig{}
			}
			if cfg.Agents[a.Name()] == nil {
				cfg.Agents[a.Name()] = &config.AgentConfig{}
			}
			cfg.Agents[a.Name()].Config = level
		})
	}

	// Step: kit credentials — show all credential-capable kits, pre-select configured ones.
	// Only show the step if at least one kit is unconfigured.
	var credKits []*kit.Kit
	hasUnconfigured := false
	for _, k := range allKits {
		if k.CredentialFunc == nil {
			continue
		}
		credKits = append(credKits, k)
		parent, _, _ := strings.Cut(k.Name, "/")
		if cfg.KitCredentialMode(parent) == "" {
			hasUnconfigured = true
		}
	}
	if hasUnconfigured {
		options := make([]tui.Option, len(credKits))
		var preSelected []int
		for i, k := range credKits {
			label := k.CredentialLabel
			if label == "" {
				label = k.Name
			}
			desc := ""
			if k.Name == "java/maven" {
				desc = "Exposes matching server entries from ~/.m2/settings.xml"
			}
			if k.Name == "github" {
				desc = "Extracts gh auth token for CLI authentication"
			}
			options[i] = tui.Option{Label: label, Description: desc}
			parent, _, _ := strings.Cut(k.Name, "/")
			if cfg.KitCredentialMode(parent) != "" {
				preSelected = append(preSelected, i)
			}
		}
		steps = append(steps, tui.WizardStep{
			Title:       "Credentials",
			Description: "Allow the sandbox to access host credentials for private registries and repositories (read-only, scoped to what the project needs).",
			Kind:       tui.StepMultiSelect,
			Options:    options,
			DefaultSel: preSelected,
		})
		appliers = append(appliers, func(result tui.StepResult) {
			cfgPath := filepath.Join(home, ".asylum", "config.yaml")
			selected := map[int]bool{}
			for _, idx := range result.MultiIdx {
				selected[idx] = true
			}
			for i, k := range credKits {
				parent, _, _ := strings.Cut(k.Name, "/")
				if selected[i] {
					if err := config.SetKitCredentials(cfgPath, parent, "auto"); err != nil {
						log.Error("save credential config: %v", err)
					}
					if cfg.Kits == nil {
						cfg.Kits = map[string]*config.KitConfig{}
					}
					if cfg.Kits[parent] == nil {
						cfg.Kits[parent] = &config.KitConfig{}
					}
					cfg.Kits[parent].Credentials = &config.Credentials{Auto: true}
				} else if cfg.KitCredentialMode(parent) == "" {
					if err := config.SetKitCredentials(cfgPath, parent, "false"); err != nil {
						log.Error("save credential config: %v", err)
					}
					if cfg.Kits == nil {
						cfg.Kits = map[string]*config.KitConfig{}
					}
					if cfg.Kits[parent] == nil {
						cfg.Kits[parent] = &config.KitConfig{}
					}
					cfg.Kits[parent].Credentials = &config.Credentials{}
				}
			}
		})
	}

	if len(steps) == 0 {
		return
	}

	results, err := tui.Wizard(steps)
	if err != nil {
		die("aborted")
	}

	for i, r := range results {
		if r.Completed {
			appliers[i](r)
		}
	}
}

func printUsage() {
	fmt.Printf(`asylum %s — Docker sandbox for AI coding agents

Usage:
  asylum [flags]                Start default agent
  asylum [flags] -- [args]      Start agent with extra args
  asylum [flags] shell          Interactive zsh shell
  asylum [flags] shell --admin  Admin shell with sudo notice
  asylum [flags] run <cmd>      Run command in container
  asylum cleanup                Remove current project's container, volumes, and data
  asylum cleanup --all          Remove all Asylum images, volumes, and cached data
  asylum version [--short]      Show version
  asylum ssh-init               Initialize SSH directory
  asylum self-update [version]  Update to latest (or specific) version
  asylum self-update --dev      Update to latest dev build
  asylum self-update --safe     Emergency update (always dev, no checks)

Flags:
  -a, --agent <name>   Agent: claude, gemini, codex (default: claude)
  -p <port>            Port forwarding (repeatable)
  -v <volume>          Additional volume mount (repeatable)
  -e KEY=VALUE         Environment variable (repeatable, last wins)
  --java <version>     Java version in container
  --kits <list>    Comma-separated kits (default: all)
  --agents <list>      Comma-separated agents (default: claude)
  -n, --new            Start new session (skip resume)
  --rebuild            Force rebuild Docker image
  --skip-onboarding    Skip project onboarding tasks
  --cleanup            Alias for cleanup command
  --version            Alias for version command
  -h, --help           Show this help
`, version)
}
