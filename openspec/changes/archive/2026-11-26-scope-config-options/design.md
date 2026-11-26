# Design

## Context

`config.Load` reads the global, project and local files with `LoadFile` (a plain `yaml.Unmarshal` into `Config`) and merges them with `Merge`. Two hand-written guards in `Load` remove `env-from-host` and `agents.*.seed` from the project layer after decoding. `NeedsMigration`/`MigrateV1ToV2` rewrite a v1 file in place before it is read. `Load` is called from six places: self-update, the main run path (up to three times because of the first-run wizard reloading), `sessionFilter` (with `$HOME` as project dir, so without project files), and `asylum config`. `internal/config` imports `internal/kit`, so kit code cannot import config types. Per-project host state lives in `~/.asylum/projects/<container>/`.

## Goals / Non-Goals

**Goals:**
- One table that says where each option is accepted, derived from the code that defines the options.
- A forgotten scope fails the tests.
- Rejections carry file and line, so the prompt can point at them.

**Non-Goals:**
- Warnings for unknown keys or typos. The node walk would make them easy, but they are a separate change.
- Validating option values.
- Protecting config files from the agent: the #47 follow-up.

## Decisions

### Filter the YAML node tree before decoding
Each layer is parsed into a `yaml.Node` and matched against the scope table. Out-of-scope nodes are removed, then the remaining tree is decoded into `Config`.

Alternatives considered:
- **Filtering the decoded struct.** Rejected for three reasons:
  - It loses line numbers.
  - It cannot see kit-specific keys outside `KitConfig` (host-mcp `servers`).
  - It cannot tell "kit enabled" from "kit not mentioned".
- **Struct tags alone.** Also rejected: a tag cannot express that the map key `kits.docker` is personal.

### The table is built from struct tags and kit declarations
- Every field of `Config`, `AgentConfig` and `KitConfig` carries a `scope` tag (`any`, `personal`, `global`), including `any`. A reflection test fails when a field has no tag, so each new option forces a decision. Fields tagged `yaml:"-"` are skipped.
- Map levels become wildcards: `agents.*.seed`, `kits.*.credentials`.
- `kit.Kit` gains:
  - `Scope`: the scope for enabling the kit; empty means `any`.
  - `OptionScopes map[string]Scope`: kit-specific options such as host-mcp's `servers`.
- The `Scope` type lives in `internal/kit`, because config imports kit and not the other way round. A separate package for one type was rejected (no utility packages).
- Lookup order: a kit's own declaration wins over the generic `kits.*.<option>` tag.

### Only plain string keys where entries are checked
In `.asylum` and `.asylum.local`, keys at the top level, under `kits` and `agents`, and directly inside a kit or agent entry must be untagged string scalars; anything else is an error. The filter looks up scopes by the key's text, so that text must equal what the decoder sees:
- A merge key (`x: &v {volumes: [...]}` plus `<<: *v`) would pull in an entry the filter never saw.
- An alias key (`*k` with `&k volumes`) or a tagged key (`!!binary dm9sdW1lcw==`) decodes to `volumes` while its text is something else.

Aliases as keys can also imitate the `disabled: true` exception; that is caught too, because every accepted kit entry is descended into. Aliases inside a value (a `packages` list) stay allowed. Expanding them before filtering was rejected: it needs alias-bomb protection and is more code than refusing a case nobody uses in a small config file.

### Disabling stays possible everywhere
In a layer that does not accept enabling a personal kit, the kit's node is kept only if its mapping is exactly `disabled: true`. Otherwise the whole node is rejected, because only narrowing is safe.

### Rejections travel on the Config
- `Config` gets `Rejected []Rejection` with `yaml:"-"`. A `Rejection` holds the file, line, path, the value re-encoded as YAML, and the scope.
- `Merge` concatenates the field.

An extra return value from `Load` was rejected: it touches all six call sites, and only two of them use it.

### Order inside Load
Per project file, in this order:
1. Resolve the path.
2. Git tracking check (local layer only).
3. Migrate.
4. Parse into a node.
5. Filter by scope.
6. Decode.

Why this order:
- The git check runs before migration, because migration writes to the file.
- The scope filter runs after migration, because it only knows v2 keys. Migration only restructures and never widens access.

The global file goes through the same filter. Every scope accepts the global layer, so nothing is rejected there.

### Git: check in Load, exclude at session start
- **Tracking check:** `git -C <project> ls-files -z -- .asylum.local .asylum.local.yaml`.
  - Only runs when the local file exists, so most projects never call git.
  - Lives in `Load`, so every command that would honour the file refuses it.
  - Any failure to run git (not a repo, git missing) counts as "not tracked".
- **Exclude step:** `git check-ignore` on both filenames, then append the ones not ignored yet to the path from `git rev-parse --git-path info/exclude`.
  - Relative output is resolved against the project directory. `--git-path` handles linked worktrees and submodules.
  - Runs once, in the main run path only, not inside `Load`. `sessionFilter` loads `$HOME` as the project dir, and self-update or `asylum config` should not write into repositories.
  - The patterns are unanchored, so they match when the project is a subdirectory of the repository.

### Prompt in the main run path
- Runs right after the first `Load`, before the first-run wizard. Rejections come only from project files, so the wizard (which writes the global config) cannot change them.
- Uses `tui.Select` with two options:
  - "Start without these options" (default)
  - "Abort"
- The prompt shows each rejection as `file:line`, the path, the YAML value, and the accepting files. The wording comes from the scope:
  - personal: "`.asylum.local` or `~/.asylum/config.yaml`"
  - global: "`~/.asylum/config.yaml`"
- **Remembered decision:** a SHA-256 over the sorted `file\0path\0value` entries, stored as one line in `~/.asylum/projects/<container>/rejected-config`.
  - It's per project, next to other per-project state. `~/.asylum/state.json` was rejected because it is global and owned by kit sync.
  - When the hash matches, asylum logs one line listing the ignored paths.
- **No terminal:** one warning per rejection, then continue.
- **self-update and `asylum config`:** the same warnings through a shared helper, no prompt.

### Remove the hand-written guards
- The two blocks in `Load` are deleted; `EnvFromHost` and `AgentConfig.Seed` get `scope:"personal"`.
- `kits.*.schemes` is tagged personal on `KitConfig`. Only browser-open uses it, so no kit-specific entry is needed.

## Risks / Trade-offs

- **Committed `.asylum` files with volumes, docker or ssh `shared` stop applying.** → The prompt shows the exact YAML and target file, plus a **Changed** CHANGELOG entry and updated docs. This repository's own `.asylum` is fixed in the same change.
- **A whole personal-kit entry is rejected even if the global config already enables the kit and `.asylum` only adds options.** → Accepted. Personal kits have almost no options, and the prompt names the entry.
- **The agent can still write `.asylum.local`, `.git/info/exclude`, or mount `~/.asylum` from `.asylum.local`, so "personal" and "global" are not hard boundaries against the agent.** → Out of scope, addressed by the #47 follow-up. The docs state it.
- **Extra git calls at startup.** → `check-ignore` and `rev-parse` add a few milliseconds per session start. `ls-files` runs only when the local file exists.
- **Asylum writes into `.git/info/exclude` without asking.** → It only writes when nothing ignores the file yet, and it logs one line. The file is local and never shared.
- **The prompt shows values from `.asylum`.** → Every option that can be rejected holds paths, names or switches, never secret values: `env` is `any` and so never rejected, and `env-from-host` holds variable names only.
- **PR #50 (host-mcp) touches the same area.** → Whichever change merges second adds `Scope: Personal` and `OptionScopes{"servers": Global}` to host-mcp and drops its own project-layer handling.

## Migration Plan

- Nothing is migrated automatically.
- Move this repository's `volumes` entry from `.asylum` to `.asylum.local`.
- Users see the prompt once per change set and move the entries themselves.
- Rollback: revert the change. No file formats change, and the `rejected-config` state file is ignored by older versions.
