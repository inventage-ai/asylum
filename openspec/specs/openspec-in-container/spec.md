# openspec-in-container Specification

## Purpose

Puts the `openspec` CLI in the container image, installed globally via npm, so an agent working on a spec-driven project can run it without a per-project install.

## Requirements

### Requirement: OpenSpec CLI available in container
The Asylum container image SHALL have the `openspec` CLI installed globally via npm.

#### Scenario: OpenSpec available
- **WHEN** the container starts
- **THEN** `openspec --version` runs successfully
