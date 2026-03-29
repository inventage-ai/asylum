# Sessions

Asylum supports multiple concurrent sessions sharing a single container, with automatic session resume for agents.

## Multi-Session

Multiple terminals can share the same container:

1. The first `asylum` invocation starts the container
2. Subsequent invocations exec into the running container
3. The container is automatically removed when the last session exits

Session count is tracked in `~/.asylum/projects/<container>/sessions` using file locking.

## Agent Resume

Agents automatically resume their previous session by default. This means if you run `asylum`, close the terminal, and run `asylum` again, the agent picks up where it left off.

Use `-n` / `--new` to start a fresh session:

```sh
asylum -n
```

### How Resume Detection Works

Each agent has its own method for detecting previous sessions:

- **Claude Code**: Looks for `.jsonl` session files in `~/.asylum/agents/claude/projects/<encoded-path>/`
- **Gemini CLI**: Checks for a `.project_root` file and chat history in `~/.asylum/agents/gemini/tmp/`
- **Codex**: Uses a marker file at `~/.asylum/agents/codex/projects/<encoded-path>/.has_session`

### First Run Exception

On the very first run (when agent config is seeded from the host), resume is skipped. The seeded data doesn't represent a container session, so the agent starts fresh.

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
