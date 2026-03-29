## 1. Port allocation package

- [x] 1.1 Create `internal/ports/ports.go` with `Range` type (`Project string`, `Start int`, `Count int`), `Allocate(projectDir string, count int) (Range, error)`, `Release(projectDir string) error`, and file-locked JSON read/write helpers
- [x] 1.2 Add tests for allocation (first project, subsequent, reuse, concurrent locking, range extension, extension blocked by neighbor)
- [x] 1.3 Add test for release

## 2. Config accessor

- [x] 2.1 Add `PortCount()` method to `Config` that reads `count` from the ports kit's `KitConfig`, defaulting to 5. Reuse an existing `KitConfig` field or add a `Count` int field.
- [x] 2.2 Add test for `PortCount` default and explicit values

## 3. Ports kit definition

- [x] 3.1 Create `internal/kit/ports.go` registering the `ports` kit with `DefaultOn: true`, no snippets, and a description
- [x] 3.2 Add the kit to the default config in `internal/config/defaults.go` (commented out like other default-on kits)

## 4. Container integration

- [x] 4.1 Add `AllocatedPorts []int` field to `RunOpts` in `internal/container/container.go`
- [x] 4.2 Update `RunArgs` to append `-p <port>:<port>` for each allocated port (after user-configured ports)
- [x] 4.3 Add test that allocated ports appear in `RunArgs` output

## 5. Sandbox rules integration

- [x] 5.1 Add `AllocatedPorts []int` parameter to `generateSandboxRules` and emit a "Forwarded Ports" section when non-empty
- [x] 5.2 Add test for rules generation with and without allocated ports

## 6. Wire up in main

- [x] 6.1 In `cmd/asylum/main.go`, before `RunArgs`: if ports kit is active, call `ports.Allocate(projectDir, cfg.PortCount())` and pass result to `RunOpts.AllocatedPorts`
- [x] 6.2 In cleanup path: remove `ports.json` when removing cached data

## 7. Changelog and reference doc

- [x] 7.1 Add entry under Unreleased/Added in `CHANGELOG.md`
- [x] 7.2 Update `assets/asylum-reference.md` to document the ports kit and port allocation
