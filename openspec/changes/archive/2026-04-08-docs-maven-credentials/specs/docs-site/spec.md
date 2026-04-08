## ADDED Requirements

### Requirement: Maven credentials documented
The `docs/kits/java.md` page SHALL document the Maven credential injection feature under `java/maven`, covering auto mode, explicit mode, disabling, and the rebuild requirement.

#### Scenario: User configures auto mode
- **WHEN** a user reads `docs/kits/java.md`
- **THEN** they SHALL find a `credentials: auto` example showing how to enable automatic pom.xml-based credential injection

#### Scenario: User configures explicit mode
- **WHEN** a user reads `docs/kits/java.md`
- **THEN** they SHALL find an explicit mode example showing how to specify server IDs directly

#### Scenario: User learns about rebuild requirement
- **WHEN** a user reads `docs/kits/java.md`
- **THEN** they SHALL find that credential changes require a container restart and that `--rebuild` applies them to a running container
