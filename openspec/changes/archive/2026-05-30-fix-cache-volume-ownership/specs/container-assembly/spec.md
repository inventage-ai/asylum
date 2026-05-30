## ADDED Requirements

### Requirement: Cache volume ownership normalization

After `docker run` succeeds and the container is ready, the system SHALL chown each cache volume mount point to the host `UID:GID` by running `docker exec --user root chown <uid>:<gid> <mountpoint>` for every entry in the aggregated cache directories map. This SHALL run on fresh container starts only, mirroring the existing shadow `node_modules` chown.

#### Scenario: Cache mountpoints chowned on fresh container start
- **WHEN** a fresh container is started with the `java/gradle` and `java/maven` kits active
- **THEN** `docker exec --user root chown <uid>:<gid> ~/.gradle` and `docker exec --user root chown <uid>:<gid> ~/.m2` SHALL be invoked after the container is ready

#### Scenario: No cache kits active
- **WHEN** a fresh container is started with no kits that contribute cache directories
- **THEN** no cache chown commands SHALL be invoked

#### Scenario: Existing container restart
- **WHEN** an existing container is reused (no fresh `RunDetached`)
- **THEN** no cache chown commands SHALL be invoked

#### Scenario: Idempotent on already-owned mountpoint
- **WHEN** a cache volume mountpoint is already owned by the host UID/GID (e.g. seeded from the image)
- **THEN** the chown SHALL still run and SHALL succeed as a no-op
