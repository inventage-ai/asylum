package agent

import "fmt"

type Agent interface {
	Name() string
	Binary() string
	NativeConfigDir() string
	ContainerConfigDir() string
	AsylumConfigDir() string
	EnvVars() map[string]string
	HasSession(projectPath string) bool
	Command(resume bool, extraArgs []string) []string
}

var agents = map[string]Agent{
	"claude": Claude{},
	"gemini": Gemini{},
	"codex":  Codex{},
}

func Get(name string) (Agent, error) {
	a, ok := agents[name]
	if !ok {
		return nil, fmt.Errorf("unknown agent: %q (valid: claude, gemini, codex)", name)
	}
	return a, nil
}

func wrapZsh(cmd string) []string {
	return []string{"zsh", "-c", "source ~/.zshrc && exec " + cmd}
}
