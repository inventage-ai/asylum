## MODIFIED Requirements

### Requirement: Rolling dev release on push to main
A CI workflow SHALL build all four binary targets, sign and notarize darwin binaries, generate checksums, and publish them as a `dev` pre-release on every push to the `main` branch.

#### Scenario: Push to main triggers dev release
- **WHEN** a commit is pushed to `main`
- **THEN** the workflow builds `asylum-{linux,darwin}-{amd64,arm64}` with `VERSION=dev`, signs and notarizes darwin binaries, generates `checksums.txt`, and publishes them under the `dev` tag

#### Scenario: Previous dev release is replaced
- **WHEN** a dev release already exists
- **THEN** the existing `dev` tag and release are deleted before creating the new one
