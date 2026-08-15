# volume-shorthand Specification

## Purpose

Accepts a shorthand volume form alongside full Docker syntax, where naming a single path mounts it at the same path inside the container. Because asylum mounts the project at its real host path, host and container paths usually match, and repeating them in config is noise.

## Requirements

### Requirement: Volume shorthand parsing
The volume parser SHALL support both standard Docker volume syntax and shorthand forms where the container path equals the host path.

#### Scenario: Standard syntax
- **WHEN** volume is `/host/path:/container/path`
- **THEN** host is `/host/path`, container is `/container/path`, no options

#### Scenario: Standard syntax with options
- **WHEN** volume is `/host/path:/container/path:ro`
- **THEN** host is `/host/path`, container is `/container/path`, options is `ro`

#### Scenario: Shorthand single path
- **WHEN** volume is `/data`
- **THEN** host is `/data`, container is `/data`, no options

#### Scenario: Shorthand with mount option
- **WHEN** volume is `/data:ro`
- **THEN** host is `/data`, container is `/data`, options is `ro`

#### Scenario: Tilde expansion
- **WHEN** volume is `~/data:/data:ro` and home is `/home/user`
- **THEN** host is `/home/user/data`, container is `/data`, options is `ro`

#### Scenario: Tilde shorthand
- **WHEN** volume is `~/data` and home is `/home/user`
- **THEN** host is `/home/user/data`, container is `/home/user/data`, no options
