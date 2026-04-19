# ast-grep Kit

AST-based code search, lint, and rewrite via [ast-grep](https://ast-grep.github.io/) (`sg`).

**Activation: Opt-in** — only active if explicitly enabled in your config.

## What's Included

- **sg** — CLI for searching, linting, and rewriting code using abstract syntax tree patterns
- **Claude Code skill** — staged inside the container at `/opt/asylum-skills/.claude/skills/ast-grep/` and loaded via Claude's `--add-dir` flag, teaching Claude how to write ast-grep rules (sourced from the upstream [agent-skill](https://github.com/ast-grep/agent-skill) project). Nothing is written to your host `~/.claude/`.

## Configuration

```yaml
kits:
  ast-grep: {}
```

## Dependencies

Depends on the [Node.js](node.md) kit (installed via npm).

## Usage

```sh
# Search for a pattern
sg run --pattern 'console.log($ARG)' --lang js

# Lint with rules
sg scan

# Rewrite matches
sg rewrite --pattern 'console.log($ARG)' --rewrite 'logger.info($ARG)' --lang js
```

Patterns use `$VAR` as wildcards to match any AST node. See the [ast-grep documentation](https://ast-grep.github.io/) for full pattern syntax.
