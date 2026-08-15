## 1. Preservation check

- [x] 1.1 Write a throwaway script (scratchpad, not committed) that, for a given spec file, strips the H1 and the `## Purpose` section and normalizes `## ADDED Requirements` / `## MODIFIED Requirements` / a missing requirements header to `## Requirements`, then diffs that transform of the working-tree file against the same transform of the file at `HEAD`. A non-empty diff means requirement text was altered.
- [x] 1.2 Confirm the script reports "no drift" for all 80 specs before any edits (proves the check itself is sound rather than trivially passing).

## 2. Specs missing only a Purpose (3 files)

- [x] 2.1 Add `# <capability> Specification` and an authored `## Purpose` to `onboarding-npm`, `project-dir-guard`, `project-onboarding`
- [x] 2.2 Run the preservation check on these 3 and `openspec validate <cap> --type spec` for each

## 3. Specs with no headers at all (8 files)

- [x] 3.1 Add H1, `## Purpose`, and a `## Requirements` header above the first `### Requirement:` in `github-kit-credentials`, `hidden-kit-flag`, `kit-credentials`, `onboarding-wizard`
- [x] 3.2 Same for `port-allocation`, `runarg-pipeline`, `sandbox-rules`, `versioned-docs`
- [x] 3.3 Run the preservation check and per-spec validate on these 8

## 4. Specs with `## MODIFIED Requirements` (5 files)

- [x] 4.1 Rename the header to `## Requirements` and add H1 + authored Purpose in `agent-install`, `profile-config-integration`, `profile-container-setup`, `profile-image-build`, `profile-system`
- [x] 4.2 Run the preservation check and per-spec validate on these 5

## 5. Specs with `## ADDED Requirements` (49 files)

- [x] 5.1 Batch 1 — `agent-companions`, `agent-interface`, `binary-signing`, `cleanup-command`, `colored-logging`, `config-command`, `config-disabled-toggle`, `config-isolation`, `config-loading`, `config-migration`
- [x] 5.2 Batch 2 — `container-exec`, `container-image`, `copilot-session`, `deep-merge-kit-config`, `dev-release`, `docker-cli-wrapper`, `docker-kit`, `dockerfile-ordering`, `docs-site`, `e2e-testing`
- [x] 5.3 Batch 3 — `first-run-onboarding`, `github-kit`, `host-user-alignment`, `image-build`, `kit-activation-prompt`, `kit-config-sync`, `kit-defaults`, `kit-dependencies`, `kit-snippet-generation`, `kit-state-tracking`
- [x] 5.4 Batch 4 — `mise-java`, `openspec-in-container`, `openspec-init-script`, `openspec-kit`, `package-tiering`, `profile-entrypoint`, `project-entrypoint`, `project-scaffold`, `readme-landing`, `resume-migration-prompt`
- [x] 5.5 Batch 5 — `rtk-kit`, `self-update`, `shadow-node-modules`, `shell-kit`, `ssh-kit`, `stale-container-detection`, `tui-prompts`, `version-targeted-update`, `volume-shorthand`
- [x] 5.6 Run the preservation check and per-spec validate after each batch; note any spec whose requirements were too vague to summarize honestly rather than inventing a Purpose

## 6. The `copilot-mcp` stub

- [x] 6.1 Delete `openspec/specs/copilot-mcp/` — a stub for unbuilt work, recoverable from git history
- [x] 6.2 Confirm nothing else references the `copilot-mcp` capability (archived changes may mention Copilot; those stay untouched)

## 7. Verification

- [x] 7.1 Run the preservation check across all 80 specs — zero drift
- [x] 7.2 Run `openspec validate --specs` — 80/80 pass
- [x] 7.3 Run `go test ./...` and `go vet ./...` to confirm nothing in the repo reads these files (expected: unaffected)
- [x] 7.4 Report the list of specs flagged as stale or too vague in 5.6, as candidates for a follow-on content change
