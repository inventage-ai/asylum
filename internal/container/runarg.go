package container

import (
	"fmt"
	"slices"
	"strings"

	"github.com/inventage-ai/asylum/internal/kit"
)

// booleanFlags are docker flags that take no value.
var booleanFlags = map[string]bool{
	"--privileged": true,
	"--rm":         true,
	"--init":       true,
	"-it":          true,
	"-i":           true,
}

// singleValueFlags are docker flags where the flag name itself is the dedup key
// (only one value allowed per flag).
var singleValueFlags = map[string]bool{
	"--name":     true,
	"--hostname": true,
	"-w":         true,
	"--add-host": true,
	"--env-file": true,
}

// dedupKey returns a (category, key) pair used for deduplication.
// Args with the same category+key are candidates for dedup/conflict.
func dedupKey(r kit.RunArg) (string, string) {
	switch r.Flag {
	case "-p":
		// Container port is the right side of host:container
		if i := strings.LastIndex(r.Value, ":"); i >= 0 {
			return "-p", r.Value[i+1:]
		}
		return "-p", r.Value

	case "-v":
		// Container path is the second colon-delimited segment
		parts := strings.SplitN(r.Value, ":", 3)
		if len(parts) >= 2 {
			return "-v", parts[1]
		}
		return "-v", r.Value

	case "--mount":
		// Parse dst= from comma-separated key=value pairs
		for _, kv := range strings.Split(r.Value, ",") {
			k, v, ok := strings.Cut(kv, "=")
			if ok && k == "dst" {
				return "--mount", v
			}
		}
		return "--mount", r.Value

	case "-e":
		// Env var name is left of =
		if k, _, ok := strings.Cut(r.Value, "="); ok {
			return "-e", k
		}
		return "-e", r.Value

	case "--cap-add":
		return "--cap-add", r.Value
	}

	if booleanFlags[r.Flag] {
		return "bool", r.Flag
	}

	if singleValueFlags[r.Flag] {
		return "single", r.Flag
	}

	// Unknown flags: no dedup (each is unique)
	return "other", r.Flag + "=" + r.Value
}

func dedupCategoryKey(r kit.RunArg) string {
	cat, key := dedupKey(r)
	return cat + ":" + key
}

// ResolveArgs collects RunArgs, deduplicates by key (higher priority wins),
// and detects conflicts (same priority, same key, different value).
// Returns the resolved args in deterministic order and any overrides that occurred.
func ResolveArgs(args []kit.RunArg) ([]kit.RunArg, []kit.Override, error) {
	type entry struct {
		arg   kit.RunArg
		index int
	}

	best := map[string]entry{}
	var overrides []kit.Override

	for i, arg := range args {
		ck := dedupCategoryKey(arg)
		existing, exists := best[ck]
		if !exists {
			best[ck] = entry{arg: arg, index: i}
			continue
		}

		if arg.Priority > existing.arg.Priority {
			overrides = append(overrides, kit.Override{Replaced: existing.arg, Winner: arg})
			best[ck] = entry{arg: arg, index: i}
		} else if arg.Priority == existing.arg.Priority {
			if arg.Value == existing.arg.Value {
				continue
			}
			return nil, nil, fmt.Errorf(
				"conflicting docker args for %s: %q (from %s) vs %q (from %s)",
				ck, existing.arg.Value, existing.arg.Source, arg.Value, arg.Source,
			)
		} else {
			overrides = append(overrides, kit.Override{Replaced: arg, Winner: existing.arg})
		}
	}

	result := make([]kit.RunArg, 0, len(best))
	for _, e := range best {
		result = append(result, e.arg)
	}
	slices.SortFunc(result, func(a, b kit.RunArg) int {
		if a.Priority != b.Priority {
			return a.Priority - b.Priority
		}
		if a.Source != b.Source {
			return strings.Compare(a.Source, b.Source)
		}
		return strings.Compare(dedupCategoryKey(a), dedupCategoryKey(b))
	})

	return result, overrides, nil
}

// FormatDebug formats resolved RunArgs and overrides as a debug table for stderr.
func FormatDebug(resolved []kit.RunArg, overrides []kit.Override) string {
	var b strings.Builder
	b.WriteString("Docker run arguments:\n")

	// Find max width for alignment
	maxWidth := 0
	for _, a := range resolved {
		w := len(a.Flag)
		if a.Value != "" {
			w += 1 + len(a.Value)
		}
		if w > maxWidth {
			maxWidth = w
		}
	}
	if maxWidth > 60 {
		maxWidth = 60
	}

	for _, a := range resolved {
		s := a.Flag
		if a.Value != "" {
			s += " " + a.Value
		}
		// Truncate long values for readability
		display := s
		if len(display) > 60 {
			display = display[:57] + "..."
		}
		fmt.Fprintf(&b, "  %-*s  %s\n", maxWidth, display, a.Source)
	}

	if len(overrides) > 0 {
		b.WriteString("\n  Overrides (higher priority won):\n")
		for _, o := range overrides {
			fmt.Fprintf(&b, "    %s (%s) → %s (%s)\n",
				o.Replaced.String(), o.Replaced.Source,
				o.Winner.String(), o.Winner.Source,
			)
		}
	}

	return b.String()
}

// FlattenArgs converts resolved RunArgs to the []string slice for docker run.
func FlattenArgs(args []kit.RunArg) []string {
	var result []string
	for _, a := range args {
		result = append(result, a.Flag)
		if a.Value != "" {
			result = append(result, a.Value)
		}
	}
	return result
}
