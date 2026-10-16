# Design

## Context

Config layers are loaded in `config.Load` and merged with `config.Merge`; `env` becomes `-e` RunArgs with source `user config (env)` and priority `PriorityConfig` in `container.go`. CLI `-e` flags are merged into `cfg.Env` by `applyFlags`. `ConfigHash` (stored as a container label) detects config drift.

## Goals / Non-Goals

**Goals:** pass host variables under another name; keep the repo layer from requesting host secrets; no new files containing values.

**Non-Goals:** masking values in `--debug` output (#40); hiding values from `docker inspect`; re-reading the host value for an already running container.

## Decisions

- **Field shape `map[container]host`.** Mirrors `env` (keys are container names), so a reader sees the same left-hand side in both maps. Alternative `${VAR}` expansion inside `env` was rejected: it would change the meaning of existing literal values containing `$` and would also work from the repo layer.
- **Filter in `Load`, not at assembly.** `Load` is the only place that knows which file a value came from. The project layer's `EnvFromHost` is cleared before merging, with a warning. Keeping the filter there means `Config` never carries a repo-provided mapping.
- **`env` wins on name collision.** Both would be `PriorityConfig` `-e` args, which `ResolveArgs` reports as a conflict. `env` holds literal values and the CLI `-e` flags; a value given explicitly is the more specific intent, so the `env-from-host` entry is skipped when `env` already has the name.
- **Values are read at assembly with `os.Getenv`.** Empty and unset are treated alike, matching the issue and the shell-workaround behavior.
- **`ConfigHash` covers the mapping, not the values.** The mapping is part of `Config` and hashed automatically; values are not, so the label never contains a hash of a secret and a rotated token does not force a restart. A changed mapping is detected as config drift like `env`.

## Risks / Trade-offs

- [A rotated host value is not picked up by a running container] → Documented; restart with `--rebuild`, same as other env changes.
- [Values are visible in `--debug` output and `docker inspect`] → Out of scope, tracked in #40; documented.
- [Silent skip of unset variables can hide a typo in the host name] → Accepted for parity with the requested behavior; the `--debug` table shows which variables were set.
