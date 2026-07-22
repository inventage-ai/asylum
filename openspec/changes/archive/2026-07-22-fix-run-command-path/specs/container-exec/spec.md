## MODIFIED Requirements

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
