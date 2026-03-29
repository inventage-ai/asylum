## 1. Dockerfile Changes

- [x] 1.1 Add `ARG USER_HOME=/home/claude` to Dockerfile.core; change `useradd` to use `-d ${USER_HOME}`; change `WORKDIR /home/${USERNAME}` to `WORKDIR ${USER_HOME}`
- [x] 1.2 Update `image.EnsureBase` to pass `USERNAME` from `user.Current().Username` and `USER_HOME` from `user.Current().HomeDir` (replace hardcoded `"claude"`)
- [x] 1.3 Verify base image hash includes USER_HOME (already included via build args hash)

## 2. Entrypoint Cleanup

- [x] 2.1 Replace all `/home/claude` references in `assets/entrypoint.core` with `$HOME` (SSH permissions, direnv allow, gitconfig)
- [x] 2.2 Verify entrypoint still works with `$HOME` — all references are in conditional blocks that naturally use the shell's HOME

## 3. Kit DockerSnippets

- [x] 3.1 Replace `USER claude` with `USER ${USERNAME}` in all kit DockerSnippets: java/maven, docker, python, github
- [x] 3.2 Replace `usermod -aG docker claude` with `usermod -aG docker ${USERNAME}` in docker kit

## 4. Kit CacheDirs

- [x] 4.1 Change all kit CacheDirs from absolute paths to tilde-prefixed: `/home/claude/.m2` → `~/.m2`, `/home/claude/.npm` → `~/.npm`, etc.
- [x] 4.2 Update `container.go` volume mounting to resolve `~/` in CacheDirs using `config.ExpandTilde` with the host home dir

## 5. Agent ContainerConfigDir

- [x] 5.1 Change all agent `ContainerConfigDir()` methods to return home-relative paths resolved from `os.UserHomeDir()`: Claude → `$HOME/.claude`, Gemini → `$HOME/.gemini`, Codex → `$HOME/.codex`, Opencode → `$HOME/.opencode`, Echo → keep `/tmp/asylum-echo`
- [x] 5.2 Update agent `EnvVars()` methods to use dynamic home dir: `CLAUDE_CONFIG_DIR`, `CODEX_HOME`

## 6. Container.go Path Resolution

- [x] 6.1 Replace `/home/claude/.ssh` with dynamic path in SSH volume mount
- [x] 6.2 Replace `/home/claude/.shell_history` with dynamic path in history volume mount
- [x] 6.3 Replace hardcoded HISTFILE env var with dynamic path
- [x] 6.4 Add `containerHome()` helper that returns the host home dir (used throughout container.go)

## 7. Dockerfile.tail Cleanup

- [x] 7.1 Replace `${USERNAME:-claude}` with `${USERNAME}` in Dockerfile.tail (the default is no longer needed since USERNAME is always passed)

## 8. Sandbox Rules and Remaining References

- [x] 8.1 Update sandbox rules template in container.go: `User: claude` → `User: <dynamic>` or remove hardcoded username
- [x] 8.2 Grep for any remaining `claude` references in Go code and assets that should be dynamic (excluding agent names like `"claude"` in the agent registry which refer to the Claude product, not the username)

## 9. Testing

- [x] 9.1 Update container tests that reference `/home/claude` or use the old hardcoded paths
- [x] 9.2 Write test: CacheDirs with tilde prefix are resolved correctly
- [x] 9.3 Write test: agent ContainerConfigDir returns path under host home dir
- [x] 9.4 Verify all existing tests pass
- [x] 9.5 Add CHANGELOG entry
