## 1. Container naming

- [x] 1.1 Add `SecondaryContainerName(projectDir, agents []string)` returning `asylum-<sha256(project+sorted_agents)[:6]>-<basename>`. *(Deviation: kept `ContainerName(projectDir)` unchanged and added a separate function instead of changing its signature — guarantees the primary name stays byte-identical and avoids churn across ~15 callers/tests.)*
- [x] 1.2 `RunOpts` gains `ContainerName` (resolved name; falls back to the primary when empty) and `Secondary bool`; `RunArgs` uses the passed name.

## 2. Agent support label & helpers

- [x] 2.1 `asylum.agents` label emission kept in `RunArgs` — value = sorted baked agent set.
- [x] 2.2 `InspectLabels` kept in `internal/docker/docker.go`. *(Deviation: removed `ContainerHasAgent` — it would be unused in production once the legacy-aware check lives at the call site; removed its and `InspectLabels`'s weak unit tests, which re-implemented stdlib logic rather than calling the functions.)*
- [x] 2.3 Legacy "no label" case resolves to "supports default agent (`claude`) only" in `containerSupportsAgent` at the call site in `main.go`.

## 3. Two-pass resolution in main

- [x] 3.1 Pass 1: resolve the primary name; if running and supports the active agent, proceed as today.
- [x] 3.2 Pass 2: otherwise resolve the secondary name from `project + baked agent set` and proceed; the existing `!docker.IsRunning(cname)` branch builds/starts it if absent.
- [x] 3.3 Removed the prototype branch that called `RemoveContainer` when the agent was missing.

## 4. Ports for secondaries

- [x] 4.1 `ports.Allocate` skipped for secondaries — `portsContainerFunc` returns early on `opts.Secondary`.
- [x] 4.2 `RunArgs` omits user-config `-p` args when `opts.Secondary`.
- [x] 4.3 `internal/ports` unchanged.

## 5. Verification

- [x] 5.1 `go build ./...`, `go vet ./...`, `go test ./...` pass (659 tests); added `TestSecondaryContainerName` covering distinctness, order-independence, and project suffix.
- [x] 5.2 Manual (Docker): `asylum -a claude` then `asylum -a pi` → two containers running, primary keeps ports, pi container has none; `asylum -a pi` again reuses the secondary. *(Verified.)*
- [x] 5.3 Manual (Docker): adding `pi` to config bakes it into the primary; `asylum -a pi` then reuses the primary with ports and creates no secondary. *(Verified.)*
- [x] 5.4 Remove the `.pi/` scratch docs.
