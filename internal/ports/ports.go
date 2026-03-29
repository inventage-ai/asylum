package ports

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

const BasePort = 10000

type Range struct {
	Project   string `json:"project"`
	Container string `json:"container"`
	Start     int    `json:"start"`
	Count     int    `json:"count"`
}

// Ports returns the individual port numbers in the range.
func (r Range) Ports() []int {
	p := make([]int, r.Count)
	for i := range r.Count {
		p[i] = r.Start + i
	}
	return p
}

type registry struct {
	Ranges []Range `json:"ranges"`
}

// Allocate returns the existing port range for projectDir, or allocates a new one.
// The containerName is stored alongside for reverse lookups during cleanup.
func Allocate(projectDir, containerName string, count int) (Range, error) {
	path, err := registryPath()
	if err != nil {
		return Range{}, err
	}

	f, err := openLocked(path)
	if err != nil {
		return Range{}, err
	}
	defer f.Close()

	reg, err := readRegistry(f)
	if err != nil {
		return Range{}, err
	}

	// Check for existing allocation
	for i, r := range reg.Ranges {
		if r.Project == projectDir {
			if r.Count >= count {
				return r, nil
			}
			// Try to extend
			end := r.Start + count
			if end > 65536 {
				return r, fmt.Errorf("cannot extend port range: would exceed 65535")
			}
			if canExtend(reg.Ranges, i, count) {
				reg.Ranges[i].Count = count
				if err := writeRegistry(f, reg); err != nil {
					return Range{}, err
				}
				return reg.Ranges[i], nil
			}
			// Can't extend, keep existing
			return r, nil
		}
	}

	// Allocate new range
	start := nextStart(reg.Ranges)
	if start+count > 65536 {
		return Range{}, fmt.Errorf("no ports available: allocation would exceed 65535")
	}

	r := Range{Project: projectDir, Container: containerName, Start: start, Count: count}
	reg.Ranges = append(reg.Ranges, r)

	if err := writeRegistry(f, reg); err != nil {
		return Range{}, err
	}
	return r, nil
}

// Release removes the port allocation for projectDir.
func Release(projectDir string) error {
	return release(func(r Range) bool { return r.Project == projectDir })
}

// ReleaseContainer removes the port allocation for the given container name.
func ReleaseContainer(containerName string) error {
	return release(func(r Range) bool { return r.Container == containerName })
}

func release(match func(Range) bool) error {
	path, err := registryPath()
	if err != nil {
		return err
	}

	f, err := openLocked(path)
	if err != nil {
		return err
	}
	defer f.Close()

	reg, err := readRegistry(f)
	if err != nil {
		return err
	}

	for i, r := range reg.Ranges {
		if match(r) {
			reg.Ranges = append(reg.Ranges[:i], reg.Ranges[i+1:]...)
			return writeRegistry(f, reg)
		}
	}
	return nil
}

func nextStart(ranges []Range) int {
	highest := BasePort
	for _, r := range ranges {
		if end := r.Start + r.Count; end > highest {
			highest = end
		}
	}
	return highest
}

func canExtend(ranges []Range, idx, newCount int) bool {
	r := ranges[idx]
	newEnd := r.Start + newCount
	for i, other := range ranges {
		if i == idx {
			continue
		}
		otherEnd := other.Start + other.Count
		if r.Start < otherEnd && newEnd > other.Start {
			return false
		}
	}
	return true
}

func registryPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home dir: %w", err)
	}
	dir := filepath.Join(home, ".asylum")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create asylum dir: %w", err)
	}
	return filepath.Join(dir, "ports.json"), nil
}

func openLocked(path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("open port registry: %w", err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		f.Close()
		return nil, fmt.Errorf("lock port registry: %w", err)
	}
	return f, nil
}

func readRegistry(f *os.File) (registry, error) {
	info, err := f.Stat()
	if err != nil {
		return registry{}, err
	}
	if info.Size() == 0 {
		return registry{}, nil
	}
	var reg registry
	if err := json.NewDecoder(f).Decode(&reg); err != nil {
		return registry{}, fmt.Errorf("parse port registry: %w", err)
	}
	return reg, nil
}

func writeRegistry(f *os.File, reg registry) error {
	if err := f.Truncate(0); err != nil {
		return fmt.Errorf("truncate port registry: %w", err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		return fmt.Errorf("seek port registry: %w", err)
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(reg)
}
