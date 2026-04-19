## Context

`internal/ports/ports.go` allocates ports starting at `BasePort = 10000`, incrementing by the per-project count (default 5). The registry is at `~/.asylum/ports.json` and is file-locked. Browsers (Chrome, Firefox, Safari) treat ports ≥ 10000 — specifically a set including 10080, 10443, and several surrounding values — as restricted and refuse to connect from the address bar with `ERR_UNSAFE_PORT`. This defeats the point of forwarding those ports back to the host.

## Goals / Non-Goals

**Goals:**
- Default new allocations to a base port that browsers don't block: `7001`.
- Transparently migrate existing projects that sit in the old ≥10000 range on their next session, so users don't have to know about the change.

**Non-Goals:**
- Making the base port user-configurable. If it ever needs to be, that's a separate change.
- Preserving the exact port numbers previously assigned to a migrated project.

## Decisions

### Base port = 7001

The 7000-series is free of Chromium's unsafe-ports list and sits away from both the X11 TCP range (6000–6063) and the IRC cluster (6665–6669, 6697) that would otherwise trip us if per-project counts ever grew. Alternatives considered:
- **6001**: works for the default count of 5 (6001–6005), but if counts grow or many projects stack, we run into X11 TCP (up to 6063) and the IRC unsafe-port cluster (6665–6669, 6697). 7001 dodges both.
- **8001**: close to common dev ports (8000/8080), higher chance of collision with user services the agent itself wants to run.
- **5001**: Apple AirTunes / macOS Control Center on some versions — real conflict risk on developer laptops.

Trade-off accepted: port `7000` itself is macOS AirPlay Receiver (Monterey+). We start at `7001`, not `7000`, so AirPlay doesn't collide. `7001` is Oracle WebLogic admin's default and Cassandra inter-node — niche on typical dev laptops and easy to work around (disable the ports kit or stop the conflicting service).

### Stale-range detection: `r.Start >= 10000`

On `Allocate`, if the existing entry for the project starts at or above `10000`, treat it as stale: drop it from the registry and fall through to the "allocate new" path (which now uses the lowered base). This is a one-shot migration — after the next session the project has a sub-10000 range and stays there.

Alternatives considered:
- **Migrate everything unconditionally on first run**: more code, more surprise, and pointless for projects already below 10000.
- **Keep old ranges and add a flag to opt in**: leaves the broken state in place for anyone who doesn't know the flag exists. The whole point is that the old ports don't work — there's nothing to preserve.

### Reassignment, not in-place rewriting

We remove the old entry and call the normal allocation path rather than mutating `Start` in place. That way:
- `nextStart` picks the next free slot in the lowered range, respecting any other projects already allocated there.
- The container name on the old entry is also discarded, which is correct — the caller is starting a new session and passes the current container name.

### `nextStart` lower bound

`nextStart` currently returns `max(BasePort, highest_end)`. That continues to work: with `BasePort = 7001`, an empty registry allocates at 7001, and a registry that still contains some legacy ≥10000 entries (belonging to *other* projects not yet migrated) would push allocation above those — which is wrong for our goal. We need `nextStart` to ignore entries ≥ 10000 when picking the next start, since those are all destined to be reclaimed.

Simpler alternative considered: proactively sweep all ≥10000 entries out of the registry the first time any project allocates. Rejected — it mutates state for projects other than the one currently running, which is surprising and racy across concurrent asylum instances on the same host.

Chosen: `nextStart` only considers entries with `Start < 10000` when computing the next free slot. Entries ≥ 10000 are left untouched; they'll be cleaned up lazily when their own project next allocates.

## Risks / Trade-offs

- **Port collision with a service already running on the host in the 7001–7099 band** (WebLogic, Cassandra, etc.) → the docker run fails with a clear bind error; user can stop the conflicting service or disable the ports kit. Same failure mode as before, just at different numbers.
- **Users with bookmarks / scripts pointing at old 10000+ URLs** → those already don't work in browsers, which is the whole motivation. Acceptable.
- **Registry file accumulates stale ≥10000 entries for projects the user never opens again** → harmless (they don't block sub-10000 allocation after the `nextStart` change) and self-limiting. Not worth a migration pass.
