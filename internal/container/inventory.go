package container

import (
	"cmp"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/inventage-ai/asylum/internal/docker"
)

const (
	StatusRunning = "running"
	StatusStopped = "stopped"
	StatusIdle    = "idle"
	StatusMissing = "missing"
)

// StaleAfter is how long a project may go unused before cleanup suggests it.
const StaleAfter = 30 * 24 * time.Hour

// Project is one asylum project, assembled from the resources asylum keeps for it.
type Project struct {
	Path       string   // "" when unknown
	Containers []string // container names its resources were created under
	Status     string
	LastUsed   time.Time // zero when unknown
	Size       int64     // -1 when Docker reported no sizes
}

func (p Project) Name() string {
	return cmp.Or(p.Path, p.Containers[0])
}

// Stale reports whether the project is a cleanup candidate.
func (p Project) Stale(now time.Time) bool {
	return p.Status == StatusMissing || !p.LastUsed.IsZero() && now.Sub(p.LastUsed) > StaleAfter
}

type InventoryInput struct {
	Now           time.Time
	Records       map[string]Usage   // container name → usage record
	DataDirs      map[string]DataDir // container name → project data directory
	Registry      map[string]string  // container name → project dir, from the port registry
	Volumes       map[string]string  // volume name → asylum.project label ("" when unlabeled)
	Containers    []docker.ContainerMounts
	VolumeSizes   map[string]int64 // nil when Docker reported no sizes
	ImageSizes    map[string]int64
	ProjectExists func(dir string) bool
}

// Inventory lists every project asylum has resources for, oldest use first
// and projects with unknown last use at the end.
func Inventory(in InventoryInput) []Project {
	paths := projectPaths(in)

	groups := map[string]*Project{}
	add := func(cname string) {
		key := cmp.Or(paths[cname], cname)
		p, ok := groups[key]
		if !ok {
			p = &Project{Path: paths[cname]}
			groups[key] = p
		}
		if !slices.Contains(p.Containers, cname) {
			p.Containers = append(p.Containers, cname)
		}
	}
	for cname := range in.Records {
		add(cname)
	}
	for cname := range in.DataDirs {
		add(cname)
	}
	for cname := range in.Registry {
		add(cname)
	}
	for name := range in.Volumes {
		if owner := volumeOwner(name); owner != "" {
			add(owner)
		}
	}
	for _, c := range in.Containers {
		add(c.Name)
	}

	imageUsers := map[string]map[*Project]bool{}
	for _, p := range groups {
		for _, cname := range p.Containers {
			img := in.Records[cname].Image
			if imageUsers[img] == nil {
				imageUsers[img] = map[*Project]bool{}
			}
			imageUsers[img][p] = true
		}
	}

	projects := make([]Project, 0, len(groups))
	for _, p := range groups {
		slices.Sort(p.Containers)
		p.Status = status(in, p)
		p.LastUsed = lastUsed(in, p)
		p.Size = size(in, p, imageUsers)
		projects = append(projects, *p)
	}
	slices.SortFunc(projects, func(a, b Project) int {
		if a.LastUsed.IsZero() != b.LastUsed.IsZero() {
			if a.LastUsed.IsZero() {
				return 1
			}
			return -1
		}
		return cmp.Or(a.LastUsed.Compare(b.LastUsed), cmp.Compare(a.Name(), b.Name()))
	})
	return projects
}

// projectPaths maps container names to project paths. Usage records are the
// most reliable source, then the port registry, then volume labels.
func projectPaths(in InventoryInput) map[string]string {
	paths := map[string]string{}
	set := func(cname, path string) {
		if path != "" && paths[cname] == "" {
			paths[cname] = path
		}
	}
	for cname, u := range in.Records {
		set(cname, u.Project)
	}
	for cname, path := range in.Registry {
		set(cname, path)
	}
	for name, label := range in.Volumes {
		set(volumeOwner(name), label)
	}
	// Resources created before the basename was added to container names
	for _, path := range slices.Collect(maps.Values(paths)) {
		set(OldContainerName(path), path)
	}
	return paths
}

func status(in InventoryInput, p *Project) string {
	exists := false
	for _, c := range in.Containers {
		if slices.Contains(p.Containers, c.Name) {
			if c.Running {
				return StatusRunning
			}
			exists = true
		}
	}
	switch {
	case p.Path != "" && !in.ProjectExists(p.Path):
		return StatusMissing
	case exists:
		return StatusStopped
	default:
		return StatusIdle
	}
}

func lastUsed(in InventoryInput, p *Project) time.Time {
	if p.Status == StatusRunning {
		return in.Now
	}
	var recorded, modified time.Time
	for _, cname := range p.Containers {
		if u, ok := in.Records[cname]; ok && u.LastUsed.After(recorded) {
			recorded = u.LastUsed
		}
		if d, ok := in.DataDirs[cname]; ok && d.ModTime.After(modified) {
			modified = d.ModTime
		}
	}
	return cmp.Or(recorded, modified)
}

func size(in InventoryInput, p *Project, imageUsers map[string]map[*Project]bool) int64 {
	if in.VolumeSizes == nil {
		return -1
	}
	var total int64
	for name := range in.Volumes {
		if slices.Contains(p.Containers, volumeOwner(name)) {
			total += in.VolumeSizes[name]
		}
	}
	var images []string
	for _, cname := range p.Containers {
		total += in.DataDirs[cname].Size
		img := in.Records[cname].Image
		if strings.HasPrefix(img, "asylum:proj-") && len(imageUsers[img]) == 1 && !slices.Contains(images, img) {
			images = append(images, img)
			total += in.ImageSizes[img]
		}
	}
	return total
}
