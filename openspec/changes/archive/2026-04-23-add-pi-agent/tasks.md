## 1. Create Pi agent implementation

- [x] 1.1 Create `internal/agent/pi.go` with Pi struct implementing the Agent interface
- [x] 1.2 Register Pi in `agents` map and `AgentInstall` in `init()` (npm install via fnm, node kit dependency, DockerPriority 6, banner line)
- [x] 1.3 Implement `Name()`, `Binary()`, config dir methods (`~/.pi`, `~/.asylum/agents/pi`)
- [x] 1.4 Implement `Command()` with `--continue` for resume, extra args passthrough, wrapped in zsh
- [x] 1.5 Implement `HasSession()` checking `~/.pi/agent/sessions/` for project-matching directories (double-dash encoded path pattern)

## 2. Fix active agent not auto-included in install map

- [x] 2.1 Ensure `--agent <name>` always includes that agent in the install map (even when not in config)

## 3. Update specs

- [x] 3.1 Verify `specs/pi-agent/spec.md` covers all pi agent requirements
- [x] 3.2 Verify `specs/agent-interface/spec.md` delta includes pi in registry and ignore list

## 4. Update changelog

- [x] 4.1 Add entry to `CHANGELOG.md` under Unreleased → Added
