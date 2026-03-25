package agent

import "strings"

func init() {
	agents["echo"] = Echo{}
}

// Echo is a minimal testing agent that runs the shell echo command.
type Echo struct{}

func (Echo) Name() string               { return "echo" }
func (Echo) Binary() string             { return "echo" }
func (Echo) NativeConfigDir() string    { return "" }
func (Echo) ContainerConfigDir() string { return "/tmp/asylum-echo" }
func (Echo) AsylumConfigDir() string    { return "~/.asylum/agents/echo" }
func (Echo) EnvVars() map[string]string { return nil }
func (Echo) HasSession(_ string) bool   { return false }

func (Echo) Command(_ bool, extraArgs []string) []string {
	if len(extraArgs) == 0 {
		return []string{"echo"}
	}
	return []string{"echo", strings.Join(extraArgs, " ")}
}
