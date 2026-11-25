package broker

import (
	"sync"
	"syscall"
)

// groups holds the process groups that route handlers started, so that they
// can be killed when the broker exits: the broker leaves via os.Exit once its
// container stops, which would otherwise orphan them.
var groups sync.Map

// TrackGroup registers the process group led by pid and returns a function
// that unregisters it.
func TrackGroup(pid int) (untrack func()) {
	groups.Store(pid, struct{}{})
	return func() { groups.Delete(pid) }
}

// KillGroups kills every tracked process group.
func KillGroups() {
	groups.Range(func(pid, _ any) bool {
		syscall.Kill(-pid.(int), syscall.SIGKILL)
		groups.Delete(pid)
		return true
	})
}
