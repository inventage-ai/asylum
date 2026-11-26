package container

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/inventage-ai/asylum/internal/docker"
)

func TestUsageRoundTrip(t *testing.T) {
	projects := t.TempDir()
	want := Usage{Project: "/home/u/app", LastUsed: time.Date(2026, 11, 25, 10, 0, 0, 0, time.UTC), Image: "asylum:proj-1a2b3c4d5e6f"}
	if err := WriteUsage(filepath.Join(projects, "asylum-a-app"), want); err != nil {
		t.Fatal(err)
	}
	os.MkdirAll(filepath.Join(projects, "asylum-b-old"), 0755)

	got, err := ReadAllUsage(projects)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || !got["asylum-a-app"].LastUsed.Equal(want.LastUsed) || got["asylum-a-app"].Project != want.Project || got["asylum-a-app"].Image != want.Image {
		t.Errorf("got %+v, want only asylum-a-app = %+v", got, want)
	}
}

func TestScanDataDirs(t *testing.T) {
	projects := t.TempDir()
	dir := filepath.Join(projects, "asylum-a-app", "history")
	os.MkdirAll(dir, 0755)
	os.WriteFile(filepath.Join(dir, "h"), make([]byte, 300), 0644)
	os.WriteFile(filepath.Join(projects, "asylum-a-app", "x"), make([]byte, 200), 0644)
	os.WriteFile(filepath.Join(projects, "stray-file"), nil, 0644)

	got, err := ScanDataDirs(projects)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got["asylum-a-app"].Size != 500 || got["asylum-a-app"].ModTime.IsZero() {
		t.Errorf("got %+v", got)
	}
}

func TestScanDataDirsMissing(t *testing.T) {
	got, err := ScanDataDirs(filepath.Join(t.TempDir(), "nope"))
	if err != nil || len(got) != 0 {
		t.Errorf("got %v, %v", got, err)
	}
}

func TestUnusedProjectImages(t *testing.T) {
	remaining := map[string]Usage{"asylum-b-other": {Image: "asylum:proj-shared"}}
	containers := []docker.ContainerMounts{{Name: "asylum-c-run", Image: "asylum:proj-running"}}
	tests := []struct {
		name       string
		candidates []string
		want       []string
	}{
		{"unused", []string{"asylum:proj-gone"}, []string{"asylum:proj-gone"}},
		{"shared with remaining project", []string{"asylum:proj-shared"}, nil},
		{"used by container", []string{"asylum:proj-running"}, nil},
		{"base image", []string{"asylum:latest"}, nil},
		{"no record", []string{""}, nil},
		{"two removed projects sharing an image", []string{"asylum:proj-gone", "asylum:proj-gone"}, []string{"asylum:proj-gone"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UnusedProjectImages(tt.candidates, remaining, containers); !slices.Equal(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
