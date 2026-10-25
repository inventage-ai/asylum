# Spec Delta

## ADDED Requirements

### Requirement: Environment values passed via stdin env file
The system SHALL pass environment variables to `docker run` through an env file read from stdin (`--env-file /dev/stdin`) instead of `-e NAME=value` arguments, so their values do not appear in the process arguments. This env file SHALL come after the project's `.env` env file, so asylum's values keep precedence.

#### Scenario: Secret not in process arguments
- **WHEN** the config contains `env: {API_KEY: secret}`
- **THEN** the `docker run` arguments do not contain `secret`, and the container has `API_KEY=secret`

#### Scenario: Asylum values override the project .env
- **WHEN** the project's `.env` sets `FOO=a` and the config sets `env: {FOO: b}`
- **THEN** the container has `FOO=b`

### Requirement: Unrepresentable variables stay inline
A variable the env file format cannot represent SHALL be passed as an `-e` argument instead: a value with a newline or invalid UTF-8, a line of 64 KiB or more, or a name with whitespace or a leading `#`.

#### Scenario: Oversized value
- **WHEN** a variable's `NAME=value` line is 70000 bytes long
- **THEN** it is passed as `-e` and the container starts with the full value
