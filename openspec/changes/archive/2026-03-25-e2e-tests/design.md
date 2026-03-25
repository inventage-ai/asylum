## Context

Asylum has unit tests for each package and integration tests that run `docker run` against the built image. But nothing exercises the actual binary end-to-end. The recent config restructure (kits, agents, migration) added significant wiring that's only tested at the unit level.

The kit system now makes fast e2e tests viable: a minimal config with `kits: {}` and `agents: {}` builds an image with just the core OS tools (~1GB instead of ~5.6GB), and a dummy agent avoids needing real credentials.

## Goals / Non-Goals

**Goals:**
- E2e tests that compile and run the `asylum` binary
- Minimal image builds (core only) for fast test cycles
- Verify: help output, image build, container start, agent exec, shell mode, run mode, container cleanup
- Dummy `echo` agent for testing the exec path without real agent CLIs

**Non-Goals:**
- Testing real agent CLIs (Claude, Gemini, etc.) — those need credentials
- Replacing existing integration tests (those test image content, these test the binary)
- Testing on multiple OS/arch (e2e tests run on the current platform only)

## Decisions

### 1. Echo agent — a shell command, not a real CLI

The `echo` agent implements the `Agent` interface but its `Command()` just runs `echo` with the provided args. It has no native config, no session detection, no resume. It's registered in the agent registry so `asylum -a echo` works.

```go
type Echo struct{}
func (Echo) Name() string { return "echo" }
func (Echo) Command(resume bool, extraArgs []string) []string {
    return []string{"echo", strings.Join(extraArgs, " ")}
}
```

No `AgentInstall` is registered — echo doesn't need anything installed in the image.

### 2. E2e tests in `e2e/` directory with `e2e` build tag

Separate from `integration/` because they have different concerns:
- `integration/`: tests the Docker image content (tools installed, entrypoint behavior)
- `e2e/`: tests the asylum binary (config loading, flag parsing, lifecycle)

Build tag `e2e` keeps them out of `go test ./...`. Run via `make test-e2e`.

### 3. Test setup: build binary + create temp project

Each test suite builds `asylum` once into a temp directory, creates a temp project directory with a minimal `.asylum` config, and sets `HOME` to a temp dir (so default config doesn't interfere). Tests call the binary via `exec.Command`.

### 4. Minimal config for fast builds

Test config:
```yaml
version: "0.2"
agent: echo
kits: {}
agents: {}
```

No kits, no agents installed → fastest possible image build. The echo agent doesn't need anything in the image since it just runs the shell `echo` command which is always available.

### 5. Test cases

- **Basic**: `asylum --help` exits 0 with usage text, `asylum --version` shows version
- **Image build**: `asylum run echo ok` builds the image and runs successfully (first run)
- **Container lifecycle**: after `asylum run echo ok`, container is created and cleaned up
- **Shell mode**: `asylum shell` with stdin piped starts and exits cleanly
- **Run mode**: `asylum run ls` outputs directory listing
- **Agent mode**: `asylum -- hello world` with echo agent outputs "hello world"
- **Config flags**: `--kits java` triggers a rebuild, `--agents claude` works

## Risks / Trade-offs

**E2e tests are slow** → Mitigated by minimal config. First image build is unavoidable (~30s for core only), but subsequent tests reuse the cached image.

**Echo agent is in production binary** → Acceptable. It's a tiny struct with no dependencies. Could be gated behind a build tag if needed later, but the overhead is negligible.
