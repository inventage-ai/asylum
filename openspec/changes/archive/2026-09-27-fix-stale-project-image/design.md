# Design

## Context

`EnsureProject` names and checks the project image by a hash of its generated Dockerfile. That Dockerfile starts with `FROM asylum:latest`, so its text never changes when the base does. The only base signal is the `baseRebuilt` flag, which is true only in the invocation that rebuilt the base.

## Goals / Non-Goals

**Goals:**

- Rebuild a project image whenever the base it was built on is no longer current.

**Non-Goals:**

- Pruning project images that no project uses anymore.

## Decisions

### D1. Record the base hash as a label

The project image gets an `asylum.base.hash` label holding the base image's `asylum.hash` at build time. `EnsureProject` compares the label with the current base hash.

- **Alternative: mix the base hash into the project tag.** Every base rebuild would then leave a stale tagged image per project, and the tag would change on every base rebuild. A label keeps the tag stable. The rebuild retags it, and stale-container detection already compares image IDs.

### D2. Drop `baseRebuilt`

A rebuilt base carries a new `asylum.hash`, so the label check covers every case the flag did. `--no-cache` still forces the project build through `noCache`.

### D3. Inspect failure

If reading the base label fails, `EnsureProject` returns the error. The caller already treats `EnsureProject` errors as "use the running container" when one runs, and as fatal otherwise, the same as for `EnsureBase`.

## Risks / Trade-offs

- **Every project image rebuilds once after upgrading**, because existing images carry no `asylum.base.hash` label. That is the correct outcome for any image built on an older base, and it is a one-time cost.
