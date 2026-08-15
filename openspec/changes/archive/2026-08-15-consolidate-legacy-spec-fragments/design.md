## Context

See proposal.md - Why.

The five fragments were compared requirement-by-requirement against every capability that could plausibly own their subject. What each one actually contains:

| Fragment | Duplicated by | Unique content |
|----------|---------------|----------------|
| `agent-install` | `config-isolation` "Config isolation levels", "First-run isolation prompt" | how the YAML field parses, and that omitting it yields an empty value |
| `profile-config-integration` | `kit-defaults` "Kit disabling" | `disabled: false` at project level re-enables a globally-disabled kit |
| `profile-container-setup` | `config-isolation` "Config isolation levels" (same paths, same three modes) | none — only an incorrect `(default)` label |
| `profile-image-build` | `host-user-alignment` "Container home directory matches host" | the base image hash includes the home directory path |
| `profile-system` | `kit-dependencies`, `kit-defaults` "Kit activation tier" | a field list naming `DefaultOn`, which no longer exists |

`container-assembly` was considered as the home for `profile-container-setup` and rejected. Its "Agent-specific mounts and env vars" requirement says the container mounts "the agent's asylum config dir" — wording that presumes `isolated` mode and does not describe the per-mode paths at all. `config-isolation` already states all three, correctly.

## Goals / Non-Goals

**Goals:**

- One capability per subject: isolation defaults stated once, kit declaration stated once.
- No requirement disappears without either a home in another spec or a written justification for dropping it.
- The corpus stops contradicting itself about the agent config isolation default.

**Non-Goals:**

- Renaming `profile-entrypoint`. It is the last legacy-named capability, but unlike the five it carries three real requirements about entrypoint assembly. Renaming a capability is a different kind of edit and belongs in its own change.
- Writing a spec for the agent install system (`AgentInstall`, `RegisterInstall`, Dockerfile snippets). It has never had one, and inventing requirements for existing code from scratch is how specs become fiction.
- Re-specifying the `Kit` struct. The 28 fields are described where they are used — `kit-snippet-generation`, `hidden-kit-flag`, `sandbox-rules`, `runarg-pipeline`, `kit-credentials` each specify the fields they introduce. A central enumeration went stale once already.

## Decisions

**Delete duplicates rather than merging them.**

Where a fragment says the same thing as an existing requirement, the existing one wins and the fragment is dropped. Merging near-identical requirements produces a spec that says the same thing twice in slightly different words, which is how the corpus arrived here. Only genuinely unique content moves.

**Correct the specs to match the code, not the reverse.**

Both wrong claims — `isolated` as the default in `profile-container-setup` and in `first-run-onboarding` — describe behavior the code does not have. `ResolveConfigDir` falls through to the native host config dir when no isolation is resolved (`internal/agent/agent.go:63`), which is `shared`. The specs are wrong; the behavior is fine and is not being touched.

**Verify by inventory, not by diff.**

The previous change could prove itself with a text-preservation check because nothing was allowed to change. That check is meaningless here — this change exists to change requirement text. The safety net instead is an explicit inventory: enumerate every `### Requirement:` heading present before and after, and account for each removed one as either relocated (name it in its new home) or dropped-as-duplicate (name the requirement that covers it). A removal that cannot be accounted for is a bug.

**Do not sweep for other stale content.**

While scoping, `ssh-kit`'s `isolated (default)` was checked and is correct — SSH key isolation genuinely defaults to `isolated`, unlike agent config isolation. Any broader consistency sweep must verify each claim against the code individually, which is a larger change than this one.

## Risks / Trade-offs

**A requirement is dropped as "duplicate" when it says something subtly different.** → The inventory names the covering requirement for every drop, so the claim is auditable rather than asserted. `profile-config-integration` is the proof this matters: it looked like a pure duplicate of "Kit disabling" until the `disabled: false` re-enable scenario turned up.

**Deleting `agent-install` removes a capability named after a real subsystem.** → The name suggested coverage that never existed; its content was a config field. Leaving a misnamed fragment in place to preserve a directory name is worse than the gap, which is now recorded in the proposal.

**Someone re-adds a central `Kit` struct requirement later.** → Non-Goals says why not, with the evidence: the last attempt named a field that has since been removed and missed 18 that were added.

## Migration Plan

Not applicable — no runtime component. Rollback is `git revert`.
