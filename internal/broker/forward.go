package broker

import (
	"io"
	"net"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/inventage-ai/asylum/internal/log"
)

// forwardTTL is how long a loopback forward stays open with no repeat request.
const forwardTTL = 5 * time.Minute

// forwarder is the broker's best-effort loopback callback forwarding service.
// It bridges a host loopback port to the same port on the container's loopback
// so an OAuth callback opened in the host browser reaches a callback server
// listening inside the container. Each forward is time-boxed and torn down when
// the broker (and thus the container) stops.
type forwarder struct {
	cname  string
	mu     sync.Mutex
	active map[int]*forward
}

type forward struct {
	ln    net.Listener
	timer *time.Timer
}

func newForwarder(cname string) *forwarder {
	return &forwarder{cname: cname, active: map[int]*forward{}}
}

// ensure starts a forward for the loopback port, or resets its timer if one is
// already active. Binding failures are logged and skipped — the service is
// best-effort, and most auth flows also accept pasting a code or URL.
func (f *forwarder) ensure(port int, ipv6 bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if fw, ok := f.active[port]; ok {
		fw.timer.Reset(forwardTTL)
		return
	}

	ip := "127.0.0.1"
	if ipv6 {
		ip = "::1"
	}
	ln, err := net.Listen("tcp", net.JoinHostPort(ip, strconv.Itoa(port)))
	if err != nil {
		log.Warn("callback forward: cannot bind host loopback port %d: %v", port, err)
		return
	}

	fw := &forward{ln: ln, timer: time.AfterFunc(forwardTTL, func() { f.stop(port) })}
	f.active[port] = fw
	log.Info("callback forward: host loopback :%d → container for %s", port, forwardTTL)
	go f.accept(ln, port, ipv6)
}

func (f *forwarder) stop(port int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if fw, ok := f.active[port]; ok {
		fw.timer.Stop()
		fw.ln.Close()
		delete(f.active, port)
	}
}

func (f *forwarder) accept(ln net.Listener, port int, ipv6 bool) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return // listener closed on expiry or shutdown
		}
		go f.tunnel(conn, port, ipv6)
	}
}

// tunnel relays one connection into the container's loopback via `docker exec`
// + socat (already present in the image). No host port is published.
func (f *forwarder) tunnel(conn net.Conn, port int, ipv6 bool) {
	defer conn.Close()

	target := "TCP4:127.0.0.1:" + strconv.Itoa(port)
	if ipv6 {
		target = "TCP6:[::1]:" + strconv.Itoa(port)
	}
	cmd := exec.Command("docker", "exec", "-i", f.cname, "socat", "-", target)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	if err := cmd.Start(); err != nil {
		return
	}
	go func() {
		io.Copy(stdin, conn)
		stdin.Close()
	}()
	io.Copy(conn, stdout)
	cmd.Wait()
}
