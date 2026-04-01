package image

import (
	"cmp"
	"slices"

	"github.com/inventage-ai/asylum/internal/agent"
	"github.com/inventage-ai/asylum/internal/kit"
)

const defaultPriority = 50

// dockerSource pairs an identifier with its priority and snippet.
type dockerSource struct {
	ID       string
	Priority int
	Snippet  string
}

// collectSources builds a list of dockerfile sources from resolved kits and agents.
func collectSources(kits []*kit.Kit, agents []*agent.AgentInstall) []dockerSource {
	var sources []dockerSource
	for _, k := range kits {
		if k.DockerSnippet == "" {
			continue
		}
		p := k.DockerPriority
		if p == 0 {
			p = defaultPriority
		}
		sources = append(sources, dockerSource{
			ID:       "kit:" + k.Name,
			Priority: p,
			Snippet:  k.DockerSnippet,
		})
	}
	for _, a := range agents {
		if a.DockerSnippet == "" {
			continue
		}
		p := a.DockerPriority
		if p == 0 {
			p = defaultPriority
		}
		sources = append(sources, dockerSource{
			ID:       "agent:" + a.Name,
			Priority: p,
			Snippet:  a.DockerSnippet,
		})
	}
	return sources
}

// computeSourceOrder determines the Dockerfile snippet order given the
// active sources and the previous build's order (from state).
//
// Algorithm:
//  1. Retained sources (present in both active and previous) keep their
//     relative order from the previous build.
//  2. If any source from the previous order was removed, everything from
//     the earliest removal point onward is re-sorted by priority.
//  3. New sources are appended at the end, sorted by priority.
func computeSourceOrder(sources []dockerSource, previousOrder []string) []string {
	active := map[string]bool{}
	for _, s := range sources {
		active[s.ID] = true
	}

	priorityOf := map[string]int{}
	for _, s := range sources {
		priorityOf[s.ID] = s.Priority
	}

	// Find retained sources in their previous order, and detect earliest removal point.
	var retained []string
	removalIdx := -1
	for i, id := range previousOrder {
		if active[id] {
			retained = append(retained, id)
		} else if removalIdx == -1 {
			removalIdx = i
		}
	}

	// Find new sources (active but not in previous order).
	prev := map[string]bool{}
	for _, id := range previousOrder {
		prev[id] = true
	}
	var added []string
	for _, s := range sources {
		if !prev[s.ID] {
			added = append(added, s.ID)
		}
	}

	// If a removal occurred, re-sort the suffix after the earliest removal point.
	if removalIdx >= 0 {
		// Count how many retained sources fall before the removal point.
		prefixLen := 0
		for _, id := range previousOrder[:removalIdx] {
			if active[id] {
				prefixLen++
			}
		}
		suffix := retained[prefixLen:]
		sortByPriority(suffix, priorityOf)
		retained = append(retained[:prefixLen], suffix...)
	}

	// Sort new sources by priority and append.
	sortByPriority(added, priorityOf)

	return append(retained, added...)
}

func sortByPriority(ids []string, priorityOf map[string]int) {
	slices.SortStableFunc(ids, func(a, b string) int {
		if c := cmp.Compare(priorityOf[a], priorityOf[b]); c != 0 {
			return c
		}
		return cmp.Compare(a, b)
	})
}
