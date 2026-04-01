## 1. Priority Field and Source Identifiers

- [x] 1.1 Add `DockerPriority int` field to `kit.Kit` struct in `internal/kit/kit.go`
- [x] 1.2 Add `DockerPriority int` field to `agent.AgentInstall` struct in `internal/agent/install.go`
- [x] 1.3 Set `DockerPriority` values on all existing kit registrations (java=10, python=12, node=14, docker=30, misc kits=40-50)
- [x] 1.4 Set `DockerPriority` values on all existing agent install registrations (claude=20, codex=22, gemini=22, opencode=22)

## 2. State Extension

- [x] 2.1 Add `DockerSourceOrder []string` field to `config.State` struct with JSON tag `docker_source_order,omitempty`

## 3. Ordering Algorithm

- [x] 3.1 Create `internal/image/order.go` with the `computeSourceOrder` function that takes active sources (identifier + priority) and previous order, returns computed order
- [x] 3.2 Implement retained-sources preservation: filter previous order to only active sources
- [x] 3.3 Implement removal detection: find earliest removal point and re-sort the suffix by priority
- [x] 3.4 Implement new-source append: sort new sources by priority (alphabetical tie-break) and append
- [x] 3.5 Write unit tests for `computeSourceOrder` in `internal/image/order_test.go` covering: no previous order, no changes, additions, removals, removal+addition, tie-breaking

## 4. Integration into Dockerfile Assembly

- [x] 4.1 Modify `assembleDockerfile` in `internal/image/image.go` to accept the computed source order and assemble snippets in that order instead of kit resolution order
- [x] 4.2 Update `EnsureBase` to call `computeSourceOrder`, pass the result to `assembleDockerfile`, and save the order to state on successful build
- [x] 4.3 Update `baseHash` to use the computed order so that reordering triggers a rebuild
- [x] 4.4 Thread the asylum dir / state through `EnsureBase` (or pass in the previous order and return the new order for the caller to save)
- [x] 4.5 Update `generateProjectDockerfile` to place agent snippets before project kit snippets (no state tracking needed, just static agents-before-kits ordering) — N/A: project images don't install agents (inherited from base), and agent priorities (5) already ensure agents-before-kits in base image

## 5. Verification

- [x] 5.1 Run `go test ./...` and fix any compilation or test failures
- [x] 5.2 Manual smoke test: replaced with automated integration tests (TestOrderingAgentsBeforeKits, TestOrderingClaudeBeforeOtherAgents, TestOrderingNewKitAppendedLast, TestOrderingStateRoundTrip, TestDockerfileSnippetOrder) — full Docker build not possible inside container
