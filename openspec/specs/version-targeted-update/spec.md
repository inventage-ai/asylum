# version-targeted-update Specification

## Purpose

Lets `self-update` take a version argument to install a specific GitHub release rather than the newest one, which is what makes downgrading out of a bad release possible.

## Requirements

### Requirement: Install a specific version
The `self-update` subcommand SHALL accept an optional positional version argument that targets a specific GitHub release tag instead of the latest release.

#### Scenario: Update to a specific version
- **WHEN** `asylum self-update 0.4.0` is run
- **THEN** the release tagged `v0.4.0` is fetched and installed

#### Scenario: Version with v prefix
- **WHEN** `asylum self-update v0.4.0` is run
- **THEN** the release tagged `v0.4.0` is fetched and installed (the `v` prefix is normalized)

#### Scenario: Version not found
- **WHEN** `asylum self-update 99.0.0` is run and no release with tag `v99.0.0` exists
- **THEN** an error is printed indicating the release was not found

#### Scenario: Version and --dev are mutually exclusive
- **WHEN** `asylum self-update --dev 0.4.0` is run
- **THEN** the CLI exits with an error indicating that `--dev` and a version argument cannot be combined

#### Scenario: Already at requested version
- **WHEN** `asylum self-update 0.5.0` is run and the current version is already `0.5.0`
- **THEN** a message is printed indicating the binary is already at the requested version, and no download occurs
