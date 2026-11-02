package broker

import (
	"os/exec"
	"syscall"
	"testing"
	"time"
)

func TestKillGroups(t *testing.T) {
	cmd := exec.Command("sh", "-c", "sleep 60 & wait")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	TrackGroup(cmd.Process.Pid)
	KillGroups()

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("process group still running after KillGroups")
	}
}

func TestUntrackedGroupNotKilled(t *testing.T) {
	cmd := exec.Command("sleep", "60")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()
	TrackGroup(cmd.Process.Pid)()
	KillGroups()
	if err := syscall.Kill(cmd.Process.Pid, 0); err != nil {
		t.Errorf("untracked process was killed: %v", err)
	}
}
