# Sessions

Asylum supports multiple concurrent sessions sharing a single container, with automatic session resume for agents.

## Multi-Session

Multiple terminals can share the same container:

1. The first `asylum` invocation starts the container
2. Subsequent invocations exec into the running container
3. The container is automatically removed when the last session exits

When a session exits, asylum checks for other active exec sessions inside the container. If none remain, the container is automatically removed.

## Agent Resume

Each `asylum` invocation starts a new agent session by default. To resume the previous session, pass `--continue` or `--resume` — both are forwarded verbatim to the underlying agent (which has its own resume picker / behaviour):

```sh
asylum --continue
asylum --resume
```

To restore the pre-0.7 behaviour of automatically resuming whenever a prior session exists, set:

```yaml
# ~/.asylum/config.yaml
default-resume: true
```

The flag is layered like every other config value — set it globally, override per project in `.asylum`, or override per checkout in `.asylum.local`.

`-n` / `--new` is a deprecated no-op kept so existing scripts continue to parse. Starting a new session is now the default.

### Upgrade Dialog

On the first `asylum` invocation after upgrading from an earlier release, asylum shows a one-time dialog explaining this change and offering to set `default-resume: true` for you. The dialog is shown once per installation and is skipped entirely for fresh installs.

### How Resume Detection Works

When `default-resume: true` is on, each agent has its own method for detecting previous sessions:

- **Claude Code**: Looks for `.jsonl` session files in `~/.asylum/agents/claude/projects/<encoded-path>/`
- **Gemini CLI**: Checks for a `.project_root` file and chat history in `~/.asylum/agents/gemini/tmp/`
- **Codex**: Uses a marker file at `~/.asylum/agents/codex/projects/<encoded-path>/.has_session`

### First Run Exception

On the very first run (when agent config is seeded from the host), resume is skipped even with `default-resume: true`. The seeded data doesn't represent a container session, so the agent starts fresh.

## Container Naming

Container names are deterministic based on the project directory:

```
asylum-<sha256(project_dir)[:6]>
```

This means the same project directory always gets the same container name, enabling session persistence and resume detection.

## Container Hostname

The container hostname is derived from the project directory basename:

```
asylum-<safe-lowercase-basename>
```

Special characters are replaced with hyphens, and the name is capped at 63 characters total.
