## 1. Inventory baseline

- [x] 1.1 Record every `### Requirement:` heading across `openspec/specs/` at `HEAD` (capability + requirement name) to a scratchpad file, as the baseline the final audit compares against

## 2. Relocations

- [x] 2.1 Add a requirement to `config-isolation` covering how `agents.<agent>.config` parses, carrying over both scenarios from `agent-install` (value set → parsed; value omitted → empty, triggering the first-run prompt)
- [x] 2.2 Extend `kit-defaults` "Kit disabling" with the `disabled: false` project-level re-enable scenario from `profile-config-integration`, and state that presence in the `kits` map means configured
- [x] 2.3 Add a requirement to `host-user-alignment` stating that the base image hash includes the home directory path, carrying over the different-home-triggers-rebuild scenario from `profile-image-build`

## 3. Deletions

- [x] 3.1 Delete `openspec/specs/agent-install/` (content relocated in 2.1)
- [x] 3.2 Delete `openspec/specs/profile-config-integration/` (unique content relocated in 2.2; remainder duplicates `kit-defaults` "Kit disabling")
- [x] 3.3 Delete `openspec/specs/profile-image-build/` (unique content relocated in 2.3; `USER_HOME` arg duplicates `host-user-alignment`)
- [x] 3.4 Delete `openspec/specs/profile-container-setup/` (duplicates `config-isolation` "Config isolation levels"; nothing to relocate)
- [x] 3.5 Delete `openspec/specs/profile-system/` (duplicates `kit-dependencies` and `kit-defaults` "Kit activation tier"; residue names the removed `DefaultOn` field)
- [x] 3.6 Confirm nothing outside `openspec/changes/archive/` references the five deleted capabilities

## 4. Corrections

- [x] 4.1 Fix the "Non-interactive mode" scenario in `first-run-onboarding` so it describes the `shared` fallback instead of "isolated config"

## 5. Verification

- [x] 5.1 Produce the post-change requirement inventory and diff it against the 1.1 baseline; account for every removed heading as relocated (naming its new home) or dropped-as-duplicate (naming the covering requirement)
- [x] 5.2 Run `openspec validate --specs` — all remaining specs pass
- [x] 5.3 Grep the corpus for any surviving claim that agent config isolation defaults to `isolated`; confirm the only `isolated (default)` hits left belong to `ssh-kit`, where it is correct
- [x] 5.4 Run `go test ./...` and `go vet ./...` to confirm nothing reads these files (expected: unaffected)
