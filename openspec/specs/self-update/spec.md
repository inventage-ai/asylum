# self-update Specification

## Purpose

Replaces the running asylum binary with a newer one from GitHub Releases, choosing the asset for the detected OS and architecture. The target comes from the stable channel by default, from the `dev` pre-release with `--dev`, or from the configured `release-channel` when no flag is given. Replacement is atomic — a temporary file in the same directory, then a rename — so an interrupted update cannot leave a half-written binary.

## Requirements

### Requirement: Update to latest stable release
The `self-update` subcommand SHALL query the GitHub Releases API for the latest non-prerelease and download the matching binary for the current OS and architecture, replacing the running binary atomically. When a version argument is provided, the specified release tag SHALL be fetched instead.

#### Scenario: Successful stable update
- **WHEN** `asylum self-update` is run and a newer stable release exists
- **THEN** the binary is downloaded, replaces the current executable, and the new version is printed

#### Scenario: Already up to date
- **WHEN** `asylum self-update` is run and the running version matches the latest stable release
- **THEN** a message is printed indicating the binary is already up to date, and no download occurs

#### Scenario: Targeted version update
- **WHEN** `asylum self-update 0.4.0` is run
- **THEN** the release tagged `v0.4.0` is fetched and installed

#### Scenario: Targeted version with v prefix
- **WHEN** `asylum self-update v0.4.0` is run
- **THEN** the release tagged `v0.4.0` is fetched and installed (the `v` prefix is normalized)

#### Scenario: Already at targeted version
- **WHEN** `asylum self-update 0.4.0` is run and the current version is `0.4.0`
- **THEN** a message is printed indicating the binary is already at the requested version

#### Scenario: Version not found
- **WHEN** `asylum self-update 99.0.0` is run and no release with tag `v99.0.0` exists
- **THEN** an error is printed indicating the release was not found

### Requirement: Version and dev mutual exclusivity
The `--dev` flag and a version argument SHALL NOT be combined. If both are provided, the CLI SHALL exit with an error.

#### Scenario: Version and --dev conflict
- **WHEN** `asylum self-update --dev 0.4.0` is run
- **THEN** the CLI exits with an error indicating that `--dev` and a version argument cannot be combined

### Requirement: Dev channel update
The `self-update` subcommand SHALL accept a `--dev` flag that targets the `dev` pre-release instead of the latest stable release.

#### Scenario: Update to dev with flag
- **WHEN** `asylum self-update --dev` is run
- **THEN** the binary from the `dev` pre-release is downloaded and installed

#### Scenario: Dev channel always downloads
- **WHEN** `asylum self-update --dev` is run and the current version is already `dev`
- **THEN** the binary is re-downloaded (dev builds have no meaningful version to compare)

### Requirement: Channel resolution from config
When the `--dev` flag is not provided, the subcommand SHALL check the `release-channel` config value. If set to `dev`, the dev channel is used. Otherwise, stable is used.

#### Scenario: Config sets dev channel
- **WHEN** `asylum self-update` is run without `--dev` and config has `release-channel: dev`
- **THEN** the dev channel is used

#### Scenario: No flag and no config defaults to stable
- **WHEN** `asylum self-update` is run without `--dev` and no `release-channel` is configured
- **THEN** the stable channel is used

#### Scenario: Flag overrides config
- **WHEN** `asylum self-update --dev` is run and config has `release-channel: stable`
- **THEN** the dev channel is used (flag wins)

### Requirement: Atomic binary replacement
The update SHALL write to a temporary file in the same directory as the current binary and use an atomic rename to replace it. A failed download SHALL NOT corrupt the existing binary.

#### Scenario: Download failure
- **WHEN** the download fails mid-transfer
- **THEN** the existing binary is unchanged and an error is printed

#### Scenario: Permission denied
- **WHEN** the binary is in a root-owned directory and the user lacks write permission
- **THEN** an error message is printed suggesting `sudo asylum self-update`

### Requirement: Platform detection
The subcommand SHALL detect the current OS (`linux`/`darwin`) and architecture (`amd64`/`arm64`) at runtime to select the correct binary asset from the release.

#### Scenario: Correct asset selected
- **WHEN** running on darwin/arm64
- **THEN** the asset named `asylum-darwin-arm64` is downloaded
