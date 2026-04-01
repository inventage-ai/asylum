## Context

Container assembly (`container.RunArgs()`) currently builds docker run arguments procedurally: a series of `append` calls interspersed with kit-specific `if` branches that check `cfg.KitActive("docker")`, `cfg.KitActive("ports")`, etc. This has two problems: (1) `KitActive` only checks the config map, so TierAlwaysOn kits like `ports` are invisible to it, breaking port forwarding entirely; (2) the orchestrator must know about each kit by name, violating the kit system's composability model.

The credential and mount systems (`CredentialFunc`, `MountFunc`) already follow a kit-driven pattern — each kit declares what it needs and the container builder iterates. Docker run arguments should work the same way.

## Goals / Non-Goals

**Goals:**
- Every docker run argument has a declared source and priority, enabling dedup, conflict detection, and debug output
- Kits declare their container-time behavior via `ContainerFunc` — no kit-specific checks in the orchestrator
- `--debug` flag shows every argument with its source before launching
- Fix port forwarding (the immediate bug)
- Remove dead code: port release system, title kit

**Non-Goals:**
- Exec-time hooks (agent command customization, e.g. title kit's `--name`) — separate lifecycle, revisit later
- Changing how `CredentialFunc`/`MountFunc` work — they already produce volume args; they'll feed into the pipeline as arg producers but their internal API stays the same
- Changing port allocation logic itself — `ports.Allocate()` stays as-is, just called from a different place

## Decisions

### 1. RunArg as the universal unit

All docker run arguments are represented as `RunArg{Flag, Value, Source, Priority}`. Sources produce `[]RunArg` slices. A central `Resolve` function collects, deduplicates, validates, and flattens them into `[]string` for `docker run`.

**Why not just `[]string` pairs?** Source and priority metadata enable debug output and conflict detection. Without them, we'd need separate tracking structures.

### 2. Priority levels: core(0) < kit(1) < config(2) < cli(3)

The source label `"core"` is used for structural defaults (not "default", to avoid ambiguity with Go zero values). Higher priority silently overrides lower for the same dedup key. Same priority + different value = hard error with both sources named.

**Alternative considered:** Warn instead of error on same-priority conflicts. Rejected because silent wrong behavior (e.g., two kits mounting different things to the same path) is worse than a clear abort.

### 3. ContainerFunc on Kit struct

```go
type ContainerOpts struct {
    ProjectDir    string
    ContainerName string
    HomeDir       string
    Config        config.Config
}

// On Kit:
ContainerFunc func(ContainerOpts) ([]RunArg, error)
```

Same shape as `CredentialFunc`/`MountFunc` — a function field, not an interface method. Nil means "no container-time args." The opts struct mirrors `CredentialOpts` but without credential-specific fields.

**Why not reuse CredentialOpts?** It carries `Mode`, `Explicit`, `Isolation` which are credential-specific. A lean struct avoids confusion.

### 4. Dedup key extraction by flag type

The resolver classifies args by flag and extracts a dedup key:

| Flag | Key extraction |
|------|---------------|
| `-p` | container port — right side of `:` (or sole value if no `:`) |
| `-v` | container path — second `:`-delimited segment |
| `--mount` | `dst=` value parsed from comma-separated options |
| `-e` | env var name — left of `=` |
| boolean flags (`--privileged`, `--rm`, `--init`, `-d`) | the flag itself |
| `--cap-add` | the capability value |
| single-value flags (`--name`, `--hostname`, `-w`, `--add-host`) | the flag itself |

Args with the same flag type + same key are considered duplicates. Different flag types never conflict (a `-p` and `-v` can't clash).

### 5. Credential and mount funcs feed into the pipeline

`CredentialFunc` and `MountFunc` already return `[]CredentialMount`. Rather than changing their API, the container builder converts their output to `[]RunArg` with source `"<kit-name> kit (credentials)"` / `"<kit-name> kit (mounts)"` and priority 1 (kit level). The staging-dir logic for `Content`-based credentials stays in the container package — it's a host-side operation that produces a host path, which then becomes a `-v` RunArg.

### 6. Port release removed entirely

Port allocations are permanent per project directory. `Release()`, `ReleaseContainer()`, and the `release()` helper are deleted. The two callsites in main.go (project cleanup and prune) are removed. `RenameContainer()` stays for migration.

### 7. Title kit deleted

The title kit's only runtime behavior was adding `--name` to the agent exec command, which was broken by the same `KitActive` issue. Its config snippet for `tab-title` and `allow-agent-terminal-title` is unrelated functionality that lives in config directly. The kit is deleted; the config fields remain but aren't gated behind a kit check.

### 8. Debug output format

When `--debug` is passed, the resolved args are printed to stderr in a table before `docker run` executes:

```
Docker run arguments:
  --privileged                          docker kit
  -v /home/user/project:/home/...      core
  -p 10000:10000                        ports kit
  -e COLORTERM=truecolor                core

  Overrides (higher priority won):
    -p 8080:3000 (ports kit) → -p 3000:3000 (user config)
```

The container still starts normally after printing. This replaces any need for `--dry-run`.

## Risks / Trade-offs

**[Risk] Arg ordering may matter for Docker** → Docker uses last-wins for `-e` and some mount behaviors depend on order. The resolver should emit args in a deterministic order: core first, then kits (sorted by name), then user config, then CLI flags. Within a priority level, order follows the source's natural ordering.

**[Risk] `--mount` syntax is complex** → Parsing `type=volume,src=X,dst=Y` for dedup key extraction is more involved than `-v`. We need a small parser for comma-separated key=value pairs. This is bounded complexity — the format is well-defined by Docker.

**[Risk] CredentialFunc staging logic stays in container package** → Some credential mounts require writing generated content to a staging directory before producing a host path. This side effect happens during arg collection, which is slightly impure. Acceptable because it's contained and the alternative (splitting credential resolution into two phases) adds complexity for no user-facing benefit.

**[Trade-off] Same-priority conflict = hard error** → This means adding two kits that both want `--privileged` would fail. In practice this is unlikely (only the docker kit needs it), and the error message names both sources so the fix is obvious.
