package broker

import (
	"net"
	"testing"
)

func TestForwarderReuseAndStop(t *testing.T) {
	p, err := FreePort()
	if err != nil {
		t.Fatal(err)
	}
	f := newForwarder("test-container")
	f.ensure(p, false)
	if len(f.active) != 1 {
		t.Fatalf("after ensure: active=%d, want 1", len(f.active))
	}
	f.ensure(p, false) // repeat request resets the timer, adds no new entry
	if len(f.active) != 1 {
		t.Fatalf("after repeat ensure: active=%d, want 1", len(f.active))
	}
	f.stop(p)
	if len(f.active) != 0 {
		t.Fatalf("after stop: active=%d, want 0", len(f.active))
	}
}

func TestForwarderBindFailureSkips(t *testing.T) {
	// Occupy a loopback port so the forwarder cannot bind it.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port

	f := newForwarder("test-container")
	f.ensure(port, false)
	if len(f.active) != 0 {
		t.Fatalf("bind failure should skip: active=%d, want 0", len(f.active))
	}
}
