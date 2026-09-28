# Tasks

## 1. Kit field

- [x] 1.1 Add `HostOS` and `Available()` to `kit.Kit` with a `hostOS` package variable (D1); verify a table test covering an empty `HostOS`, a matching OS, and a different OS
- [x] 1.2 Set `HostOS: "darwin"` on `iterm` and `dropshare`; verify a test that both are unavailable when `hostOS` is `linux` and available when it is `darwin`

## 2. Offering surfaces

- [x] 2.1 Skip unavailable kits in `kit.AssembleConfigSnippets`; verify with `hostOS` set to `linux` that the output mentions neither kit, and with `darwin` that it does
- [x] 2.2 Skip unavailable kits in the first-run picker and generated kits block (`isSelectable`, `buildKitsBlock`); verify `BuildConfig` on `linux` mentions neither kit
- [x] 2.3 Filter unavailable kits in kit sync before new-kit detection and out of every `KnownKits` write, including the first-run and migration path (D3); verify tests that on `linux` neither kit is prompted, commented, or recorded as known, and that on `darwin` they are
- [x] 2.4 Skip unavailable kits in the `asylum config` Kits tab; verify that saving the tab on `linux` leaves an existing `dropshare:` entry unchanged
- [x] 2.5 Skip unavailable kits in the sandbox rules' "Disabled Kits" list; verify on `linux` that neither kit is listed when inactive, and that an enabled `dropshare` still appears under "Active Kits"

## 3. Docs

- [x] 3.1 Add the missing `iterm` row to `docs/kits/index.md`, marked macOS, and note in both kit pages that they are only offered on macOS; verify `mkdocs build` succeeds
- [x] 3.2 Extend the unreleased `iterm` and `dropshare` entries in `CHANGELOG.md` to say they are only offered on macOS; verify with `go vet ./... && go vet -tags integration ./integration/ && go vet -tags e2e ./e2e/ && go test ./...`
