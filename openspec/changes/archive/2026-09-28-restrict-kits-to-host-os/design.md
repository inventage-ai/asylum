# Design

## Context

Kits reach users through seven places:

| Surface | Code |
|---|---|
| First-run wizard picker | `firstrun/config_writer.go` `isSelectable` |
| First-run generated kits block | `firstrun/config_writer.go` `buildKitsBlock` |
| Non-interactive default config | `kit.AssembleConfigSnippets`, used by `WriteDefaults` and `asylum config` |
| Kit-sync prompt and declined-kit comment | `config/kitsync.go` |
| First-run and migration "mark all known" | `config/kitsync.go` (`state.KnownKits = kit.All()`) |
| `asylum config` Kits tab | `cmd/asylum/config.go` |
| Sandbox rules "Disabled Kits" | `container/container.go` |

`Hidden` already keeps kits out of pickers, but hidden kits still get comment lines and stay fully usable. A kit for another OS needs neither.

## Goals / Non-Goals

**Goals:**

- One field and one check, applied at each surface.
- Tests that exercise both hosts on Linux CI.

**Non-Goals:**

- Blocking or warning when a kit for another OS is enabled.
- Removing comment lines earlier versions already wrote into Linux configs.
- Support for several OSes per kit.

## Decisions

### D1. `HostOS string` on `kit.Kit`

The field holds a `runtime.GOOS` value, and an empty string means any host. A package variable `hostOS = runtime.GOOS` lets tests pretend to be another OS. The method `k.Available()` reports `k.HostOS == "" || k.HostOS == hostOS`.

- **Alternative: `Platforms []string`.** Both kits need exactly one OS, and CLAUDE.md warns against speculative flexibility.
- **Alternative: reuse `Hidden`.** Hidden kits still get comment lines and are meant to be used, so they need different behaviour.

### D2. Filter at the offering surfaces, not in `kit.Resolve`

Resolution stays untouched, so a kit enabled on the wrong OS works exactly as before. Each surface in the Context table adds the `Available()` check next to its existing tier and `Hidden` checks.

`kit.All()` stays the full registry, because resolution and config parsing still need every kit.

### D3. Unknown on the wrong host, not known

Kit sync filters kits for another OS out before computing new kits, and leaves them out of every `KnownKits` write. That includes the first-run and migration path, which today marks the whole registry known. Such a kit then stays "unknown" on Linux and is skipped each run at no cost. A Mac using the same `~/.asylum` still gets it offered.

- **Alternative: mark it known everywhere.** A Linux first run would then suppress the offer on a Mac forever.

## Risks / Trade-offs

- **A new surface could forget the check.** Every surface already checks `Hidden`, so `Available()` goes right next to it, and a test per surface covers both hosts.
