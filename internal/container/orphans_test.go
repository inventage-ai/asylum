package container

import (
	"maps"
	"slices"
	"testing"

	"github.com/inventage-ai/asylum/internal/docker"
)

func TestVolumeOwner(t *testing.T) {
	tests := map[string]string{
		"asylum-abc-app-cache-npm":          "asylum-abc-app",
		"asylum-abc-app-npm-0123456789a":    "asylum-abc-app",
		"asylum-abc-my-cache-app-cache-pip": "asylum-abc-my-cache-app",
		"asylum-abc-my-npm-app-npm-0123":    "asylum-abc-my-npm-app",
		"asylum-abc":                        "",
	}
	for in, want := range tests {
		if got := volumeOwner(in); got != want {
			t.Errorf("volumeOwner(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestClassifyOrphans(t *testing.T) {
	existing := map[string]bool{"/code/app": true}
	baseInput := func() OrphanInput {
		return OrphanInput{
			Registry:      map[string]string{},
			ProjectExists: func(dir string) bool { return existing[dir] },
			ExpectedNodeModules: func(cname, _ string) map[string]string {
				return map[string]string{cname + "-npm-keep": "/code/app/node_modules"}
			},
		}
	}
	names := func(vs []OrphanVolume) []string {
		var out []string
		for _, v := range vs {
			out = append(out, v.Name)
		}
		return out
	}

	t.Run("project directory deleted", func(t *testing.T) {
		in := baseInput()
		in.Volumes = map[string]string{"asylum-old-old-cache-npm": "/code/old"}
		r := ClassifyOrphans(in)
		if !slices.Equal(names(r.Removable), []string{"asylum-old-old-cache-npm"}) || r.Removable[0].Reason != ReasonProjectMissing {
			t.Errorf("unexpected removable %+v", r.Removable)
		}
		if r.MissingProjects["asylum-old-old"] != "/code/old" {
			t.Errorf("expected missing project, got %v", r.MissingProjects)
		}
	})

	t.Run("stale node_modules volume of existing project", func(t *testing.T) {
		in := baseInput()
		in.Volumes = map[string]string{
			"asylum-app-app-npm-keep":  "/code/app",
			"asylum-app-app-npm-gone":  "/code/app",
			"asylum-app-app-cache-npm": "/code/app",
		}
		r := ClassifyOrphans(in)
		if !slices.Equal(names(r.Removable), []string{"asylum-app-app-npm-gone"}) || r.Removable[0].Reason != ReasonNodeModulesGone {
			t.Errorf("unexpected removable %+v", r.Removable)
		}
		if len(r.MissingProjects) != 0 {
			t.Errorf("existing project must not be missing: %v", r.MissingProjects)
		}
	})

	t.Run("unlabeled volume resolved via port registry", func(t *testing.T) {
		in := baseInput()
		in.Volumes = map[string]string{"asylum-abc123def456-old-cache-npm": ""}
		in.Registry = map[string]string{"asylum-abc123def456-old": "/code/old"}
		r := ClassifyOrphans(in)
		if !slices.Equal(names(r.Removable), []string{"asylum-abc123def456-old-cache-npm"}) {
			t.Errorf("unexpected removable %+v", r.Removable)
		}
	})

	t.Run("unlabeled old-format volume resolved via port registry", func(t *testing.T) {
		in := baseInput()
		old := OldContainerName("/code/old")
		in.Volumes = map[string]string{old + "-cache-npm": ""}
		in.Registry = map[string]string{ContainerName("/code/old"): "/code/old"}
		r := ClassifyOrphans(in)
		if !slices.Equal(names(r.Removable), []string{old + "-cache-npm"}) {
			t.Errorf("unexpected removable %+v", r.Removable)
		}
	})

	t.Run("unresolvable volume", func(t *testing.T) {
		in := baseInput()
		in.Volumes = map[string]string{"asylum-zzz-x-cache-npm": "", "asylum-weird": ""}
		r := ClassifyOrphans(in)
		if len(r.Removable) != 0 || !slices.Equal(r.Unknown, []string{"asylum-weird", "asylum-zzz-x-cache-npm"}) {
			t.Errorf("unexpected report %+v", r)
		}
	})

	t.Run("volume of running container", func(t *testing.T) {
		in := baseInput()
		in.Volumes = map[string]string{"asylum-old-old-cache-npm": "/code/old"}
		in.Containers = []docker.ContainerMounts{{Name: "asylum-old-old", Running: true, Mounts: []string{"asylum-old-old-cache-npm"}}}
		r := ClassifyOrphans(in)
		if len(r.Removable) != 0 || !slices.Equal(names(r.InUse), []string{"asylum-old-old-cache-npm"}) {
			t.Errorf("unexpected report %+v", r)
		}
		if len(r.MissingProjects) != 0 || len(r.StoppedContainers) != 0 {
			t.Errorf("running project must not be cleaned up: %+v", r)
		}
	})

	t.Run("deleted project leftovers", func(t *testing.T) {
		in := baseInput()
		in.Volumes = map[string]string{"asylum-old-old-cache-npm": "/code/old"}
		in.Containers = []docker.ContainerMounts{{Name: "asylum-old-old", Mounts: []string{"/code/old"}}}
		r := ClassifyOrphans(in)
		if !slices.Equal(r.StoppedContainers, []string{"asylum-old-old"}) {
			t.Errorf("expected stopped container removal, got %v", r.StoppedContainers)
		}
		if !maps.Equal(r.MissingProjects, map[string]string{"asylum-old-old": "/code/old"}) {
			t.Errorf("unexpected missing projects %v", r.MissingProjects)
		}
	})

	t.Run("stopped container holds stale node_modules volume", func(t *testing.T) {
		in := baseInput()
		in.Volumes = map[string]string{"asylum-app-app-npm-gone": "/code/app"}
		in.Containers = []docker.ContainerMounts{{Name: "asylum-app-app", Mounts: []string{"asylum-app-app-npm-gone"}}}
		r := ClassifyOrphans(in)
		if !slices.Equal(r.StoppedContainers, []string{"asylum-app-app"}) || len(r.Removable) != 1 {
			t.Errorf("unexpected report %+v", r)
		}
	})
}
