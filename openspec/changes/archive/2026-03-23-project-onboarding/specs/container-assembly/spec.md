## REMOVED Requirements

### Requirement: PreRunCmds in exec args
**Reason**: Replaced by the onboarding framework which runs commands via separate `docker exec` calls from Go, not by wrapping the agent command in `bash -c`.
**Migration**: Use onboarding tasks instead of `PreRunCmds`.
