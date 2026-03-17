## Approach

Add a single `git config` line to `assets/entrypoint.sh` in the git config section, after the gitconfig is set up (either copied from host or created fresh). This ensures the safe.directory setting is always present regardless of which gitconfig path was taken.

## Decision: Wildcard vs Specific Path

**Wildcard (`'*'`)** chosen over scoping to `$HOST_PROJECT_DIR` because:
- Users may mount additional repos via custom volumes
- The container is ephemeral and contains only user-chosen content
- Avoids needing to enumerate all mounted repos

## Placement

After the existing git config block (line ~95 in entrypoint.sh), before the MCP configuration check. The gitconfig must be in place first since `git config --global` writes to it.
