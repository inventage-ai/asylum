package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
)

// State holds machine-managed state persisted at ~/.asylum/state.json.
type State struct {
	KnownKits        []string `json:"known_kits"`
	DockerSourceOrder []string `json:"docker_source_order,omitempty"`
}

// LoadState reads state.json from the asylum directory. Returns an empty
// State if the file doesn't exist.
func LoadState(asylumDir string) (State, error) {
	path := filepath.Join(asylumDir, "state.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return State{}, nil
		}
		return State{}, err
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return State{}, err
	}
	return s, nil
}

// SaveState writes state.json to the asylum directory.
func SaveState(asylumDir string, s State) error {
	slices.Sort(s.KnownKits)
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(filepath.Join(asylumDir, "state.json"), data, 0644)
}

// NewKits returns kit names that are in registered but not in state.KnownKits.
func NewKits(registered []string, state State) []string {
	known := map[string]bool{}
	for _, name := range state.KnownKits {
		known[name] = true
	}
	var newKits []string
	for _, name := range registered {
		if !known[name] {
			newKits = append(newKits, name)
		}
	}
	return newKits
}
