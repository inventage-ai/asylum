# Host MCP Kit

Use stdio MCP servers that can only run on the host, such as [XcodeBuildMCP](https://github.com/cameroncooke/XcodeBuildMCP) or `xcrun mcpbridge`, which need Xcode and the Simulator. The agent starts them from inside the container through the host broker. No ports, no manual proxy.

**Activation: Opt-in**

```yaml
kits:
  host-mcp:
```

## Defining Servers

Servers are defined in `~/.asylum/config.yaml` only. Definitions in `.asylum` or `.asylum.local` are ignored, because the agent can write those files.

```yaml
# ~/.asylum/config.yaml
kits:
  host-mcp:
    servers:
      xcodebuild:
        command: ["/opt/homebrew/bin/xcodebuildmcp", "mcp"]
        env: { XCODEBUILDMCP_SENTRY_DISABLED: "true" }
        projects: ["/Users/me/code/trail-app"]   # optional
      cupertino:
        command: ["/opt/homebrew/bin/cupertino", "serve"]
```

| Key | Meaning |
|---|---|
| `command` | Executable and arguments. The first element must be an absolute path. |
| `env` | Extra environment variables for the server. |
| `projects` | Host project directories that may start the server. Without it, every project with the kit enabled can. |

Edits take effect on the next connection. No container restart is needed.

## Using a Server

Claude sessions that asylum starts get every server the project may use automatically. Asylum passes them to Claude with `--mcp-config` for that session only and writes nothing into Claude's config, so a server you remove from `~/.asylum/config.yaml` is gone from the next session.

Other agents, and `claude` started by hand inside `asylum shell`, need the server registered like any local stdio server:

```bash
claude mcp add xcodebuild -- asylum-mcp xcodebuild
```

If you registered a server by hand before and Claude now lists it twice, remove the manual entry with `claude mcp remove <name> -s local`.

`asylum-mcp <name>` connects to the host broker, which starts the server on the host in the project directory and relays stdin and stdout. Each connection gets its own server process. It is stopped, together with any processes it started, when the agent closes the connection or the container stops.

The server's stderr goes to `~/.asylum/projects/<container>/host-mcp.log` on the host.

## Security

A host MCP server runs with your full rights on the host. Treat each definition like a command you run yourself.

- **Use absolute paths, avoid package runners.** The server runs in the project directory, which the agent controls. Runners such as `npx` or `uvx` pick up `.npmrc`, `pyproject.toml` or `node_modules/.bin` from there. Install the server and point `command` at its binary.
- **Minimal environment.** Servers only get `HOME`, `PATH`, `LANG`, `TMPDIR`, `USER` and `LOGNAME`, plus their `env`. Asylum's container token is never passed on.
- **The token keeps out other containers and local processes, not the agent.** The agent in the container has the token, and can start any server you made available to the project. Use `projects` to narrow that down.
- **The kit controls who can start a server, not what the server does.** Many servers read project-local configuration the agent can change. XcodeBuildMCP, for example, reads `.xcodebuildmcp/config.yaml`, and its debugging tools include LLDB, which can run host commands. Pin settings through `env` where the server supports it, and only expose servers whose tools you are comfortable handing to the agent.
- **Project config is writable by the agent.** Until [#47](https://github.com/inventage-ai/asylum/issues/47) is resolved, a project `.asylum` can add volume mounts that reach `~/.asylum` at the next container start, so the "global config only" rule is not a hard boundary yet.
