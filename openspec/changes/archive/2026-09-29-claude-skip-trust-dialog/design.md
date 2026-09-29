# Design

## Context

Claude Code 2.1.284 decides trust like this. The behaviour was read from the binary.

| Check | Order |
|---|---|
| `CLAUDE_CODE_SANDBOXED` non-empty | 1st; trusted immediately |
| `projects[<cwd>].hasTrustDialogAccepted` in Claude's config | 2nd |
| Same flag on a parent directory | 3rd |
| Home directory | Trust is session-only; never persisted |

The variable has 11 uses in the binary. All of them are trust decisions, the project permission-rule gate, env-passing lists, or env scrubbing. None of them changes the system prompt or tool descriptions.

## Goals / Non-Goals

**Goals:**
- No trust dialog in asylum Claude sessions, home directory included.

**Non-Goals:**
- A dedicated config option. `env:` covers the opt-out.
- Changing behaviour for other agents.

## Decisions

### Env var over seeding `.claude.json`
Under `shared` isolation, which is the default, Claude's config file is the host's `~/.claude/.claude.json`. A trust entry written by asylum would also make host sessions trust the folder. A seeded entry also doesn't help in the home directory. The env var exists only inside the container.

### Set in `Claude.EnvVars()`
The variable has a fixed value and needs no container name, so it fits the agent's static env map next to `CLAUDE_CODE_IDE_HOST_OVERRIDE`. `coreEnvVars` already merges companion env vars, so companion sessions need no extra code.

### Opt out with an empty string
User `env:` entries resolve at `PriorityConfig`, which beats core agent env when `ResolveArgs` dedupes by variable name. Claude checks the value for JavaScript truthiness, so `"0"` still counts as on and only `""` turns it off. The docs spell this out.

## Risks / Trade-offs

- [Repo-supplied hooks, MCP servers and `apiKeyHelper` run at session start without a prompt] → Documented in the security model. Asylum already runs with `--dangerously-skip-permissions` and tells users not to run distrusted code inside it.
- [The variable is undocumented and may change] → Failure means the dialog comes back. Nothing else breaks.
- [Users try `"0"` to opt out] → The docs name `""` explicitly.
