## ADDED Requirements

### Requirement: Shadow node_modules volumes
During volume assembly, the system SHALL detect `node_modules` directories in the project and shadow them with named Docker volumes so host-built native binaries are not visible inside the container.

#### Scenario: Node.js project with node_modules
- **WHEN** the project has a `package.json` and a `node_modules` directory
- **THEN** `--mount type=volume,src=<named-volume>,dst=<node_modules_path>` is added after the project directory mount

#### Scenario: Non-Node.js project
- **WHEN** the project has no `package.json`
- **THEN** no shadow volume mounts are added

#### Scenario: Feature disabled via config
- **WHEN** config has `features: { shadow-node-modules: false }`
- **THEN** no shadow volume mounts are added regardless of project contents
