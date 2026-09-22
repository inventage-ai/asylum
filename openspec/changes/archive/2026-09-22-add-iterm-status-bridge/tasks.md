## 1. Verify blocking assumptions before building

- [x] 1.1 From an iTerm2 tab, note `$ITERM_SESSION_ID`, switch to a second tab, and run the host `cc-status` from a subshell that exits immediately so the job reparents to launchd, passing the first tab's id and a synthetic `PreToolUse` payload on stdin. Confirm the first tab's status bar updates. A permission dialog or no update means the iTerm2 API rejects a reparented caller, the handler cannot live in the detached broker, and this change must be re-proposed around a per-session helper
- [x] 1.2 Confirm a Claude Code hook subprocess inherits the `docker exec` environment: start a session with `docker exec -e ASYLUM_PROBE=1`, point a temporary hook at a script that writes `$ASYLUM_PROBE` to a file, and check the value arrives
- [x] 1.3 Record where `it2` lives on the host and whether it is on `PATH` outside an iTerm2 shell (`command -v it2`), and which variables iTerm2 exports into a session environment (`env | grep -i iterm`). The answers fix the `PATH` the route hands the child and the fields the envelope carries

## 2. Session envelope in the broker package

- [x] 2.1 Add envelope seal/open to `internal/broker`: load-or-create a 32-byte key at `~/.asylum/session.key` with mode `0600`, seal a JSON payload with AES-256-GCM under a random nonce, and return it base64url-encoded
- [x] 2.2 Open rejects a short, malformed, or tampered blob, and returns a distinguishable "no envelope" case for an absent one
- [x] 2.3 Table-driven tests: round-trip, tampered ciphertext rejected, wrong key rejected, absent envelope distinguished from invalid, key file created once and reused

## 3. Per-session environment at exec

- [x] 3.1 Add a per-session environment field to `container.ExecOpts` and emit `-e KEY=value` from `ExecArgs` (`internal/container/container.go:781`) ahead of the container name, for all three exec modes
- [x] 3.2 In `cmd/asylum/main.go`, read `TERM_SESSION_ID` from asylum's own environment (gated on `ITERM_SESSION_ID` marking the terminal as iTerm2) when the `iterm` kit is active, seal it (with the credential fields task 1.3 identified) and pass the sealed value as a per-session variable. Omit the variable entirely when the host value is absent
- [x] 3.3 Tests in `internal/container` asserting the flag is emitted for shell, command, and agent modes, and omitted when no per-session values are supplied

## 4. The iterm kit

- [x] 4.1 Add `internal/kit/iterm.go` registering an `iterm` kit at `TierOptIn`, with a config snippet and comment matching the conventions in `internal/kit/browser_open.go`
- [x] 4.2 `MountFunc` writes the shim `0755` into the per-container staging directory and returns it as a `CredentialMount.HostPath` destined for the container-side hook path. Add a comment explaining why the generated-content path is not used (`Content` is staged `0600`)
- [x] 4.3 The shim: exit `0` immediately when the envelope variable is absent; otherwise POST stdin to `/iterm-status` with the broker bearer token and the envelope header, over the Unix socket when `ASYLUM_BROKER_SOCK` is set and loopback TCP otherwise; bound with `--max-time`; discard all output; exit `0` unconditionally
- [x] 4.4 `RulesSnippetFunc` describing the integration in one short paragraph for the in-container agent
- [x] 4.5 Kit tests covering registration, tier, the mount destination, and that the shim is written executable

## 5. The broker route

- [x] 5.1 Register a `/iterm-status` route on the kit. The handler opens the envelope from the request header, rejects with `400` when it is absent or unopenable, caps the request body, and runs the host `cc-status` to completion with the body on stdin
- [x] 5.2 Resolve `cc-status` from a fixed host path and respond `503` without running anything when it is missing. Prepend iTerm2's utilities directory to the child's `PATH` so `cc-status` finds `it2` regardless of what the broker inherited
- [x] 5.3 Serialize execution per session id so consecutive events cannot be applied out of order
- [x] 5.4 Handler tests: missing envelope rejected, tampered envelope rejected, oversized body rejected, argv fixed regardless of body contents, host binary absent yields `503`

## 5b. Fix: accepting an opt-in kit must enable it

- [x] 5b.1 Move `snippetIsActive`/`uncommentSnippet`/`commentSnippet` from `internal/firstrun/config_writer.go` into `internal/kit/snippet.go` as exported helpers, and rewire the first-run writer to call them
- [x] 5b.2 In `internal/config/kitsync.go`, normalise an accepted kit's `ConfigSnippet` to active form before writing it
- [x] 5b.3 Tests in `internal/kit` for the three helpers, plus a registry-wide test asserting every `TierOptIn` kit's snippet becomes active when uncommented
- [x] 5b.4 **Fixed** entry in `CHANGELOG.md`

## 5c. Verification-pass test gaps

- [x] 5c.1 Test that the handler serializes updates per terminal session, confirmed to fail when the lock is removed
- [x] 5c.2 Test the shim end to end against a stub broker: forwards stdin and the envelope header, sends nothing without an envelope, exits `0` when the broker is unreachable
- [x] 5c.3 Tests that an accepted opt-in kit lands in the config active and a declined one stays commented, confirmed to fail against the pre-fix `kitsync`
- [x] 5c.4 Test that the envelope key sits directly in `~/.asylum` rather than the per-container directory that is bind-mounted into the container

## 6. Documentation

- [x] 6.1 Add an iTerm2 status integration section to `assets/asylum-reference.md` stating what the kit does, that it is opt-in, and that it is inert outside iTerm2
- [x] 6.2 Add a docs page under `docs/` covering enabling the kit, the trust boundary (why `cc-status` and not `it2`), and the failure modes
- [x] 6.3 Add an **Added** entry to the Unreleased section of `CHANGELOG.md`

## 7. Verification

- [x] 7.1 `go test ./...` and `go vet ./...`
- [x] 7.2 With the kit enabled, start a session in an iTerm2 tab and confirm the status dot turns orange during tool calls and green at rest, with the detail text naming the running tool
- [x] 7.3 Open a second tab on the same project and confirm each tab's status bar tracks its own session while both are attached, and that closing the first leaves the second working
- [x] 7.4 Confirm no hook errors appear in a session with the kit enabled, and that the ten iTerm2 hooks no longer fail
- [x] 7.5 Confirm the feature is inert with the kit disabled, and inert with the kit enabled but the session started from a non-iTerm2 terminal
- [x] 7.6 Confirm a session survives a killed broker: stop the broker mid-session and check that tool calls continue with only the status bar going stale
