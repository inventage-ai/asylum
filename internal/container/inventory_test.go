package container

import (
	"slices"
	"testing"
	"time"

	"github.com/inventage-ai/asylum/internal/docker"
)

func TestInventory(t *testing.T) {
	now := time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)
	day := 24 * time.Hour
	app, old := ContainerName("/p/app"), ContainerName("/p/old")
	exists := func(dir string) bool { return dir != "/p/gone" }

	tests := []struct {
		name string
		in   InventoryInput
		want []Project
	}{
		{
			name: "recorded idle project",
			in: InventoryInput{
				Records:  map[string]Usage{app: {Project: "/p/app", LastUsed: now.Add(-3 * day)}},
				DataDirs: map[string]DataDir{app: {ModTime: now.Add(-100 * day)}},
			},
			want: []Project{{Path: "/p/app", Containers: []string{app}, Status: StatusIdle, LastUsed: now.Add(-3 * day), Size: -1}},
		},
		{
			name: "running project",
			in: InventoryInput{
				Records:    map[string]Usage{app: {Project: "/p/app", LastUsed: now.Add(-3 * day)}},
				Containers: []docker.ContainerMounts{{Name: app, Running: true}},
			},
			want: []Project{{Path: "/p/app", Containers: []string{app}, Status: StatusRunning, LastUsed: now, Size: -1}},
		},
		{
			name: "deleted project with stopped container is missing",
			in: InventoryInput{
				Records:    map[string]Usage{"asylum-x-gone": {Project: "/p/gone", LastUsed: now.Add(-3 * day)}},
				Containers: []docker.ContainerMounts{{Name: "asylum-x-gone"}},
			},
			want: []Project{{Path: "/p/gone", Containers: []string{"asylum-x-gone"}, Status: StatusMissing, LastUsed: now.Add(-3 * day), Size: -1}},
		},
		{
			name: "stopped container",
			in: InventoryInput{
				Registry:   map[string]string{app: "/p/app"},
				Containers: []docker.ContainerMounts{{Name: app}},
			},
			want: []Project{{Path: "/p/app", Containers: []string{app}, Status: StatusStopped, Size: -1}},
		},
		{
			name: "known only from port registry, last use from data dir",
			in: InventoryInput{
				Registry: map[string]string{old: "/p/old"},
				DataDirs: map[string]DataDir{old: {ModTime: now.Add(-40 * day)}},
			},
			want: []Project{{Path: "/p/old", Containers: []string{old}, Status: StatusIdle, LastUsed: now.Add(-40 * day), Size: -1}},
		},
		{
			name: "unknown path listed by container name",
			in: InventoryInput{
				DataDirs: map[string]DataDir{"asylum-123456789abc-mystery": {}},
				Volumes:  map[string]string{"asylum-123456789abc-mystery-cache-npm": ""},
			},
			want: []Project{{Containers: []string{"asylum-123456789abc-mystery"}, Status: StatusIdle, Size: -1}},
		},
		{
			name: "secondary container and old-format volumes merged by path",
			in: InventoryInput{
				Records: map[string]Usage{app: {Project: "/p/app"}},
				Volumes: map[string]string{
					"asylum-sec-app-cache-npm":                   "/p/app",
					OldContainerName("/p/app") + "-cache-gradle": "",
				},
			},
			want: []Project{{Path: "/p/app", Containers: slices.Sorted(slices.Values([]string{app, "asylum-sec-app", OldContainerName("/p/app")})), Status: StatusIdle, Size: -1}},
		},
		{
			name: "size counts volumes, data and an image no other project records",
			in: InventoryInput{
				Records: map[string]Usage{
					app: {Project: "/p/app", LastUsed: now.Add(-2 * day), Image: "asylum:proj-own"},
					old: {Project: "/p/old", LastUsed: now.Add(-1 * day), Image: "asylum:latest"},
				},
				DataDirs:    map[string]DataDir{app: {Size: 5}},
				Volumes:     map[string]string{app + "-cache-npm": "/p/app", old + "-cache-npm": "/p/old"},
				VolumeSizes: map[string]int64{app + "-cache-npm": 100, old + "-cache-npm": 7},
				ImageSizes:  map[string]int64{"asylum:proj-own": 1000, "asylum:latest": 5000},
			},
			want: []Project{
				{Path: "/p/app", Containers: []string{app}, Status: StatusIdle, LastUsed: now.Add(-2 * day), Size: 1105},
				{Path: "/p/old", Containers: []string{old}, Status: StatusIdle, LastUsed: now.Add(-1 * day), Size: 7},
			},
		},
		{
			name: "shared image counted for neither project",
			in: InventoryInput{
				Records: map[string]Usage{
					app: {Project: "/p/app", Image: "asylum:proj-shared"},
					old: {Project: "/p/old", Image: "asylum:proj-shared"},
				},
				VolumeSizes: map[string]int64{},
				ImageSizes:  map[string]int64{"asylum:proj-shared": 1000},
			},
			want: []Project{
				{Path: "/p/app", Containers: []string{app}, Status: StatusIdle},
				{Path: "/p/old", Containers: []string{old}, Status: StatusIdle},
			},
		},
		{
			name: "sorted oldest first, unknown last use at the end",
			in: InventoryInput{
				Records: map[string]Usage{
					app: {Project: "/p/app", LastUsed: now.Add(-1 * day)},
					old: {Project: "/p/old", LastUsed: now.Add(-9 * day)},
				},
				DataDirs: map[string]DataDir{"asylum-123456789abc-mystery": {}},
			},
			want: []Project{
				{Path: "/p/old", Containers: []string{old}, Status: StatusIdle, LastUsed: now.Add(-9 * day), Size: -1},
				{Path: "/p/app", Containers: []string{app}, Status: StatusIdle, LastUsed: now.Add(-1 * day), Size: -1},
				{Containers: []string{"asylum-123456789abc-mystery"}, Status: StatusIdle, Size: -1},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.in.Now = now
			tt.in.ProjectExists = exists
			got := Inventory(tt.in)
			if !slices.EqualFunc(got, tt.want, func(a, b Project) bool {
				return a.Path == b.Path && slices.Equal(a.Containers, b.Containers) && a.Status == b.Status && a.LastUsed.Equal(b.LastUsed) && a.Size == b.Size
			}) {
				t.Errorf("got  %+v\nwant %+v", got, tt.want)
			}
		})
	}
}

func TestProjectStale(t *testing.T) {
	now := time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		p    Project
		want bool
	}{
		{"missing", Project{Status: StatusMissing, LastUsed: now}, true},
		{"used 45 days ago", Project{Status: StatusIdle, LastUsed: now.Add(-45 * 24 * time.Hour)}, true},
		{"used 3 days ago", Project{Status: StatusIdle, LastUsed: now.Add(-3 * 24 * time.Hour)}, false},
		{"unknown last use", Project{Status: StatusIdle}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.Stale(now); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
