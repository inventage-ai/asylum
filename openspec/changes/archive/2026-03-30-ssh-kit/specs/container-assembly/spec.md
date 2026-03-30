## REMOVED Requirements

### Requirement: Hardcoded SSH volume mount
**Reason**: SSH mounting is now handled by the `ssh` kit's credential function, which returns `CredentialMount` entries processed by the existing credential loop.
**Migration**: No user action needed. The `~/.asylum/ssh/` → `~/.ssh` bind mount is replaced by individual file mounts from the kit's credential function.
