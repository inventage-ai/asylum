## 1. Remove agents from priority ordering

- [x] 1.1 Remove the `DockerPriority` field from `AgentInstall` in `internal/agent/install.go` and update its doc comment
- [x] 1.2 Remove `DockerPriority:` lines from all agent registrations (`claude.go`, `codex.go`, `gemini.go`, `copilot.go`, `opencode.go`, `pi.go`)
- [x] 1.3 In `internal/image/order.go`, change `collectSources` to take kits only (drop the agents/versions parameters) and emit only `kit:*` sources

## 2. Assemble agents as the top layer

- [x] 2.1 In `internal/image/image.go`, change `assembleDockerfile` to write ordered kit snippets, then a versioned agent snippet block (via `agent.AssembleVersionedAgentSnippets`), then the tail
- [x] 2.2 Update `EnsureBase` to pass agents/versions to `assembleDockerfile` (not `collectSources`) and keep saving a kit-only `docker_source_order`
- [x] 2.3 Update `baseHash` to hash the assembled agent block (and its position) so an agent version bump still changes the hash

## 3. Remove vestigial project-image rule

- [x] 3.1 Confirm `generateProjectDockerfile` takes no agents and remove any agents-before-kits handling/comments referencing it

## 4. Tests

- [x] 4.1 Replace `TestOrderingAgentsBeforeKits` with a test asserting agents are emitted after all kit snippets and before the tail
- [x] 4.2 Remove `TestOrderingClaudeBeforeOtherAgents` (or repoint it at the agent-block deterministic order)
- [x] 4.3 Update `TestDockerfileSnippetOrder` so it asserts `javaIdx < claudeIdx` (kit before agent)
- [x] 4.4 Update `collectSources` call sites in `order_test.go` / `assembly_test.go` for the new signature
- [x] 4.5 Add a test that an agent-only version change leaves kit snippet bytes/positions unchanged in the assembled Dockerfile

## 5. Verify

- [x] 5.1 `go test ./...` and `go vet ./...` pass
- [x] 5.2 Manually assemble a base Dockerfile (or run a build) and confirm kit RUNs precede agent RUNs, with agents immediately before the tail
- [x] 5.3 Add a CHANGELOG entry under Unreleased (Changed): agents now build as the top base-image layer to avoid rebuilding kit layers on version bumps
