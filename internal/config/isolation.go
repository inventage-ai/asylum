package config

import (
	"os"
	"strings"
)

// SetAgentIsolation writes the config isolation level for an agent to the
// given config file. Uses text-based editing to preserve blank lines and comments.
func SetAgentIsolation(path, agentName, level string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")

	agentsIdx := findSection(lines, "agents")
	if agentsIdx == -1 {
		lines = append(lines, "", "agents:", "  "+agentName+":", "    config: "+level)
		return writeLines(path, lines)
	}

	agentIdx := findKey(lines, agentsIdx, agentName)
	if agentIdx == -1 {
		lines = insertAfter(lines, agentsIdx, 2, agentName+":")
		agentIdx = agentsIdx + 1
		lines = insertAfter(lines, agentIdx, 4, "config: "+level)
		return writeLines(path, lines)
	}

	configIdx := findKey(lines, agentIdx, "config")
	if configIdx != -1 {
		indent := leadingSpaces(lines[configIdx])
		lines[configIdx] = strings.Repeat(" ", indent) + "config: " + level
	} else {
		parentIndent := leadingSpaces(lines[agentIdx])
		lines = insertAfter(lines, agentIdx, parentIndent+2, "config: "+level)
	}

	return writeLines(path, lines)
}

// SetKitCredentials writes `credentials: <value>` under the named kit in a
// config file. Uses text-based editing to preserve blank lines and comments.
func SetKitCredentials(path, kitName, value string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")

	kitsIdx := findSection(lines, "kits")
	if kitsIdx == -1 {
		lines = append(lines, "", "kits:", "  "+kitName+":", "    credentials: "+value)
		return writeLines(path, lines)
	}

	kitIdx := findKey(lines, kitsIdx, kitName)
	if kitIdx == -1 {
		lines = insertAfter(lines, kitsIdx, 2, kitName+":")
		kitIdx = kitsIdx + 1
		lines = insertAfter(lines, kitIdx, 4, "credentials: "+value)
		return writeLines(path, lines)
	}

	credIdx := findKey(lines, kitIdx, "credentials")
	if credIdx != -1 {
		indent := leadingSpaces(lines[credIdx])
		lines[credIdx] = strings.Repeat(" ", indent) + "credentials: " + value
	} else {
		parentIndent := leadingSpaces(lines[kitIdx])
		lines = insertAfter(lines, kitIdx, parentIndent+2, "credentials: "+value)
	}

	return writeLines(path, lines)
}

// findSection returns the line index of a top-level key (e.g. "agents:").
func findSection(lines []string, key string) int {
	for i, line := range lines {
		if strings.TrimSpace(line) == key+":" && leadingSpaces(line) == 0 {
			return i
		}
	}
	return -1
}

// findKey returns the line index of a key nested under parentIdx.
// Searches only within the parent's indented block.
func findKey(lines []string, parentIdx int, key string) int {
	parentIndent := leadingSpaces(lines[parentIdx])
	for i := parentIdx + 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		indent := leadingSpaces(lines[i])
		if indent <= parentIndent {
			break
		}
		if strings.HasPrefix(trimmed, key+":") && indent == parentIndent+2 {
			return i
		}
	}
	return -1
}

// insertAfter inserts a line with the given indent after idx.
func insertAfter(lines []string, idx, indent int, content string) []string {
	line := strings.Repeat(" ", indent) + content
	return append(lines[:idx+1], append([]string{line}, lines[idx+1:]...)...)
}

func leadingSpaces(s string) int {
	return len(s) - len(strings.TrimLeft(s, " "))
}

func writeLines(path string, lines []string) error {
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
}
