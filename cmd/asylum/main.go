package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/inventage-ai/asylum/internal/agent"
	"github.com/inventage-ai/asylum/internal/config"
	"github.com/inventage-ai/asylum/internal/container"
	"github.com/inventage-ai/asylum/internal/docker"
	"github.com/inventage-ai/asylum/internal/firstrun"
	"github.com/inventage-ai/asylum/internal/image"
	"github.com/inventage-ai/asylum/internal/kit"
	"github.com/inventage-ai/asylum/internal/log"
	"github.com/inventage-ai/asylum/internal/onboarding"
	"github.com/inventage-ai/asylum/internal/selfupdate"
	"github.com/inventage-ai/asylum/internal/term"
	"github.com/inventage-ai/asylum/internal/tui"
	"github.com/inventage-ai/asylum/internal/versions"
	"github.com/inventage-ai/asylum/internal/workspace"
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
	case "config":
		runConfig()
		return
	}

	projectDir, err := filepath.Abs(".")
	if err != nil {
		die("resolve project dir: %v", err)
	}

	if home, err := os.UserHomeDir(); err != nil {
		die("home dir: %v", err)
	} else if dir, redirected, err := workspace.Resolve(projectDir, home); err != nil {
		die("create workspace: %v", err)
	} else if redirected {
		log.Warn("Your home directory can't be sandboxed. Started a fresh workspace:")
		log.Warn("  %s", dir)
		projectDir = dir
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

	containerMode := resolveMode(subcommand, flags.Admin)

	home, err := os.UserHomeDir()
	if err != nil {
		die("home dir: %v", err)
	}

	// Existing-user signal for the resume-migration dialog. Probes
	// ~/.asylum/agents/, which is materialised lazily by EnsureAgentConfig —
	// the early default-config write at ~/.asylum/config.yaml does not touch
	// it, so this probe is unaffected by initialisation order.
	existingInstall := firstrun.IsExistingInstall(home)

	// First-run detection must happen before any config.yaml write.
	isFirstRun := firstrun.IsFirstRun(home)
	cfgPath := filepath.Join(home, ".asylum", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0755); err != nil {
		log.Error("create config directory: %v", err)
	}

	cliFlags := config.CLIFlags{
		Agent:   flags.Agent,
		Kits:    flags.Kits,
		Agents:  flags.Agents,
		Ports:   flags.Ports,
		Volumes: flags.Volumes,
		Env:     flags.Env,
		Java:    flags.Java,
	}

	// On first-run we defer config writing until the wizard supplies the
	// user's selections; on subsequent runs the default file is already in
	// place.
	if !isFirstRun {
		if err := config.WriteDefaults(cfgPath, kitSnippets); err != nil {
			log.Error("write default config: %v", err)
		}
	}

	cfg, err := config.Load(projectDir, cliFlags, kitSnippets)
	if err != nil {
		die("load config: %v", err)
	}

	// Pre-resolve kits for credentials gating; wizard inspects this.
	preWizardKits, err := kit.Resolve(cfg.KitNames(), cfg.DisabledKits())
	if err != nil {
		die("%v", err)
	}

	wizardOut, err := firstrun.Run(firstrun.WizardInput{
		Home:       home,
		IsFirstRun: isFirstRun,
		Cfg:        &cfg,
		AllKits:    preWizardKits,
		Reload: func() (*config.Config, []*kit.Kit, error) {
			c, err := config.Load(projectDir, cliFlags, kitSnippets)
			if err != nil {
				return nil, nil, err
			}
			k, err := kit.Resolve(c.KitNames(), c.DisabledKits())
			if err != nil {
				return nil, nil, err
			}
			return &c, k, nil
		},
	})
	if err != nil {
		die("first-run wizard: %v", err)
	}

	// When the wizard writes a new config (first-run agents/kits selections
	// change image inputs), re-resolve so EnsureBase/EnsureProject see the
	// updated layer. The wizard reloaded internally between phases, but the
	// caller-side `cfg`/`preWizardKits` references still need the update —
	// the wizard mutates `*in.Cfg` in place, but downstream re-resolutions
	// (allKits, agentInstalls, etc.) run from `cfg` here.
	if wizardOut.WroteConfig {
		cfg, err = config.Load(projectDir, cliFlags, kitSnippets)
		if err != nil {
			die("load config: %v", err)
		}
	}

	// Resolve all active kits from config
	allKits, err := kit.Resolve(cfg.KitNames(), cfg.DisabledKits())
	if err != nil {
		die("%v", err)
	}

	// Resolve global-tier kits (from ~/.asylum/config.yaml only) for base image.
	globalKits, projectKits := resolveKitTiers(projectDir, allKits)

	// Resolve agent installs (nil defaults to claude-only)
	agentName := cfg.Agent
	if agentName == "" {
		agentName = "claude"
	}
	agentMap := agentConfigToMap(cfg.Agents)
	// Ensure the active agent is in the install map so it gets baked into the image
	if agentMap == nil {
		agentMap = map[string]bool{agentName: true}
	} else if !agentMap[agentName] {
		agentMap[agentName] = true
	}
	kitNames := make([]string, len(allKits))
	for i, k := range allKits {
		kitNames[i] = k.Name
	}
	agentInstalls, err := agent.ResolveInstalls(agentMap, kitNames)
	if err != nil {
		die("%v", err)
	}

	cacheDirs := kit.AggregateCacheDirs(allKits)

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

	// Resolve the container for this run. The primary container (named for the
	// project) serves the configured agent set. If it is already running but
	// does not support the requested agent, spill to a separate container named
	// for the project plus baked agent set, leaving the primary untouched. Such
	// secondary containers are for ad-hoc/review use and forward no ports.
	cname := container.ContainerName(projectDir)
	secondary := false
	if docker.IsRunning(cname) && !containerSupportsAgent(cname, agentName) {
		names := make([]string, len(agentInstalls))
		for i, ai := range agentInstalls {
			names[i] = ai.Name
		}
		cname = container.SecondaryContainerName(projectDir, names)
		secondary = true
		log.Info("agent %q is not in the project container; using a separate container (no port forwarding)", agentName)
	}
	// If we freshly seed the agent config from the host this run, suppress
	// asylum-injected resume even when the user has opted into
	// `default-resume: true` — the host's session markers don't represent a
	// container session.
	suppressResumeFromSeed := false
	freshContainer := false

	// Always ensure images are up to date (cheap when nothing changed)
	asylumDir := filepath.Join(home, ".asylum")
	state, err := config.LoadState(asylumDir)
	if err != nil {
		log.Warn("load state: %v", err)
	}

	// One-time resume-migration dialog for users who installed asylum before
	// the default-new-session behaviour change. New users are pre-marked here
	// so they never see the dialog on a future asylum version. Mutates cfg
	// when the user opts into legacy behaviour.
	_, promptStateDirty := firstrun.MaybeShowResumeMigrationPrompt(
		asylumDir, &state, &cfg, containerMode == container.ModeAgent, existingInstall,
	)

	// Load version pinning — blocking if file doesn't exist, background update otherwise.
	versionsPath := filepath.Join(asylumDir, "versions.json")
	versionsVM, err := versions.Read(versionsPath)
	if err != nil {
		log.Warn("load versions: %v (proceeding without version pinning)", err)
		versionsVM = nil
	}
	if versionsVM == nil {
		log.Info("fetching latest agent versions (this may take a moment)...")
		versionsVM = versions.FetchAll()
		if len(versionsVM) > 0 {
			if err := versions.Write(versionsPath, versionsVM); err != nil {
				log.Warn("write versions: %v", err)
			}
		}
	} else if stale, err := versions.NeedsRefresh(versionsPath, versionsVM, time.Hour); err != nil {
		log.Warn("check versions staleness: %v", err)
	} else if stale {
		go func() {
			vm := versions.FetchAll()
			if len(vm) == 0 {
				return
			}
			if err := versions.Write(versionsPath, vm); err != nil {
				log.Warn("update versions: %v", err)
			}
		}()
	}

	containerRunning := docker.IsRunning(cname)
	imageTag, stateChanged := ensureImages(globalKits, projectKits, allKits, agentInstalls, cfg, version, versionsVM, flags.Rebuild, &state, containerRunning)
	if stateChanged || promptStateDirty {
		if err := config.SaveState(asylumDir, state); err != nil {
			log.Warn("save state: %v", err)
		}
	}

	cfgHash := config.ConfigHash(cfg)

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

	// Check for stale running container (image or config changed)
	if docker.IsRunning(cname) {
		checkStaleContainer(cname, imageTag, cfgHash)
	}

	// If no container running, build images and start one detached
	if !docker.IsRunning(cname) {
		// Ensure agent config exists — behavior depends on isolation level.
		// Unset/empty defaults to "shared".
		switch cfg.AgentIsolation(a.Name()) {
		case "isolated":
			seeded, err := container.EnsureAgentConfig(home, a)
			if err != nil {
				die("%v", err)
			}
			if seeded {
				suppressResumeFromSeed = true
			}
		case "project":
			projConfigDir := filepath.Join(home, ".asylum", "projects", cname, a.Name()+"-config")
			seeded, err := container.EnsureAgentConfigAt(home, a, projConfigDir)
			if err != nil {
				die("%v", err)
			}
			if seeded {
				suppressResumeFromSeed = true
			}
		default: // "shared" or empty — host dir used directly
		}

		runArgs, resolved, overrides, err := container.RunArgs(container.RunOpts{
			Config:        cfg,
			Agent:         a,
			AgentInstalls: agentInstalls,
			ImageTag:      imageTag,
			ProjectDir:    projectDir,
			ContainerName: cname,
			Secondary:     secondary,
			CacheDirs:     cacheDirs,
			Kits:          allKits,
			Version:       version,
			ConfigHash:    cfgHash,
			Debug:         flags.Debug,
		})
		if err != nil {
			die("%v", err)
		}

		if flags.Debug {
			fmt.Fprint(os.Stderr, container.FormatDebug(resolved, overrides))
		}

		// Remove any stale stopped container with the same name
		docker.RemoveContainer(cname)

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

		// Fix ownership of named volumes (Docker creates them as root)
		uid := fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid())
		if !cfg.ShadowNodeModulesOff() {
			for _, nm := range container.FindNodeModulesDirs(projectDir) {
				docker.Exec(cname, "root", "chown", uid, nm)
			}
		}
		for _, dst := range cacheDirs {
			docker.Exec(cname, "root", "chown", uid, config.ExpandTilde(dst, home))
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
		DefaultResume: cfg.ResumeByDefault() && !suppressResumeFromSeed,
		Config:        cfg,
		Kits:          allKits,
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
	Agent          string
	Kits           *[]string
	Agents         *[]string
	Ports          []string
	Volumes        []string
	Env            map[string]string
	Java           string
	Rebuild        bool
	Cleanup        bool
	Help           bool
	Version        bool
	Short          bool
	All            bool
	Admin          bool
	Dev            bool
	SkipOnboarding bool
	Safe           bool
	Debug          bool
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
			// Deprecated no-op: starting a new session is now the default.
			// Kept so existing scripts/aliases continue to parse.
			i++
		case arg == "--continue" || arg == "--resume":
			// Passthrough: the underlying agent owns resume semantics.
			extraArgs = append(extraArgs, arg)
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
		case arg == "--debug":
			flags.Debug = true
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
		case arg == "config":
			subcommand = "config"
			i++
			if i < len(args) {
				return cliFlags{}, "", nil, fmt.Errorf("unexpected argument %q after config", args[i])
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

// ensureImages runs EnsureBase and EnsureProject unconditionally to determine
// the expected image tag. When a container is already running and image checks
// fail (e.g. docker inspect errors), it falls through gracefully.
func ensureImages(globalKits, projectKits, allKits []*kit.Kit, agentInstalls []*agent.AgentInstall, cfg config.Config, version string, versions versions.VersionMap, noCache bool, state *config.State, containerRunning bool) (imageTag string, stateChanged bool) {
	baseRebuilt, newOrder, err := image.EnsureBase(globalKits, agentInstalls, cfg.KitSnippetConfig, version, versions, noCache, state.DockerSourceOrder)
	if err != nil {
		if containerRunning {
			log.Warn("image check: %v (using running container)", err)
			return "", false
		}
		die("%v", err)
	}
	if newOrder != nil {
		state.DockerSourceOrder = newOrder
		stateChanged = true
	}

	imageTag, err = image.EnsureProject(projectKits, allKits, collectPackages(cfg), cfg.KitSnippetConfig, version, baseRebuilt, noCache)
	if err != nil {
		if containerRunning {
			log.Warn("image check: %v (using running container)", err)
			return "", stateChanged
		}
		die("%v", err)
	}
	return imageTag, stateChanged
}

// containerSupportsAgent reports whether a running container can serve the
// named agent, based on its asylum.agents label. A container created before
// agent labels existed (no label) is treated as supporting only the default
// agent, so existing single-agent setups keep working untouched.
func containerSupportsAgent(cname, agentName string) bool {
	labels, err := docker.InspectLabels(cname)
	if err != nil {
		return false
	}
	label := labels["asylum.agents"]
	if label == "" {
		return agentName == "claude"
	}
	return slices.Contains(strings.Split(label, ","), agentName)
}

// checkStaleContainer compares the running container's image and config hash
// against expected values. If stale: kills silently when no active sessions,
// prompts when sessions are active. Config-only drift produces a warning.
func checkStaleContainer(cname, imageTag, cfgHash string) {
	if imageTag == "" {
		return // image check was skipped (error fallthrough)
	}

	containerImg, err := docker.ContainerImageID(cname)
	if err != nil {
		return // can't inspect, skip check
	}
	expectedImg, err := docker.ImageID(imageTag)
	if err != nil {
		return
	}

	if containerImg != expectedImg {
		// HasOtherSessions returns false on error (for cleanup safety).
		// Here we want the opposite default: if we can't check, assume
		// sessions exist and prompt rather than silently killing.
		hasSession, checked := docker.CheckOtherSessions(cname)
		if hasSession || !checked {
			confirmed, err := tui.Confirm("Image has changed. Restart container?", true)
			if err != nil || !confirmed {
				return // user declined or aborted, exec into stale container
			}
		} else {
			log.Info("config changed, restarting container...")
		}
		docker.RemoveContainer(cname)
		return
	}

	// Image matches — check config drift
	existing, err := docker.InspectLabel(cname, "asylum.config.hash")
	if err != nil || existing == "" {
		return // no label (legacy container), skip
	}
	if existing != cfgHash {
		log.Warn("config changed (volumes/env/ports) — restart with --rebuild to apply")
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
  asylum config                 Configure kits, credentials, and isolation
  asylum cleanup                Remove current project's container, volumes, and data
  asylum cleanup --all          Remove all Asylum images, volumes, and cached data
  asylum version [--short]      Show version
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
  --continue           Forwarded to agent (resume previous session)
  --resume             Forwarded to agent (resume previous session)
  -n, --new            (deprecated, no-op — new session is now the default)
  --rebuild            Force rebuild Docker image
  --skip-onboarding    Skip project onboarding tasks
  --cleanup            Alias for cleanup command
  --version            Alias for version command
  -h, --help           Show this help
`, version)
}
