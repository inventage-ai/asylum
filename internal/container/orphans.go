package container

import (
	"cmp"
	"maps"
	"slices"
	"strings"

	"github.com/inventage-ai/asylum/internal/docker"
)

const (
	ReasonProjectMissing  = "project missing"
	ReasonNodeModulesGone = "node_modules gone"
)

type OrphanVolume struct {
	Name      string
	Project   string
	Container string // container name the volume was created for
	Reason    string
}

type OrphanInput struct {
	Volumes             map[string]string // volume name → asylum.project label ("" when unlabeled)
	Registry            map[string]string // container name → project dir, from the port registry
	Containers          []docker.ContainerMounts
	ProjectExists       func(dir string) bool
	ExpectedNodeModules func(cname, projectDir string) map[string]string
}

type OrphanReport struct {
	Removable         []OrphanVolume
	InUse             []OrphanVolume
	Unknown           []string
	StoppedContainers []string          // must be removed before their volumes
	MissingProjects   map[string]string // container name → project dir whose leftovers can go
}

// ClassifyOrphans decides which asylum volumes no longer belong to anything.
func ClassifyOrphans(in OrphanInput) OrphanReport {
	registry := map[string]string{}
	for cname, project := range in.Registry {
		registry[cname] = project
		// Volumes created before the basename was added to container names
		registry[OldContainerName(project)] = project
	}

	mountedBy := map[string][]docker.ContainerMounts{}
	for _, c := range in.Containers {
		for _, m := range c.Mounts {
			mountedBy[m] = append(mountedBy[m], c)
		}
	}

	report := OrphanReport{MissingProjects: map[string]string{}}
	exists := map[string]bool{}
	expected := map[string]map[string]string{}
	stopped := map[string]bool{}

	for _, name := range slices.Sorted(maps.Keys(in.Volumes)) {
		cname := volumeOwner(name)
		project := cmp.Or(in.Volumes[name], registry[cname])
		if cname == "" || project == "" {
			report.Unknown = append(report.Unknown, name)
			continue
		}

		if _, ok := exists[project]; !ok {
			exists[project] = in.ProjectExists(project)
		}
		v := OrphanVolume{Name: name, Project: project, Container: cname}
		switch {
		case !exists[project]:
			v.Reason = ReasonProjectMissing
		case strings.HasPrefix(name, cname+"-npm-"):
			if expected[cname] == nil {
				expected[cname] = in.ExpectedNodeModules(cname, project)
			}
			if _, ok := expected[cname][name]; ok {
				continue
			}
			v.Reason = ReasonNodeModulesGone
		default:
			continue
		}

		if slices.ContainsFunc(mountedBy[name], func(c docker.ContainerMounts) bool { return c.Running }) {
			report.InUse = append(report.InUse, v)
			continue
		}
		for _, c := range mountedBy[name] {
			stopped[c.Name] = true
		}
		report.Removable = append(report.Removable, v)
		if v.Reason == ReasonProjectMissing {
			report.MissingProjects[cname] = project
		}
	}

	// A project whose volume is still in use is not gone yet as far as asylum is concerned.
	for _, v := range report.InUse {
		delete(report.MissingProjects, v.Container)
	}
	for _, c := range in.Containers {
		if _, ok := report.MissingProjects[c.Name]; ok && !c.Running {
			stopped[c.Name] = true
		}
	}
	report.StoppedContainers = slices.Sorted(maps.Keys(stopped))
	return report
}

// volumeOwner returns the container name a volume was created for, derived
// from the <container>-cache-<tool> and <container>-npm-<hash> naming.
func volumeOwner(volume string) string {
	i := max(strings.LastIndex(volume, "-cache-"), strings.LastIndex(volume, "-npm-"))
	if i <= 0 {
		return ""
	}
	return volume[:i]
}
