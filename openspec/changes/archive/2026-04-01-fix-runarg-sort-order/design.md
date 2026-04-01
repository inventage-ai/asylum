## Context

The unified RunArg pipeline treats `run` and `-d` as regular RunArgs with `Source: "core"` and `Priority: 0`. `ResolveArgs` sorts all args deterministically by priority → source → dedup key. Since `run` gets dedup key `"other:run="` and `--cap-add` gets `"--cap-add:SYS_ADMIN"`, `--cap-add` sorts before `run`. Docker then receives `docker --cap-add ... run` and fails with "unknown flag".

## Goals / Non-Goals

**Goals:**
- `docker run -d` always appears as the first two args passed to `docker`, regardless of sort order
- Minimal change to the RunArg pipeline — the fix should not require new priority levels or special sorting logic

**Non-Goals:**
- Rethinking the sort algorithm or dedup key scheme
- Changing how `--rm`, `--init`, or other structural flags are handled (they work fine anywhere after `run`)

## Decisions

**Remove `run`/`-d` from the RunArg pipeline; prepend them as a fixed prefix.**

`run` is the docker subcommand, not an option. `-d` is tightly coupled (detached mode is not subject to override). Neither should participate in dedup, conflict detection, or sorting.

The change: stop emitting `core("run", "")` and `core("-d", "")` in `RunArgs()`. Instead, prepend `["run", "-d"]` to the flattened `[]string` before appending the image and command.

Alternative considered: adding a sort-order field or special priority. Rejected because it adds complexity to handle something that isn't really a RunArg at all.

## Risks / Trade-offs

- **`--debug` output will no longer show `run` and `-d`** — acceptable, they're always present and showing them adds no information.
- **`-d` is removed from `booleanFlags`** — it no longer flows through dedup. Since nothing else ever emits `-d`, this has no practical impact.
