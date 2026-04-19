## MODIFIED Requirements

### Requirement: Claude Code skill generation
The kit SHALL generate the agent-browser Claude Code skill at image-build time using `npx skills add vercel-labs/agent-browser` and stage it under the shared kit-skills root at `/opt/asylum-skills/.claude/skills/agent-browser/`. The kit SHALL set `ProvidesSkills: true` so the Claude agent launcher passes `--add-dir /opt/asylum-skills`. The kit SHALL NOT emit an entrypoint snippet that creates directories under `$HOME/.claude/skills/` or that bind-mounts the staged skill into `$HOME/.claude/skills/agent-browser/`.

#### Scenario: Skill staged under shared root at build time
- **WHEN** the Docker image is built with the agent-browser kit active
- **THEN** `/opt/asylum-skills/.claude/skills/agent-browser/SKILL.md` exists in the image and contains the upstream skill content

#### Scenario: Skill discoverable via --add-dir at runtime
- **WHEN** the container starts with the agent-browser kit active and the configured agent is Claude
- **THEN** the Claude launch command includes `--add-dir /opt/asylum-skills` and the skill is discoverable by Claude

#### Scenario: Host ~/.claude/skills/agent-browser not created in shared mode
- **WHEN** the container runs in shared agent-config mode and `~/.claude/skills/agent-browser/` does not exist on the host before the run
- **THEN** after container exit, `~/.claude/skills/agent-browser/` still does not exist on the host
