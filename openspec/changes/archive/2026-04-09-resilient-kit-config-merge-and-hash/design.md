## Context

`mergeKitConfig` and `ConfigHash` both manually enumerate `KitConfig` / `Config` fields. The `kit-credentials` change added a `Credentials` field to `KitConfig` but missed both functions — credentials were silently dropped during merge and excluded from the config hash. This is not a one-off oversight; it's a structural problem where any new field addition requires remembering to update two separate functions.

## Goals / Non-Goals

**Goals:**
- New `KitConfig` fields participate in merge automatically with a sensible default (last-wins)
- New `Config` fields participate in the config hash automatically
- Accumulating fields (Packages, Build) remain explicitly declared, but at the struct level

**Non-Goals:**
- Changing merge or hash behavior for any existing field
- Changing the `Config` or `KitConfig` struct shapes
- Addressing the top-level `Merge()` function (it has far fewer fields and less churn)

## Decisions

### mergeKitConfig: reflection with struct tags

Use `reflect` to iterate `KitConfig` fields. A `merge:"concat"` struct tag marks fields that accumulate (base + overlay). All other fields use last-wins (overlay replaces base when the overlay value is non-zero). The function becomes ~10 lines with no field enumeration.

**Why struct tags?** The merge strategy is a property of the field, not the merge function. Declaring it at the struct level means a developer adding a new field sees the `merge` tags on neighboring fields and can decide whether theirs needs one. The default (no tag = last-wins) is correct for the majority of fields.

**Why not a separate "strategy map"?** A map of field names to strategies is disconnected from the struct — the same "forgot to update" risk, just moved elsewhere.

**Alternative considered:** Starting from `result := *overlay` and only filling zero-valued fields from base. This avoids reflection but still requires knowing which fields accumulate, and the zero-value check for `int` fields (like `Count`) means `0` can never be an intentional overlay value — same limitation as today, but less explicit.

### ConfigHash: YAML serialization of the full config

Replace the hand-rolled field-by-field hash with `yaml.Marshal(cfg)` after zeroing non-runtime fields (`Version`, `Agent`, `ReleaseChannel`, `Agents`). `yaml.v3` sorts map keys deterministically. `Volumes` and `Ports` are sorted before marshaling since their YAML order is not semantically meaningful.

**Why zero non-runtime fields instead of listing runtime fields?** Listing runtime fields is the current approach and the source of the bug. Zeroing non-runtime fields inverts the default: new fields are included (safe — false positive warning) rather than excluded (unsafe — silent stale container).

**Why not hash everything including non-runtime fields?** Changing `agent` or `release-channel` doesn't affect the running container. A false warning for those would be confusing.

## Risks / Trade-offs

**[One-time hash mismatch on upgrade]** → All running containers will get a config drift warning after this change because the hash computation changed entirely. Acceptable — the warning just says to restart, and it only happens once.

**[Reflection has a runtime cost]** → Negligible. `mergeKitConfig` runs once at startup on a 12-field struct. The `reflect` overhead is unmeasurable against the Docker operations that follow.

**[`Count: 0` cannot be set intentionally via overlay]** → Same limitation as before. `Count` uses `int` where zero is both "not set" and a valid value. This is a pre-existing issue unrelated to this change; fixing it would require changing `Count` to `*int`.
