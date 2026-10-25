package kit

import (
	"bufio"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/inventage-ai/asylum/internal/broker"
)

// hostMCPHome sets HOME to a temp dir whose global config holds the given
// YAML, and returns the home.
func hostMCPHome(t *testing.T, config string) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.MkdirAll(filepath.Join(home, ".asylum"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".asylum", "config.yaml"), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	return home
}

func TestLoadHostMCPServers(t *testing.T) {
	tests := []struct {
		name   string
		config string // empty means no config file
		want   []string
	}{
		{"missing file", "", nil},
		{"no kit", "kits:\n  java: {}\n", nil},
		{"kit without servers", "kits:\n  host-mcp:\n", nil},
		{"servers", "kits:\n  host-mcp:\n    servers:\n      a:\n        command: [/bin/a]\n      b:\n        command: [/bin/b]\n", []string{"a", "b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config == "" {
				t.Setenv("HOME", t.TempDir())
			} else {
				hostMCPHome(t, tt.config)
			}
			servers, err := loadHostMCPServers()
			if err != nil {
				t.Fatal(err)
			}
			var names []string
			for name := range servers {
				names = append(names, name)
			}
			slices.Sort(names)
			if !slices.Equal(names, tt.want) {
				t.Errorf("servers = %v, want %v", names, tt.want)
			}
		})
	}
}

func TestLoadHostMCPServersIgnoresProjectConfig(t *testing.T) {
	hostMCPHome(t, "kits:\n  host-mcp:\n")
	project := t.TempDir()
	t.Chdir(project)
	if err := os.WriteFile(filepath.Join(project, ".asylum"), []byte("kits:\n  host-mcp:\n    servers:\n      evil:\n        command: [/bin/sh]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	servers, err := loadHostMCPServers()
	if err != nil {
		t.Fatal(err)
	}
	if len(servers) != 0 {
		t.Errorf("project-defined servers leaked: %v", servers)
	}
}

func TestLookupHostMCPServer(t *testing.T) {
	servers := map[string]hostMCPServer{
		"ok":       {Command: []string{"/bin/cat"}},
		"relative": {Command: []string{"npx", "some-mcp"}},
		"empty":    {},
		"pinned":   {Command: []string{"/bin/cat"}, Projects: []string{"/code/app/"}},
	}
	tests := []struct {
		name, server, project string
		wantStatus            int
	}{
		{"allowed", "ok", "/code/other", 0},
		{"unknown", "nope", "/code/app", http.StatusNotFound},
		{"relative command", "relative", "/code/app", http.StatusForbidden},
		{"no command", "empty", "/code/app", http.StatusForbidden},
		{"project allowed", "pinned", "/code/app", 0},
		{"project not allowed", "pinned", "/code/other", http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, status, err := lookupHostMCPServer(servers, tt.server, tt.project)
			if status != tt.wantStatus {
				t.Errorf("status = %d (%v), want %d", status, err, tt.wantStatus)
			}
			if (err != nil) != (tt.wantStatus != 0) {
				t.Errorf("err = %v, want error: %v", err, tt.wantStatus != 0)
			}
		})
	}
}

func TestHostMCPEnv(t *testing.T) {
	t.Setenv("HOME", "/home/u")
	t.Setenv("PATH", "/usr/bin")
	t.Setenv("LANG", "C")
	t.Setenv("ASYLUM_BROKER_TOKEN", "secret")
	t.Setenv("SOME_SHELL_VAR", "x")

	env := hostMCPEnv(map[string]string{"LANG": "de_CH.UTF-8", "FOO": "bar"})
	for _, want := range []string{"HOME=/home/u", "PATH=/usr/bin", "LANG=de_CH.UTF-8", "FOO=bar"} {
		if !slices.Contains(env, want) {
			t.Errorf("env missing %q: %v", want, env)
		}
	}
	for _, e := range env {
		if strings.HasPrefix(e, "ASYLUM_BROKER_TOKEN=") || strings.HasPrefix(e, "SOME_SHELL_VAR=") {
			t.Errorf("env leaks %q", e)
		}
	}
}

func TestHostMCPSnippets(t *testing.T) {
	k := Get("host-mcp")
	if k == nil {
		t.Fatal("host-mcp kit not registered")
	}
	if k.Tier != TierOptIn {
		t.Errorf("tier = %v, want TierOptIn", k.Tier)
	}
	for _, want := range []string{"/usr/local/bin/asylum-mcp", "/usr/local/lib/asylum-mcp.js"} {
		if !strings.Contains(k.DockerSnippet, want) {
			t.Errorf("DockerSnippet does not install %s", want)
		}
	}
	if !strings.Contains(k.RulesSnippet, "asylum-mcp <name>") {
		t.Error("RulesSnippet does not explain asylum-mcp <name>")
	}
}

// startHostMCPBroker serves the host-mcp route for projectDir on a fresh TCP
// endpoint and returns the endpoint.
func startHostMCPBroker(t *testing.T, ep broker.Endpoint, projectDir string) {
	t.Helper()
	go broker.Serve("test-container", projectDir, ep, "tok", Get("host-mcp").Routes)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if c, err := net.Dial(ep.Network, ep.Addr); err == nil {
			c.Close()
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("broker did not come up")
}

func tcpBrokerEndpoint(t *testing.T) broker.Endpoint {
	t.Helper()
	p, err := broker.FreePort()
	if err != nil {
		t.Fatal(err)
	}
	return broker.Endpoint{Network: "tcp", Addr: "127.0.0.1:" + strconv.Itoa(p)}
}

// hostMCPServerConfig returns a global config defining server "echo", which
// records its pid and working directory into dir and then echoes stdin.
func hostMCPServerConfig(t *testing.T, dir string) string {
	t.Helper()
	script := filepath.Join(dir, "server.sh")
	body := "#!/bin/sh\necho $$ > " + filepath.Join(dir, "pid") + "\npwd > " + filepath.Join(dir, "cwd") + "\necho started >&2\nexec cat\n"
	if err := os.WriteFile(script, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	return "kits:\n  host-mcp:\n    servers:\n      echo:\n        command: [" + script + "]\n"
}

// waitGone waits until the process recorded in dir/pid no longer exists.
func waitGone(t *testing.T, dir string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, "pid"))
	if err != nil {
		t.Fatal(err)
	}
	pid, _ := strconv.Atoi(strings.TrimSpace(string(data)))
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if exec.Command("kill", "-0", strconv.Itoa(pid)).Run() != nil {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Errorf("server process %d still running after the connection closed", pid)
}

func TestHostMCPRelay(t *testing.T) {
	work := t.TempDir()
	home := hostMCPHome(t, hostMCPServerConfig(t, work))
	project := t.TempDir()
	ep := tcpBrokerEndpoint(t)
	startHostMCPBroker(t, ep, project)

	conn, err := net.Dial("tcp", ep.Addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	conn.Write([]byte("GET /mcp/echo HTTP/1.1\r\nHost: x\r\nAuthorization: Bearer tok\r\nConnection: Upgrade\r\nUpgrade: asylum-stdio\r\n\r\n"))
	r := bufio.NewReader(conn)
	resp, err := http.ReadResponse(r, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("status = %d, want 101", resp.StatusCode)
	}

	conn.Write([]byte(`{"jsonrpc":"2.0","id":1}` + "\n"))
	line, err := r.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	if line != `{"jsonrpc":"2.0","id":1}`+"\n" {
		t.Errorf("echoed %q", line)
	}

	cwd, _ := os.ReadFile(filepath.Join(work, "cwd"))
	if got, _ := filepath.EvalSymlinks(strings.TrimSpace(string(cwd))); got != mustEvalSymlinks(t, project) {
		t.Errorf("server cwd = %q, want %q", got, project)
	}

	conn.Close()
	waitGone(t, work)

	log, _ := os.ReadFile(filepath.Join(home, ".asylum", "projects", "test-container", "host-mcp.log"))
	if !strings.Contains(string(log), "started") {
		t.Errorf("server stderr not logged: %q", log)
	}
}

func mustEvalSymlinks(t *testing.T, p string) string {
	t.Helper()
	r, err := filepath.EvalSymlinks(p)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// runShim runs the container-side relay with the given broker environment and
// stdin, and returns its stdout, stderr and error.
func runShim(t *testing.T, env []string, server, stdin string) (string, string, error) {
	t.Helper()
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not available")
	}
	js := filepath.Join(t.TempDir(), "asylum-mcp.js")
	if err := os.WriteFile(js, []byte(hostMCPRelay), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(node, js, server)
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdin = strings.NewReader(stdin)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	return stdout.String(), stderr.String(), err
}

func TestHostMCPShim(t *testing.T) {
	work := t.TempDir()
	hostMCPHome(t, hostMCPServerConfig(t, work))

	tcp := tcpBrokerEndpoint(t)
	startHostMCPBroker(t, tcp, t.TempDir())
	unix := broker.Endpoint{Network: "unix", Addr: filepath.Join(t.TempDir(), "broker.sock")}
	startHostMCPBroker(t, unix, t.TempDir())
	host, port, _ := net.SplitHostPort(tcp.Addr)

	transports := map[string][]string{
		"tcp":  {"ASYLUM_BROKER_SOCK=", "ASYLUM_BROKER_HOST=" + host, "ASYLUM_BROKER_PORT=" + port, "ASYLUM_BROKER_TOKEN=tok"},
		"unix": {"ASYLUM_BROKER_SOCK=" + unix.Addr, "ASYLUM_BROKER_TOKEN=tok"},
	}
	for name, env := range transports {
		t.Run(name, func(t *testing.T) {
			msg := `{"jsonrpc":"2.0","id":1,"method":"initialize"}` + "\n"
			out, stderr, err := runShim(t, env, "echo", msg)
			if err != nil {
				t.Fatalf("shim: %v (%s)", err, stderr)
			}
			if out != msg {
				t.Errorf("stdout = %q, want %q", out, msg)
			}
			waitGone(t, work)
		})
	}

	t.Run("unknown server", func(t *testing.T) {
		_, stderr, err := runShim(t, transports["tcp"], "nope", "")
		if err == nil {
			t.Fatal("shim succeeded for an unknown server")
		}
		if !strings.Contains(stderr, `no host MCP server "nope"`) {
			t.Errorf("stderr = %q", stderr)
		}
	})

	t.Run("wrong token", func(t *testing.T) {
		env := append(slices.Clone(transports["tcp"]), "ASYLUM_BROKER_TOKEN=wrong")
		if _, _, err := runShim(t, env, "echo", ""); err == nil {
			t.Fatal("shim succeeded with a wrong token")
		}
	})
}

func TestHostMCPStopsStubbornServer(t *testing.T) {
	work := t.TempDir()
	script := filepath.Join(work, "server.sh")
	// Ignores stdin and leaves a child behind in its process group.
	body := "#!/bin/sh\ntrap '' TERM\nsleep 60 &\necho $! > " + filepath.Join(work, "pid") + "\necho ready\nwhile :; do sleep 1; done\n"
	if err := os.WriteFile(script, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	hostMCPHome(t, "kits:\n  host-mcp:\n    servers:\n      stubborn:\n        command: ["+script+"]\n")
	ep := tcpBrokerEndpoint(t)
	startHostMCPBroker(t, ep, t.TempDir())

	conn, err := net.Dial("tcp", ep.Addr)
	if err != nil {
		t.Fatal(err)
	}
	conn.Write([]byte("GET /mcp/stubborn HTTP/1.1\r\nHost: x\r\nAuthorization: Bearer tok\r\nConnection: Upgrade\r\nUpgrade: asylum-stdio\r\n\r\n"))
	r := bufio.NewReader(conn)
	if resp, err := http.ReadResponse(r, nil); err != nil || resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("upgrade failed: %v %v", resp, err)
	}
	if line, _ := r.ReadString('\n'); line != "ready\n" {
		t.Fatalf("got %q, want ready", line)
	}
	conn.Close()
	waitGone(t, work)
}

func TestHostMCPServers(t *testing.T) {
	config := `kits:
  host-mcp:
    servers:
      zeta:
        command: [/bin/zeta]
      alpha:
        command: [/bin/alpha]
      pinned:
        command: [/bin/pinned]
        projects: [/code/other]
      relative:
        command: [npx, thing]
`
	active := []*Kit{Get("host-mcp")}
	tests := []struct {
		name   string
		kits   []*Kit
		config string // empty means no config file
		want   []string
	}{
		{"kit inactive", nil, config, nil},
		{"allowed servers sorted", active, config, []string{"alpha", "zeta"}},
		{"missing config", active, "", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config == "" {
				t.Setenv("HOME", t.TempDir())
			} else {
				hostMCPHome(t, tt.config)
			}
			got, err := HostMCPServers(tt.kits, "/code/app")
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("HostMCPServers = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHostMCPServersBadConfig(t *testing.T) {
	hostMCPHome(t, "kits: [not, a, mapping\n")
	if _, err := HostMCPServers([]*Kit{Get("host-mcp")}, "/code/app"); err == nil {
		t.Error("want an error for an unparseable global config")
	}
}
