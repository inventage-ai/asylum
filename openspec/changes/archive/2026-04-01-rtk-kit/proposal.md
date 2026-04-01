## Why

RTK is a high-performance CLI proxy that reduces LLM token consumption by 60-90% by intelligently filtering and compressing shell command outputs. It intercepts commands like `ls`, `grep`, `git diff`, `docker ps`, etc. and strips noise (comments, whitespace, boilerplate) while preserving essential context. Adding RTK as an opt-in kit gives asylum users an easy way to reduce token usage and costs for any agent running in the sandbox.

## What Changes

- New `rtk` kit (opt-in) that installs RTK and generates hook artifacts at Docker build time
- Entrypoint mounts RTK hooks and awareness doc into `~/.claude/`, registers the PreToolUse hook in settings.json (same pattern as ast-grep skill mounting)
- Rules snippet documenting RTK commands (`rtk gain`, `rtk discover`) for agent awareness
- Banner line showing RTK version
- Documentation page for the kit

## Capabilities

### New Capabilities
- `rtk-kit`: Kit definition, installation, build-time hook generation, runtime mounting, rules, and documentation for the RTK token-reduction proxy

### Modified Capabilities

_(none)_

## Impact

- **Code**: New file `internal/kit/rtk.go` + docs page `docs/kits/rtk.md`
- **Dependencies**: RTK installed via its install script (curl-based, no new Go dependencies)
- **Images**: Base image size increases slightly (~5MB single Rust binary)
