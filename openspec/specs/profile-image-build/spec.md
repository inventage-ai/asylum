# profile-image-build Specification

## Purpose

Covers the `USER_HOME` build argument that aligns the container user's home directory with the host's. The path participates in the base image hash, so moving to a machine with a different home directory triggers a rebuild rather than leaving stale paths baked into the image.

## Requirements

### Requirement: Base image assembly from global profiles
The base image Dockerfile SHALL be assembled with a `USER_HOME` build argument that sets the container user's home directory to match the host. The image hash SHALL include the home directory path.

#### Scenario: Build args include home directory
- **WHEN** the base image is built
- **THEN** build args include USER_ID, GROUP_ID, USERNAME, and USER_HOME from the host

#### Scenario: Home directory change triggers rebuild
- **WHEN** a user builds on a different machine with a different home directory
- **THEN** the base image hash changes and a rebuild is triggered
