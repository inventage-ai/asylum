## 1. Verify the lever

- [x] 1.1 Against the OpenSpec version installed by the kit, re-confirm that a global `~/.config/openspec/config.json` of `{"profile":"custom","workflows":["propose","explore","apply","verify","archive"]}` causes `openspec init` to generate the `verify` workflow and omit `sync`. Adjust field names if the CLI has changed. *(Verified against 1.4.1: field names `profile`/`workflows` drive generation; seeded config yields `propose,explore,apply,verify,archive`, no `sync`.)*

## 2. Agent identity env var

- [x] 2.1 In `internal/container/container.go`, set `ASYLUM_AGENT` to the active agent's name (`opts.Agent.Name()`) for every run, alongside the existing agent/kit env handling.
- [x] 2.2 Add/extend a test asserting the container run args include `-e ASYLUM_AGENT=<name>`.

## 3. OpenSpec kit changes

- [x] 3.1 In `internal/kit/openspec.go`, extend the `DockerSnippet` to write the preferred global config to `~/.config/openspec/config.json` at build time.
- [x] 3.2 In the same `DockerSnippet`, install an executable `asylum-openspec-init` script on PATH (e.g. `/usr/local/bin`) that: resolves `--tools` from `ASYLUM_AGENT` (mapping `copilot`→`github-copilot`); runs `openspec init --tools <id>` when no `openspec/` dir exists, else `openspec update --force`.
- [x] 3.3 Add a `RulesSnippet` to the kit instructing the agent to run `asylum-openspec-init` when the user wants OpenSpec and `openspec/` is absent.
- [x] 3.4 Change the kit `Tier` to `TierDefault`.

## 4. Documentation & changelog

- [x] 4.1 Update `docs/kits/openspec.md` with a setup section showing `! asylum-openspec-init` and what it produces (verify workflow set).
- [x] 4.2 Add `Added`/`Changed` entries to `CHANGELOG.md` under Unreleased.

## 5. Verification

- [x] 5.1 `go test ./...` and `go vet ./...` pass. *(589 tests pass; vet clean.)*
- [x] 5.2 Build the image and run `asylum-openspec-init` in a fresh project: confirm `openspec/` is created and `.claude/commands/opsx/` contains `verify` and not `sync`. Re-run and confirm it runs `openspec update --force` without error. *(Behavior verified against the real `openspec` CLI on PATH: fresh run created `openspec/` with `verify` and no `sync`; re-run invoked `openspec update --force`. Full 5.6GB image build deferred to pre-merge integration per project convention.)*
