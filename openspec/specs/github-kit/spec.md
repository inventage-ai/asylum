# github-kit Specification

## Purpose

Installs the GitHub CLI (`gh`) from the official apt repository, so an agent can work with issues, pull requests and releases from inside the sandbox. Host authentication is supplied separately by [github-kit-credentials](../github-kit-credentials/spec.md).

## Requirements

### Requirement: GitHub CLI kit
The system SHALL provide a `github` kit that installs the GitHub CLI (`gh`) via the official apt repository. The kit SHALL be default-on.

#### Scenario: GitHub kit active
- **WHEN** the github kit is active
- **THEN** the `gh` CLI is available in the container

#### Scenario: GitHub kit disabled
- **WHEN** the github kit is disabled
- **THEN** the `gh` CLI is not installed
