## Context

Already implemented in the cli-entrypoint change. This change records the design for completeness.

## Goals / Non-Goals

**Goals:**
- SSH init per PLAN.md section 5.6: create `~/.asylum/ssh/`, copy known_hosts, generate Ed25519 key
- Cleanup per PLAN.md section 5.7: remove images, prompt for cache cleanup, preserve agent config

**Non-Goals:**
- No code changes needed — already done

## Decisions

- SSH key comment uses `asylum@<hostname>` format
- Cleanup prompts interactively for cache removal (y/N default no)
- Agent config is explicitly preserved and user is informed

## Risks / Trade-offs

None — already implemented and working.
