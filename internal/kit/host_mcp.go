package kit

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/inventage-ai/asylum/internal/broker"

	"gopkg.in/yaml.v3"
)

// hostMCPServer is a stdio MCP server the user allows containers to start on
// the host.
type hostMCPServer struct {
	Command  []string          `yaml:"command"`
	Env      map[string]string `yaml:"env"`
	Projects []string          `yaml:"projects"`
}

// hostMCPEnvKeys are the broker environment variables a server inherits. The
// broker's own environment holds the full user shell plus the container token,
// none of which a server needs.
var hostMCPEnvKeys = []string{"HOME", "PATH", "LANG", "TMPDIR", "USER", "LOGNAME"}

// hostMCPStopGrace is how long a server gets to exit after SIGTERM.
const hostMCPStopGrace = 2 * time.Second

// hostMCPCommand runs the node relay with whichever node the agent's
// environment provides, falling back to fnm's default alias for agents that
// spawn MCP servers with a reduced PATH.
const hostMCPCommand = `#!/bin/sh
node=$(command -v node || echo "$HOME/.local/share/fnm/aliases/default/bin/node")
exec "$node" /usr/local/lib/asylum-mcp.js "$@"
`

// hostMCPRelay connects to the broker's /mcp/<name> route, upgrades the
// connection and pipes stdio through it.
const hostMCPRelay = `const http = require('http');
const name = process.argv[2];
if (!name) {
  process.stderr.write('usage: asylum-mcp <server>\n');
  process.exit(1);
}
const env = process.env;
const opts = {
  path: '/mcp/' + encodeURIComponent(name),
  headers: {
    Authorization: 'Bearer ' + env.ASYLUM_BROKER_TOKEN,
    Connection: 'Upgrade',
    Upgrade: 'asylum-stdio',
  },
};
if (env.ASYLUM_BROKER_SOCK) {
  opts.socketPath = env.ASYLUM_BROKER_SOCK;
} else {
  opts.host = env.ASYLUM_BROKER_HOST;
  opts.port = env.ASYLUM_BROKER_PORT;
}
const fail = (msg) => {
  process.stderr.write('asylum-mcp: ' + msg + '\n');
  process.exit(1);
};
const req = http.request(opts);
req.on('upgrade', (res, socket, head) => {
  if (head.length) process.stdout.write(head);
  process.stdin.pipe(socket);
  socket.pipe(process.stdout, { end: false });
  socket.on('close', () => process.stdin.destroy());
});
req.on('response', (res) => {
  let body = '';
  res.on('data', (c) => (body += c));
  res.on('end', () => fail(body.trim() || 'broker answered ' + res.statusCode));
});
req.on('error', (e) => fail(e.message));
req.end();
`

func init() {
	Register(&Kit{
		Name:        "host-mcp",
		Description: "Host MCP servers relayed into the container",
		Tier:        TierOptIn,
		Deps:        []string{"node"},
		ConfigSnippet: `  # host-mcp:           # Host MCP servers (define servers in ~/.asylum/config.yaml)
`,
		ConfigNodes:   configNodes("host-mcp", "Host MCP servers (define servers in ~/.asylum/config.yaml)", nil),
		ConfigComment: "host-mcp:             # Host MCP servers (define servers in ~/.asylum/config.yaml)",
		DockerSnippet: hostMCPDockerSnippet(),
		RulesSnippet: `### Host MCP servers (host-mcp kit)
MCP servers the user configured on the host run through the stdio command ` + "`asylum-mcp <name>`" + `. Claude sessions started by asylum already have them registered. Anywhere else, register one like any local stdio server, e.g. ` + "`claude mcp add <name> -- asylum-mcp <name>`" + `. The server runs on the host, in this project's directory, for as long as the MCP connection lasts. An unknown name fails with an error; ask the user which servers are configured.
`,
		Tools:  []string{"asylum-mcp"},
		Routes: []broker.Route{{Path: "/mcp/", Handler: hostMCPHandler}},
	})
}

// hostMCPDockerSnippet stages the relay and its launcher. Both travel as
// base64 because the Dockerfile has no heredoc support.
func hostMCPDockerSnippet() string {
	b64 := func(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }
	return `# Host MCP servers via the asylum host broker.
RUN echo '` + b64(hostMCPRelay) + `' | base64 -d | sudo tee /usr/local/lib/asylum-mcp.js >/dev/null && \
    echo '` + b64(hostMCPCommand) + `' | base64 -d | sudo tee /usr/local/bin/asylum-mcp >/dev/null && \
    sudo chmod +x /usr/local/bin/asylum-mcp
`
}

// loadHostMCPServers reads the server definitions from the global config only.
// Project config files are writable from inside the container, so they must
// never decide what runs on the host. Reading on every request also lets edits
// apply without restarting the container.
func loadHostMCPServers() (map[string]hostMCPServer, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(home, ".asylum", "config.yaml"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var cfg struct {
		Kits struct {
			HostMCP struct {
				Servers map[string]hostMCPServer `yaml:"servers"`
			} `yaml:"host-mcp"`
		} `yaml:"kits"`
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("~/.asylum/config.yaml: %w", err)
	}
	return cfg.Kits.HostMCP.Servers, nil
}

// lookupHostMCPServer returns the named server if projectDir may start it, or
// the HTTP status and reason for refusing.
func lookupHostMCPServer(servers map[string]hostMCPServer, name, projectDir string) (hostMCPServer, int, error) {
	srv, ok := servers[name]
	if !ok {
		return srv, http.StatusNotFound, fmt.Errorf("no host MCP server %q in ~/.asylum/config.yaml", name)
	}
	if len(srv.Command) == 0 || !filepath.IsAbs(srv.Command[0]) {
		return srv, http.StatusForbidden, fmt.Errorf("host MCP server %q: command must start with an absolute path", name)
	}
	if len(srv.Projects) > 0 && !slices.ContainsFunc(srv.Projects, func(p string) bool { return filepath.Clean(p) == filepath.Clean(projectDir) }) {
		return srv, http.StatusForbidden, fmt.Errorf("host MCP server %q is not enabled for %s", name, projectDir)
	}
	return srv, 0, nil
}

// HostMCPServers returns the sorted names of the host MCP servers projectDir
// may start, or nil when the host-mcp kit is not among kits.
func HostMCPServers(kits []*Kit, projectDir string) ([]string, error) {
	if !slices.ContainsFunc(kits, func(k *Kit) bool { return k.Name == "host-mcp" }) {
		return nil, nil
	}
	servers, err := loadHostMCPServers()
	if err != nil {
		return nil, err
	}
	var names []string
	for name := range servers {
		if _, _, err := lookupHostMCPServer(servers, name, projectDir); err == nil {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	return names, nil
}

// hostMCPEnv builds a server's environment from scratch: the allowlisted
// broker variables, overridden by the server's configured env.
func hostMCPEnv(extra map[string]string) []string {
	env := map[string]string{}
	for _, k := range hostMCPEnvKeys {
		if v, ok := os.LookupEnv(k); ok {
			env[k] = v
		}
	}
	for k, v := range extra {
		env[k] = v
	}
	var out []string
	for k, v := range env {
		out = append(out, k+"="+v)
	}
	slices.Sort(out)
	return out
}

// openHostMCPLog opens the per-container server log for appending.
func openHostMCPLog(cname string) (*os.File, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(home, ".asylum", "projects", cname)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	return os.OpenFile(filepath.Join(dir, "host-mcp.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
}

// hostMCPHandler starts a configured MCP server on the host and relays its
// stdio over the upgraded connection. The server is started before the
// upgrade so that failures still reach the shim as HTTP errors, and it lives
// exactly as long as the connection.
func hostMCPHandler(ctx broker.Ctx, w http.ResponseWriter, r *http.Request) {
	servers, err := loadHostMCPServers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	name := strings.TrimPrefix(r.URL.Path, "/mcp/")
	srv, status, err := lookupHostMCPServer(servers, name, ctx.ProjectDir())
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	logFile, err := openHostMCPLog(ctx.Container())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer logFile.Close()
	fmt.Fprintf(logFile, "--- %s start %s\n", time.Now().Format(time.RFC3339), name)

	cmd := exec.Command(srv.Command[0], srv.Command[1:]...)
	cmd.Dir = ctx.ProjectDir()
	cmd.Env = hostMCPEnv(srv.Env)
	cmd.Stderr = logFile
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := cmd.Start(); err != nil {
		http.Error(w, fmt.Sprintf("host MCP server %q: %v", name, err), http.StatusBadGateway)
		return
	}
	defer broker.TrackGroup(cmd.Process.Pid)()
	defer stopHostMCPServer(cmd)

	conn, buf, err := http.NewResponseController(w).Hijack()
	if err != nil {
		return
	}
	defer conn.Close()
	buf.WriteString("HTTP/1.1 101 Switching Protocols\r\nConnection: Upgrade\r\nUpgrade: asylum-stdio\r\n\r\n")
	if err := buf.Flush(); err != nil {
		return
	}

	inputDone := make(chan struct{})
	outputDone := make(chan struct{})
	go func() {
		io.Copy(stdin, buf)
		stdin.Close()
		close(inputDone)
	}()
	go func() {
		io.Copy(conn, stdout)
		close(outputDone)
	}()
	// A client ends an MCP stdio session by closing the server's stdin, so give
	// the server a moment to flush and exit on its own before stopping it.
	select {
	case <-outputDone:
	case <-inputDone:
		select {
		case <-outputDone:
		case <-time.After(hostMCPStopGrace):
		}
	}
}

// stopHostMCPServer terminates the server's process group, escalating to
// SIGKILL when it does not exit in time. The final SIGKILL also reaches
// children that outlive the group leader.
func stopHostMCPServer(cmd *exec.Cmd) {
	pgid := cmd.Process.Pid
	syscall.Kill(-pgid, syscall.SIGTERM)
	exited := make(chan struct{})
	go func() {
		cmd.Wait()
		close(exited)
	}()
	select {
	case <-exited:
	case <-time.After(hostMCPStopGrace):
	}
	syscall.Kill(-pgid, syscall.SIGKILL)
	<-exited
}
