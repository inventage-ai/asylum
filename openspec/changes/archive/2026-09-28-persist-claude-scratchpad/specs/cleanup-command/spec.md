# Spec Delta

## ADDED Requirements

### Requirement: Cleanup removes read-only subtrees
When removing the project data directory, the projects directory, or the cache directory, cleanup SHALL also remove subtrees whose directories lack write permission, such as a Go module cache left in Claude's temp root. If a removal fails with a permission error, cleanup SHALL grant the owner write permission on the directories in the tree and retry once.

#### Scenario: Read-only tree in project data
- **WHEN** `~/.asylum/projects/<container-name>/tmp` contains a directory with mode 0555 holding files with mode 0444
- **AND** `asylum cleanup` is run for the project
- **THEN** the project data directory is removed completely and cleanup reports success

#### Scenario: Retry still fails
- **WHEN** removal still fails after the permissions are granted
- **THEN** cleanup reports the error as it does today
