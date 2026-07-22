## ADDED Requirements

### Requirement: Detect running container
The docker package SHALL provide a function to check if a container with a given name is currently running.

#### Scenario: Container is running
- **WHEN** a container named `asylum-<hash>` is running
- **THEN** `IsRunning("asylum-<hash>")` returns `true`

#### Scenario: Container is not running
- **WHEN** no container with that name exists
- **THEN** `IsRunning("asylum-<hash>")` returns `false`

#### Scenario: Container exists but is stopped
- **WHEN** a container with that name exists but is in exited/dead state
- **THEN** `IsRunning("asylum-<hash>")` returns `false`

### Requirement: Exec into running container for shell mode
When a container is already running for the current project and the user runs `asylum shell`, asylum SHALL exec into the running container instead of starting a new one.

#### Scenario: Shell with running container
- **WHEN** the user runs `asylum shell` and a container is running for the project
- **THEN** asylum runs `docker exec -it <container-name> /bin/zsh`

#### Scenario: Admin shell with running container
- **WHEN** the user runs `asylum shell --admin` and a container is running for the project
- **THEN** asylum runs `docker exec -it -u root <container-name> /bin/zsh`

### Requirement: Exec into running container for run mode
When a container is already running and the user runs `asylum run <cmd>`, asylum SHALL exec the command in the running container through a login shell so that the command resolves tools exactly as an interactive `asylum shell` session would. asylum SHALL source `~/.zshrc` (which sets up `~/.local/bin`, fnm, and mise on `PATH`) and then `exec` the command, replacing the shell so the command's exit status and signals pass through unchanged. The command's arguments SHALL be shell-quoted before being passed to the shell so argument boundaries and metacharacters are preserved.

#### Scenario: Run command resolves a PATH-managed tool
- **WHEN** the user runs `asylum run claude auth login` and a container is running
- **THEN** asylum runs `docker exec -it <container-name> zsh -c "source ~/.zshrc && exec 'claude' 'auth' 'login'"` (each argument shell-quoted) and the `claude` binary in `~/.local/bin` is found and executed

#### Scenario: Run command resolves an fnm/mise-managed tool
- **WHEN** the user runs `asylum run node --version` (or another fnm- or mise-managed tool) and a container is running
- **THEN** the tool is resolved via the environment set up by `~/.zshrc` and executes successfully rather than failing with "executable file not found in $PATH"

#### Scenario: Arguments with spaces are preserved
- **WHEN** the user runs `asylum run node -e "console.log('a b')"`
- **THEN** the argument boundaries are preserved through shell-quoting so the command receives the same argv it would in a shell

#### Scenario: Exit status passes through
- **WHEN** a command run via `asylum run` exits with a non-zero status
- **THEN** because the shell `exec`s the command, asylum observes the command's own exit status rather than a wrapper's

#### Scenario: Command on the default PATH still works
- **WHEN** the user runs `asylum run ls -la` and a container is running
- **THEN** the command executes successfully as before

### Requirement: Skip image build when exec-ing
When asylum detects it will exec into a running container, it SHALL skip the image build step.

#### Scenario: No image build on exec
- **WHEN** a container is running and any mode is used
- **THEN** `EnsureBase` and `EnsureProject` are not called

### Requirement: Detached container lifecycle
When no container is running, asylum SHALL start the container in detached mode with an idle process, then exec the session into it.

#### Scenario: First invocation starts detached container
- **WHEN** no container is running and the user runs `asylum`
- **THEN** the container is started detached with an idle process, then the agent is exec'd into it

#### Scenario: First invocation still builds images
- **WHEN** no container is running
- **THEN** `EnsureBase` and `EnsureProject` are called before starting the container

### Requirement: Exec agent into running container
When a container is already running for the current project and the user runs `asylum` (agent mode), asylum SHALL exec the agent into the running container. By default, asylum SHALL start a new agent session — it SHALL NOT auto-resume from local session markers. Resume happens only when the user explicitly passes `--continue` or `--resume` (which asylum forwards to the agent), or when the resolved config has `default-resume: true`.

#### Scenario: Agent exec with running container
- **WHEN** the user runs `asylum` and a container is running for the project
- **THEN** asylum execs the agent command into the running container via `docker exec -it`

#### Scenario: Default starts a new session
- **WHEN** the user runs `asylum` with no resume-related flags and `default-resume` is unset (or `false`)
- **THEN** the exec'd agent starts a fresh session — no `--continue`/`--resume` is injected, regardless of whether a local session marker exists

#### Scenario: --continue passthrough
- **WHEN** the user runs `asylum --continue`
- **THEN** `--continue` is included verbatim in the agent's argv, and asylum does NOT additionally inject its own resume flag

#### Scenario: --resume passthrough
- **WHEN** the user runs `asylum --resume`
- **THEN** `--resume` is included verbatim in the agent's argv

#### Scenario: default-resume restores previous behaviour
- **WHEN** the resolved config has `default-resume: true` and a local session marker indicates a prior session exists
- **THEN** asylum injects the agent's native resume flag (e.g. `--continue` for Claude, `--resume` for Gemini/Copilot, `resume --last` for Codex) as it did before this change

#### Scenario: default-resume with no prior session
- **WHEN** `default-resume: true` is set but `HasSession` returns false
- **THEN** asylum starts a new session — there is nothing to resume

#### Scenario: -n/--new no-op with default-resume on
- **WHEN** `default-resume: true` is set AND the user runs `asylum -n`
- **THEN** `-n` is ignored (no-op) and asylum still resumes per `default-resume`

### Requirement: Container cleanup after last session
After any exec'd session exits, asylum SHALL check if other sessions remain in the container by querying active exec processes and remove the container if none do.

#### Scenario: Last session exits
- **WHEN** the last exec'd session in a container exits
- **THEN** asylum runs a process check inside the container, finds no other exec sessions, and removes the container

#### Scenario: Other sessions still running
- **WHEN** an exec'd session exits but other sessions are still running
- **THEN** asylum runs a process check inside the container, detects the other sessions, and leaves the container running

#### Scenario: Container already dead at cleanup
- **WHEN** an exec'd session exits and the container has already stopped
- **THEN** asylum treats it as "no sessions" and attempts removal (which is a no-op)

#### Scenario: Background tasks do not prevent cleanup
- **WHEN** the last exec'd session exits but user-started background processes (e.g., dev servers) are still running inside the container
- **THEN** asylum removes the container because background tasks are not exec sessions

### Requirement: Runtime exec session detection
The docker package SHALL provide a function to detect whether other exec sessions are active in a container by running `ps` inside the container and checking for processes with PPID=0 (excluding PID 1 and the check process itself).

#### Scenario: No other sessions
- **WHEN** `HasOtherSessions` is called and only the check process has PPID=0
- **THEN** it returns `false`

#### Scenario: Other sessions active
- **WHEN** `HasOtherSessions` is called and other exec sessions have PPID=0
- **THEN** it returns `true`

#### Scenario: Container not reachable
- **WHEN** `HasOtherSessions` is called but the container is stopped or missing
- **THEN** it returns `false`

### Requirement: SIGHUP signal forwarding
The CLI SHALL forward SIGHUP (in addition to SIGINT and SIGTERM) to the docker exec process so agents receive clean shutdown signals on terminal close.

#### Scenario: Terminal closes during session
- **WHEN** the terminal is closed while an agent session is running
- **THEN** SIGHUP is forwarded to the docker exec process

### Requirement: Independent session exit
Each asylum session SHALL be able to exit independently without affecting other running sessions in the same container.

#### Scenario: First session exits, others continue
- **WHEN** the first `asylum` session exits and a second session is still running
- **THEN** the second session continues running, the container stays alive
