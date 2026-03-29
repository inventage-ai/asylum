## MODIFIED Requirements

### Requirement: Sandbox rules file generation
The system SHALL generate a markdown rules file at `~/.asylum/projects/<container-name>/sandbox-rules.md` each time a new container is started. The file SHALL contain a core section, kit tools, kit snippets, and a section showing the project's allocated port range.

#### Scenario: Container start with allocated ports
- **WHEN** a container is started with ports 10000-10004 allocated
- **THEN** the rules file SHALL contain a "Forwarded Ports" section listing the ports and explaining they are accessible from the host at `http://localhost:<port>`

#### Scenario: Container start without ports kit
- **WHEN** the ports kit is disabled
- **THEN** the rules file SHALL NOT contain a "Forwarded Ports" section
