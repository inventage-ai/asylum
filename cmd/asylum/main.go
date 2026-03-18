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
	"github.com/inventage-ai/asylum/internal/ssh"
)

var version = "dev"

func die(format string, args ...any) {
	log.Error(format, args...)
	os.Exit(1)
}

func main() {
	flags, positional, passthrough, err := parseArgs(os.Args[1:])
	if err != nil {
		die("%v", err)
	}

	if flags.Help {
		printUsage()
		return
	}

	if flags.Version {
		fmt.Printf("asylum %s\n", version)
		return
	}

	if flags.Cleanup {
		runCleanup()
		return
	}

	containerMode, isSSHInit, extraArgs, err := resolveMode(positional, passthrough)
	if err != nil {
		die("%v", err)
	}

	if isSSHInit {
		if err := ssh.Init(); err != nil {
			die("%v", err)
		}
		return
	}

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

	baseRebuilt, err := image.EnsureBase(version, flags.Rebuild)
	if err != nil {
		die("%v", err)
	}

	imageTag, err := image.EnsureProject(cfg.Packages, version, baseRebuilt, flags.Rebuild)
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
}

func parseArgs(args []string) (cliFlags, []string, []string, error) {
	var flags cliFlags
	var positional []string
	var passthrough []string

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
			passthrough = append(passthrough, args[i+1:]...)
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
		case strings.HasPrefix(arg, "-"):
			passthrough = append(passthrough, args[i:]...)
			i = len(args)
		default:
			positional = append(positional, arg)
			if arg != "shell" && arg != "ssh-init" {
				passthrough = append(passthrough, args[i+1:]...)
				i = len(args)
			} else {
				i++
			}
		}
		if err != nil {
			return cliFlags{}, nil, nil, err
		}
	}

	return flags, positional, passthrough, nil
}

func resolveMode(positional, passthrough []string) (container.Mode, bool, []string, error) {
	if len(positional) == 0 {
		return container.ModeAgent, false, passthrough, nil
	}

	switch positional[0] {
	case "shell":
		if len(positional) > 1 {
			return 0, false, nil, fmt.Errorf("unexpected argument %q after shell", positional[1])
		}
		for _, arg := range passthrough {
			if arg == "--admin" {
				return container.ModeAdminShell, false, nil, nil
			}
		}
		return container.ModeShell, false, nil, nil
	case "ssh-init":
		if len(positional) > 1 {
			return 0, false, nil, fmt.Errorf("unexpected argument %q after ssh-init", positional[1])
		}
		return 0, true, nil, nil
	default:
		return container.ModeCommand, false, append(positional, passthrough...), nil
	}
}

func runCleanup() {
	log.Info("removing asylum images...")

	if err := docker.RemoveImages("asylum:latest"); err != nil {
		log.Error("remove asylum:latest: %v", err)
	}

	out, err := exec.Command("docker", "images", "--format", "{{.Repository}}:{{.Tag}}", "--filter", "reference=asylum:proj-*").Output()
	if err == nil {
		for _, img := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			if img != "" {
				if err := docker.RemoveImages(img); err != nil {
					log.Error("remove %s: %v", img, err)
				}
			}
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
		os.RemoveAll(filepath.Join(home, ".asylum", "cache"))
		os.RemoveAll(filepath.Join(home, ".asylum", "projects"))
		log.Success("cached data removed")
	}

	log.Info("agent config (~/.asylum/agents/) preserved — delete manually if needed")
}

func printUsage() {
	fmt.Printf(`asylum %s — Docker sandbox for AI coding agents

Usage:
  asylum                     Start default agent in YOLO mode
  asylum -a gemini           Start Gemini CLI in YOLO mode
  asylum shell               Interactive zsh shell
  asylum shell --admin       Admin shell with sudo notice
  asylum ssh-init            Initialize SSH directory
  asylum <cmd> [args...]     Run arbitrary command in container

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
