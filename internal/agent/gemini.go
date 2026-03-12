package agent

import (
	"os"
	"path/filepath"
	"strings"
)

type Gemini struct{}

func (Gemini) Name() string             { return "gemini" }
func (Gemini) Binary() string           { return "gemini" }
func (Gemini) NativeConfigDir() string  { return "~/.gemini" }
func (Gemini) ContainerConfigDir() string { return "/home/claude/.gemini" }
func (Gemini) AsylumConfigDir() string  { return "~/.asylum/agents/gemini" }

func (Gemini) EnvVars() map[string]string { return nil }

func (Gemini) HasSession(projectPath string) bool {
	configDir, err := expandHome("~/.asylum/agents/gemini")
	if err != nil {
		return false
	}
	tmpDir := filepath.Join(configDir, "tmp")
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return false
	}
	return len(entries) > 0
}

func (Gemini) Command(resume bool, extraArgs []string) []string {
	parts := []string{"gemini", "--yolo"}
	if resume {
		parts = append(parts, "--resume")
	}
	parts = append(parts, extraArgs...)
	return wrapZsh(strings.Join(parts, " "))
}
