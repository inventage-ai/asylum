## Context

`mergeKitConfig` was introduced by the `deep-merge-kit-config` change. Its design document lists the field-level merge semantics for every `KitConfig` field at the time of writing. The `Credentials` field was added by the separate `kit-credentials` change (same date) and was never included in `mergeKitConfig`. The omission meant that any credential configuration in an overlay layer (project `.asylum` or `.asylum.local`) was silently dropped during merge, and the lower-layer value was retained instead.

## Goals / Non-Goals

**Goals:**
- `Credentials` participates in `mergeKitConfig` with last-wins semantics, consistent with all other non-accumulating scalar fields

**Non-Goals:**
- No other fields, behaviors, or config formats are changed
- No migration needed — the fix only affects users who were relying on overlay configs for credentials (which never worked)

## Decisions

### Credentials uses last-wins semantics

`if overlay.Credentials != nil { result.Credentials = overlay.Credentials }` — identical pattern to `Disabled`, `DefaultVersion`, `ShadowNodeModules`, etc.

**Why not accumulate?** Credentials represent a complete intent: "use exactly these server IDs" or "use auto". There is no meaningful way to concatenate two credential configurations. An overlay that sets credentials should replace, not augment.

**Alternative considered:** Merging the explicit ID lists (append). Rejected — if a user sets `credentials: [nexus]` in `.asylum.local`, they want exactly that, not the union with whatever the base declared.

## Risks / Trade-offs

**[No behavioral change for existing users]** → Anyone who had credentials configured only in the global `~/.asylum/config.yaml` or in `.asylum` (not overridden by a deeper layer) is unaffected. The bug only manifested when an overlay tried to set credentials.
