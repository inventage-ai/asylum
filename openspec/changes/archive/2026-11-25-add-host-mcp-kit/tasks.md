# Tasks

## 1. Broker plumbing

- [x] 1.1 Pass the project directory from `ensureBroker` to `__broker` (`--project`) and expose it as `Ctx.ProjectDir()`; verify with a broker test that a handler sees the directory
- [x] 1.2 Add a broker-wide registry of spawned process groups that the `docker wait` exit path kills before `os.Exit`; verify with a unit test that registered groups are signalled on shutdown
- [x] 1.3 Update the `host-broker` main spec references in `docs/` (if the broker is documented) for request-time global config and project directory; verify the docs build

## 2. host-mcp kit (host side)

- [x] 2.1 Read `kits.host-mcp.servers` from `~/.asylum/config.yaml` with a minimal yaml struct, without touching project layers; verify with a table-driven test (missing file, no kit, servers present, project file ignored)
- [x] 2.2 Validate a server request: unknown name → 404, non-absolute `command[0]` → 403, project not in `projects` → 403; verify with table-driven handler tests using `httptest`
- [x] 2.3 Build the minimal server environment (`HOME`, `PATH`, `LANG`, `TMPDIR`, `USER`, `LOGNAME` + `env`); verify with a test that `ASYLUM_BROKER_TOKEN` is absent and configured `env` overrides
- [x] 2.4 Spawn the server (`Setpgid`, cwd = project dir, stderr appended to `~/.asylum/projects/<container>/host-mcp.log`), hijack, answer `101`, relay stdio, kill the group on close; verify with a test that uses `cat` as the server: bytes round-trip, and the process is gone after the client closes
- [x] 2.5 Register the kit (`TierOptIn`, `Deps: ["node"]`, `Routes` with `/mcp/`); verify `go test ./internal/kit/...` and that `asylum config` lists the kit as opt-in

## 3. Container shim and rules

- [x] 3.1 Add the Node `asylum-mcp` shim to the kit's `DockerSnippet` (Unix socket and TCP, error body to stderr, exit 1 on non-101); verify with a snippet test that `/usr/local/bin/asylum-mcp` is installed
- [x] 3.2 Add the `RulesSnippet` explaining `claude mcp add <name> -- asylum-mcp <name>` (and the generic stdio form for other agents); verify with a snippet test
- [x] 3.3 Check that `node` is on `PATH` for MCP servers spawned by Claude in a container, with the fnm default-alias fallback in the shim; verified manually with Claude (Codex and Gemini deferred by decision)

## 4. Config and docs

- [x] 4.1 Verify that kit sync, migration and `asylum config` keep `kits.host-mcp.servers` intact when they rewrite `~/.asylum/config.yaml`; fix and add a round-trip test if they don't
- [x] 4.2 Write `docs/kits/host-mcp.md`:
  - configuration example (XcodeBuildMCP)
  - why `command` must be absolute and why `npx`/`uvx` are risky
  - what the token does and does not protect
  - that a server can be steered by the agent (project-local server config, LLDB)
  - reference to #47

  Add the page to `docs/kits/index.md` and verify the docs build.
- [x] 4.3 Add a CHANGELOG entry under **Unreleased → Added**; verify it is present

## 5. Integration check

- [x] 5.1 End-to-end on macOS (TCP transport): configure XcodeBuildMCP globally, register it with `claude mcp add xcodebuild -- asylum-mcp xcodebuild` and run a tool call; verify the server stops when the session ends and that a definition in the project `.asylum` is ignored
- [x] 5.2 Run `go vet ./...` and `go test ./...`, and verify both pass

## 6. Automatic registration for Claude sessions

- [x] 6.1 Add `kit.HostMCPServers(kits, projectDir)` that returns the sorted names of the servers the project may start, or nothing when the kit is inactive; verify with a table-driven test (kit inactive, allowlist, invalid command skipped, missing config)
- [x] 6.2 Add the server list to `agent.CmdOpts` and emit `--mcp-config` with stdio `asylum-mcp <name>` entries in `Claude.Command`, placed before `--add-dir`; verify with `TestClaudeCommand` cases (JSON content, ordering with passthrough args and `--add-dir`) and that other agents ignore it
- [x] 6.3 Fill the list in `container.agentCommand` and warn when the global config cannot be read; verify with an `ExecArgs` test
- [x] 6.4 Update the rules snippet, `docs/kits/host-mcp.md` and the CHANGELOG entry (auto-registration, manual registration for other agents and `asylum shell`, removing duplicate manual entries); verify the docs build
- [x] 6.5 Verify manually on macOS that a Claude session started with `asylum-dev` lists `xcodebuild` under `/mcp` without a manual registration

## Workflow follow-up

- Archive the change after review with `/opsx:archive`.
