## Context

See proposal.md - Why.

The 80 specs fall into four shapes, which is what makes the work batchable:

| Shape | Count | Missing |
|-------|-------|---------|
| `# <name> Specification` + `## Purpose` + `## Requirements` | 14 | nothing — already valid |
| `## ADDED Requirements` only | 49 | H1, Purpose; header is delta vocabulary |
| `## MODIFIED Requirements` only | 5 | H1, Purpose; header is delta vocabulary |
| No `##` headers at all — file opens on `### Requirement:` | 8 | H1, Purpose, Requirements |
| `## Requirements` but no Purpose | 4 | H1, Purpose |

The validator's rule is narrow: a spec must contain `## Purpose` and `## Requirements`. It does not check the H1, and it does not check that every requirement has a scenario at the file level — but every requirement in the corpus already has at least one scenario, verified by counting `#### Scenario:` blocks per `### Requirement:` block. So no scenario authoring is needed anywhere.

The 14 valid specs were all produced by `openspec archive`'s spec-sync path, which emits the canonical shape. The other 66 predate it or were merged by hand. That makes the target shape unambiguous: match what the tooling generates.

## Goals / Non-Goals

**Goals:**

- All 80 specs pass `openspec validate --specs`.
- Requirement and scenario text is preserved byte-for-byte. This change must be provably presentation-only.
- Each `## Purpose` is worth reading — it tells someone opening the file cold what the capability is for.

**Non-Goals:**

- Rewriting, splitting, merging, or pruning requirements. Some specs are certainly stale, but deciding that requires judgment per capability and belongs in its own change.
- Wiring `openspec validate --specs` into CI. Worth doing once the corpus is green, but a separate concern — this change makes it possible, not mandatory.
- Reorganizing the capability namespace (80 flat directories, some clearly related).

## Decisions

**Author each Purpose from the spec's own requirements rather than templating it.**

A generated line like "Defines the port-allocation capability." passes the validator and carries zero information. The Purpose section is the one place a reader learns what a capability is *for* before reading a wall of SHALL statements, so a formulaic one actively wastes the reader's attention while making the corpus look maintained. The cost is real — 66 specs have to be read — but it is the only part of this change that produces value beyond a green checkmark. Alternative considered: template now, improve later. Rejected because "later" never arrives for cosmetic debt, and the templated text would then be indistinguishable from a considered one.

**Derive Purpose only from text already in the file.**

The risk of authoring 66 summaries is inventing behavior. Each Purpose must be a restatement of requirements present in that file, not knowledge about the codebase. If a spec's requirements are too thin to summarize, that is a signal the spec is stale — note it, don't paper over it with invented intent.

**Rename `ADDED`/`MODIFIED` headers rather than restructuring the content beneath them.**

Those words are delta vocabulary: they describe what a change did to a spec, not what the spec says. They leaked into main specs through hand-merges. The requirement blocks beneath them are already in the correct format, so the fix is the header line alone.

**Verify preservation mechanically, not by review.**

Reviewing 66 diffs by eye will miss a dropped line. Instead: for each file, strip the H1 and the Purpose section, normalize `## ADDED Requirements` / `## MODIFIED Requirements` / absent-header to `## Requirements`, and diff the result against the same transform applied to the file at `HEAD`. Any non-empty diff is a bug in the edit. This turns "did we preserve the text" from a judgment call into a check, and it is the reason this change can safely touch 66 files at once.

**Batch by shape, not alphabetically.**

Each of the four groups is a uniform edit, so grouping by shape keeps the transformation consistent within a batch and makes a mistake systematic (and therefore obvious) rather than scattered.

## Risks / Trade-offs

**Requirement text silently altered while editing 66 files.** → The mechanical preservation check above catches any drift before commit. It is a task in its own right, not an afterthought.

**Authored Purpose text misrepresents a capability.** → Constrained to restating requirements present in the file. Where a spec is too vague to summarize honestly, the task list calls for flagging it rather than guessing.

**Deleting `copilot-mcp` loses a record of intended work.** → It documents deferred, unbuilt work (`COPILOT_MCP_DIR`, a mount path that "Copilot may consume"). Formalizing it into a proper Requirement + Scenario would make an unimplemented idea look like a specified capability, which is worse than losing it: `openspec/specs/` is read as a description of what the system does. Deletion is reversible through git history, and the archived changes that reference Copilot remain untouched.

**Conflicts with any in-flight change that archives a spec.** → The change is doc-only and touches no code, so a conflict is resolvable by re-running the transform on the conflicted file. Worth landing quickly rather than letting it sit.

## Migration Plan

Not applicable — no runtime component, no data, nothing deployed. Rollback is `git revert`.
