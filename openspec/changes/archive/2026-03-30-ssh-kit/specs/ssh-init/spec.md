## REMOVED Requirements

### Requirement: SSH directory initialization
**Reason**: Replaced by the always-on `ssh` kit which generates keys and mounts credentials automatically via its `CredentialFunc`.
**Migration**: Keys in `~/.asylum/ssh/` continue to work. The `asylum ssh-init` command is removed; key generation happens automatically on first container start.
