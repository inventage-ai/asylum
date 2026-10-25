# Spec Delta

## MODIFIED Requirements

### Requirement: Same restrictions as literal env vars
Variables from `env-from-host` SHALL be subject to the same restrictions as `env`: the reserved active-agent variable SHALL be skipped with a warning, and a value containing a newline SHALL abort container creation with an error naming the variable. A variable the env file cannot represent unchanged (invalid UTF-8, a line of 64 KiB or more, an unsupported name) SHALL abort the same way instead of falling back to the command line.

#### Scenario: Value with newline
- **WHEN** the mapped host variable contains a newline
- **THEN** container creation fails with an error naming the container variable, without printing the value

#### Scenario: Value too large for the env file
- **WHEN** the mapped host variable is 70000 bytes long
- **THEN** container creation fails with an error naming the container variable, without printing the value

## REMOVED Requirements

### Requirement: Values stay off the docker command line
**Reason**: Generalized to all environment variables by the container-assembly requirement "Environment values passed via stdin env file", which also removes the need for the `DOCKER_*` exception.
**Migration**: None; mapped values are still kept off the `docker run` arguments.
