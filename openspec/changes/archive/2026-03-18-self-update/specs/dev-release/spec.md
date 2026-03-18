## ADDED Requirements

### Requirement: Rolling dev release on push to main
A CI workflow SHALL build all four binary targets and publish them as a `dev` pre-release on every push to the `main` branch.

#### Scenario: Push to main triggers dev release
- **WHEN** a commit is pushed to `main`
- **THEN** the workflow builds `asylum-{linux,darwin}-{amd64,arm64}` with `VERSION=dev` and publishes them under the `dev` tag

#### Scenario: Previous dev release is replaced
- **WHEN** a dev release already exists
- **THEN** the existing `dev` tag and release are deleted before creating the new one

### Requirement: Dev release is a pre-release
The `dev` release SHALL be marked as a pre-release so that the GitHub API's `releases/latest` endpoint continues to return the latest tagged stable release.

#### Scenario: Latest endpoint returns stable
- **WHEN** a dev release exists alongside tagged stable releases
- **THEN** `GET /repos/{owner}/{repo}/releases/latest` returns the stable release, not dev
