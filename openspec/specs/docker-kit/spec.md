# docker-kit Specification

## Purpose

Installs a full Docker engine inside the container so an agent can build and run containers itself, not just talk to a socket. Because that needs `--privileged`, the flag is tied to this kit being active and is not granted otherwise. The Docker CLI stays in the core image either way, so `docker` against an external daemon works without the kit.

## Requirements

### Requirement: Docker kit definition
The system SHALL provide a `docker` kit that installs Docker engine (docker-ce, buildx, compose), adds the container user to the docker group, and starts the Docker daemon at container startup.

#### Scenario: Docker kit active
- **WHEN** the docker kit is present in the kits config
- **THEN** the base image includes Docker engine, buildx, and compose plugins, and the entrypoint starts the Docker daemon

#### Scenario: Docker kit inactive
- **WHEN** the docker kit is not present in the kits config
- **THEN** the base image does not include Docker engine, and no daemon startup occurs

### Requirement: Docker kit active by default
The docker kit SHALL be included in the default config generated on first run.

#### Scenario: Default config
- **WHEN** a new default config is generated
- **THEN** the docker kit is present under kits

### Requirement: Container privileged mode conditional on docker kit
The container SHALL only run with `--privileged` when the docker kit is active.

#### Scenario: Docker kit active
- **WHEN** the docker kit is active
- **THEN** the container runs with `--privileged` and `ASYLUM_DOCKER=1`

#### Scenario: Docker kit inactive
- **WHEN** the docker kit is not active
- **THEN** the container runs without `--privileged` and without `ASYLUM_DOCKER=1`

### Requirement: Docker CLI remains in core
The Docker CLI binary SHALL remain in the core Dockerfile regardless of whether the docker kit is active.

#### Scenario: Docker kit inactive but host socket mounted
- **WHEN** the docker kit is inactive and a host Docker socket is mounted
- **THEN** the `docker` CLI is available and can communicate with the host Docker daemon
