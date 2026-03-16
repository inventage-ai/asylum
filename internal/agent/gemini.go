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
	// Scan for a .project_root file matching this project path
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		root, err := os.ReadFile(filepath.Join(tmpDir, e.Name(), ".project_root"))
		if err != nil {
			continue
		}
		if strings.TrimSpace(string(root)) != projectPath {
			continue
		}
		chats, err := os.ReadDir(filepath.Join(tmpDir, e.Name(), "chats"))
		if err != nil {
			return false
		}
		return len(chats) > 0
	}
	return false
}

func (Gemini) Command(resume bool, extraArgs []string) []string {
	parts := []string{"gemini", "--yolo"}
	if resume {
		parts = append(parts, "--resume")
	}
	parts = append(parts, extraArgs...)
	return wrapZsh(strings.Join(parts, " "))
}
