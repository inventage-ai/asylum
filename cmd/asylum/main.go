package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/inventage-ai/asylum/internal/agent"
	"github.com/inventage-ai/asylum/internal/config"
	"github.com/inventage-ai/asylum/internal/container"
	"github.com/inventage-ai/asylum/internal/docker"
	"github.com/inventage-ai/asylum/internal/image"
	"github.com/inventage-ai/asylum/internal/log"
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
		if commit != "" {
			fmt.Printf("asylum %s (%s)\n", version, commit)
		} else {
			fmt.Printf("asylum %s\n", version)
		}
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

	if subcommand == "self-update" {
		projectDir, err := filepath.Abs(".")
		if err != nil {
			die("resolve project dir: %v", err)
		}
		cfg, err := config.Load(projectDir, config.CLIFlags{})
		if err != nil {
			die("load config: %v", err)
		}
		channel := selfupdate.ResolveChannel(flags.Dev, cfg.ReleaseChannel)
		execPath, err := os.Executable()
		if err != nil {
			die("resolve executable: %v", err)
		}
		if err := selfupdate.Run(version, channel, execPath); err != nil {
			die("%v", err)
		}
		return
	}

	containerMode := resolveMode(subcommand, flags.Admin)

	projectDir, err := filepath.Abs(".")
	if err != nil {
		die("resolve project dir: %v", err)
	}

	cfg, err := config.Load(projectDir, config.CLIFlags{
		Agent:   flags.Agent,
		Ports:   flags.Ports,
		Volumes: flags.Volumes,
		Java:    flags.Java,
	})
	if err != nil {
		die("load config: %v", err)
	}

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

	// For shell/run modes, exec into a running container instead of starting a new one
	cname := container.ContainerName(projectDir)
	if (containerMode == container.ModeShell || containerMode == container.ModeAdminShell || containerMode == container.ModeCommand) && docker.IsRunning(cname) {
		args := container.ExecArgs(cname, containerMode, extraArgs)
		dockerBin, err := exec.LookPath("docker")
		if err != nil {
			die("docker not found in PATH")
		}
		fullArgs := append([]string{"docker"}, args...)
		if err := syscall.Exec(dockerBin, fullArgs, os.Environ()); err != nil {
			die("exec docker: %v", err)
		}
	}

	baseRebuilt, err := image.EnsureBase(version, flags.Rebuild)
	if err != nil {
		die("%v", err)
	}

	imageTag, err := image.EnsureProject(cfg.Packages, cfg.Versions["java"], version, baseRebuilt, flags.Rebuild)
	if err != nil {
		die("%v", err)
	}

	args, err := container.RunArgs(container.RunOpts{
		Config:     cfg,
		Agent:      a,
		ImageTag:   imageTag,
		ProjectDir: projectDir,
		Mode:       containerMode,
		NewSession: flags.New,
		ExtraArgs:  extraArgs,
	})
	if err != nil {
		die("%v", err)
	}

	dockerBin, err := exec.LookPath("docker")
	if err != nil {
		die("docker not found in PATH")
	}

	if containerMode == container.ModeAgent {
		if c, ok := a.(interface{ WriteMarker(string) error }); ok {
			if err := c.WriteMarker(projectDir); err != nil {
				log.Error("write session marker: %v", err)
			}
		}
	}

	fullArgs := append([]string{"docker"}, args...)
	if err := syscall.Exec(dockerBin, fullArgs, os.Environ()); err != nil {
		die("exec docker: %v", err)
	}
}

type cliFlags struct {
	Agent   string
	Ports   []string
	Volumes []string
	Java    string
	New     bool
	Rebuild bool
	Cleanup bool
	Help    bool
	Version bool
	Admin   bool
	Dev     bool
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
		case arg == "--java":
			flags.Java, err = next(arg)
		case arg == "-n" || arg == "--new":
			flags.New = true
			i++
		case arg == "--rebuild":
			flags.Rebuild = true
			i++
		case arg == "--cleanup":
			flags.Cleanup = true
			i++
		case arg == "--version":
			flags.Version = true
			i++
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
				if args[i] == "--dev" {
					flags.Dev = true
					i++
				} else {
					return cliFlags{}, "", nil, fmt.Errorf("unknown flag %q for self-update (only --dev is supported)", args[i])
				}
			}
		case arg == "run":
			subcommand = "run"
			rest := args[i+1:]
			if len(rest) > 0 && rest[0] == "--" {
				rest = rest[1:]
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
	log.Info("removing asylum images...")

	if err := docker.RemoveImages("asylum:latest"); err != nil {
		log.Error("remove asylum:latest: %v", err)
	}

	if imgs, err := docker.ListImages("asylum:proj-*"); err != nil {
		log.Error("list project images: %v", err)
	} else if len(imgs) > 0 {
		if err := docker.RemoveImages(imgs...); err != nil {
			log.Error("remove project images: %v", err)
		}
	}

	log.Success("images removed")

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

Flags:
  -a, --agent <name>   Agent: claude, gemini, codex (default: claude)
  -p <port>            Port forwarding (repeatable)
  -v <volume>          Additional volume mount (repeatable)
  --java <version>     Java version in container
  -n, --new            Start new session (skip resume)
  --rebuild            Force rebuild Docker image
  --cleanup            Remove Asylum images and cached data
  --version            Show version
  -h, --help           Show this help
`, version)
}
