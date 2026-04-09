## Context

The `ConfigHash` function serializes runtime-relevant config into a deterministic hash stored as a Docker label at container creation time. Asylum compares this label on the running container to detect config drift and warn the user to rebuild. The original implementation covered volumes, env vars, and ports but omitted kit credentials — a runtime-relevant value because credential mounts are injected at container start via bind mounts.

## Goals / Non-Goals

**Goals:**
- Kit credential settings (auto mode, or the sorted explicit ID list) are included in the config hash

**Non-Goals:**
- Changing what constitutes "runtime-relevant" config beyond credentials
- Altering the hash format or label name

## Decisions

### Credentials serialized as `c:<kitName>=auto` or `c:<kitName>=<sorted-ids>`

Iterate `cfg.Kits` in sorted key order. For each kit with non-nil `Credentials`: write `c:<kitName>=auto` for auto mode, or `c:<kitName>=<comma-joined-sorted-ids>` for explicit mode. Kits with nil credentials are skipped (off = no contribution to hash, same as absent).

**Why sorted?** Determinism — explicit ID lists could be written in any order in the YAML.

**Why skip nil?** Absent credentials don't affect container behavior, consistent with how absent volumes/env don't contribute.

## Risks / Trade-offs

**[Existing running containers get a false drift warning on first run after upgrade]** → Any container started before this fix has no credential contribution in its stored hash. After the fix, the computed hash will differ if the kit has credentials configured, triggering a warning. This is acceptable — the container genuinely needs to be restarted to pick up credential mounts.
