## Context

The `apt` kit is a configuration-only container — it has no tools, no Docker snippet, no rules snippet. Its sole purpose is to let users list extra apt packages in their `.asylum` config, which the image builder picks up. Showing it in the config TUI and new-kit sync prompt alongside real kits is misleading.

Three UI surfaces currently enumerate kits:
1. `asylum config` TUI — Kits tab (`cmd/asylum/config.go`)
2. New-kit sync prompt (`cmd/asylum/main.go` via `config.SyncNewKits`)
3. Disabled-kits section in sandbox rules (`internal/container/container.go`)

All three need to skip hidden kits.

## Goals / Non-Goals

**Goals:**
- Exclude `apt` from all interactive kit selection surfaces
- Keep `apt` fully functional when configured manually in `.asylum`
- Provide a general mechanism so future configuration-only kits can also be hidden

**Non-Goals:**
- Changing the apt kit's activation logic or tier
- Adding hidden-kit support to the firstrun wizard (it's currently a no-op shell)

## Decisions

**Add a `Hidden` bool field to Kit, not a new Tier**

A new `TierHidden` tier would conflate visibility with activation semantics. The `apt` kit is genuinely opt-in (users must explicitly add packages) — it just shouldn't appear in selection UIs. A separate `Hidden` field keeps the two concerns orthogonal.

**Filter at each UI callsite, not in kit.All()**

`kit.All()` is the authoritative registry of all kits and is used for state tracking, config sync, and image building. Filtering there would break those consumers. Instead, each UI callsite adds a `k.Hidden` check alongside its existing `k.Tier == TierAlwaysOn` check.

## Risks / Trade-offs

- **Risk**: Future kits added as hidden might confuse users who can't find them in the UI. → Mitigation: Hidden kits still appear in the reference doc (`asylum-reference.md`) and can be configured manually.
