## Why

`openspec validate --specs` fails on 66 of the 80 capability specs in `openspec/specs/`, so the command reports nothing useful and cannot be used as a gate. The failures are not substantive: the corpus accumulated three different document shapes as changes were archived over the project's life, and the validator now requires exactly one. A validator that always fails is a validator nobody reads, which means a genuinely malformed spec would land unnoticed.

## What Changes

- Every spec under `openspec/specs/` gains the two sections the validator requires, `## Purpose` and `## Requirements`, matching the shape the archive tooling produces today (a `# <capability> Specification` heading, then Purpose, then Requirements).
- Delta-relative headers left behind by archiving (`## ADDED Requirements`, `## MODIFIED Requirements`) are renamed to `## Requirements`. These describe a *change* to a spec, not the spec itself, and are meaningless in a merged main spec.
- Each spec gains a `## Purpose` section written from its own requirements: one to three sentences on what the capability is for. This is the only non-mechanical part of the change.
- `openspec/specs/copilot-mcp/spec.md` is **deleted**. It is the one spec with no `### Requirement:` blocks at all — a stub in a bespoke format specifying `COPILOT_MCP_DIR`, a mount path Copilot "may consume", that ends "Implementation deferred after spec refinement". `openspec/specs/` should record what exists, not what was once planned. Recoverable from git history if the work is ever picked up.
- No code changes. No behavior changes. No requirement text is altered, added, or removed.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

(none — no requirement text changes. This change edits the presentation of existing specs only, so `.openspec.yaml` sets `skip_specs: true`.)

## Impact

- 66 of 80 files under `openspec/specs/*/spec.md`: 65 edited, `copilot-mcp` deleted. The other 14 are already valid and are not touched.
- `COPILOT.md`, which pointed at the deleted `copilot-mcp` spec as "Asylum's MCP plumbing contract". That pointer is replaced with a statement that asylum contributes no MCP plumbing for Copilot, since nothing in the Go source ever referenced `COPILOT_MCP_DIR`.
- No Go source, no assets, no docs site, no CHANGELOG entry — nothing here is user-facing.
- After this change, `openspec validate --specs` passes repo-wide and can be wired into CI as a follow-on if wanted.
