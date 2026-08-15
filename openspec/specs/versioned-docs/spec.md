# versioned-docs Specification

## Purpose

Publishes the documentation site once per release rather than only from the tip of main, so a user on an older asylum can read the docs that match their binary. Versions are deployed with mike under their own URL prefixes, `dev` tracks main, release tags produce stable versions, and the root URL redirects to the newest stable one.

## Requirements

### Requirement: Version selector
The docs site SHALL display a version selector dropdown in the header that lists all deployed versions. The selector SHALL allow users to switch between versions without losing their current page context.

#### Scenario: Version dropdown visible
- **WHEN** a user visits any page on the docs site
- **THEN** a version selector dropdown is visible in the header

#### Scenario: Switch version preserves page
- **WHEN** a user is on `/0.6.0/kits/node/` and selects version `dev` from the dropdown
- **THEN** the browser navigates to `/dev/kits/node/`

### Requirement: Dev version deployment
The `dev` version SHALL be deployed on every push to `main` that changes `docs/` or `mkdocs.yml`. It SHALL always reflect the latest docs from the main branch.

#### Scenario: Dev docs updated on push
- **WHEN** a commit that modifies `docs/` is pushed to `main`
- **THEN** the `dev` version on the docs site is rebuilt and deployed

#### Scenario: Dev version overwritten
- **WHEN** a new push to main triggers a docs deployment
- **THEN** the `dev` version replaces the previous `dev` content (not appended as a new version)

### Requirement: Stable version deployment
A new stable docs version SHALL be deployed when a release tag (`v*`) is pushed. The version identifier SHALL be the tag name without the `v` prefix (e.g., tag `v0.6.0` deploys as `0.6.0`).

#### Scenario: Docs deployed on release tag
- **WHEN** a tag matching `v*` is pushed
- **THEN** the docs are deployed as a new version using the bare version number

#### Scenario: Latest alias updated
- **WHEN** a stable version is deployed
- **THEN** the `latest` alias is updated to point to that version

### Requirement: Root URL redirect
The root URL (`/`) SHALL redirect to the `latest` alias, which resolves to the most recent stable release. Before any stable version is deployed, the root URL SHALL redirect to `dev` as a fallback.

#### Scenario: Root redirects to latest
- **WHEN** a user navigates to `asylum.inventage.ai/` and at least one stable version has been deployed
- **THEN** they are redirected to the latest stable version's docs

#### Scenario: Root redirects to dev before first release
- **WHEN** a user navigates to `asylum.inventage.ai/` and no stable version has been deployed yet
- **THEN** they are redirected to the dev version's docs

### Requirement: Version URL structure
Each version SHALL be served from its own URL prefix. Versioned URLs SHALL follow the pattern `/<version>/` (e.g., `/dev/`, `/0.6.0/`).

#### Scenario: Version-prefixed URLs
- **WHEN** a user navigates to `/0.6.0/getting-started/`
- **THEN** they see the getting started page as it existed when version 0.6.0 was deployed

### Requirement: MkDocs version provider configuration
`mkdocs.yml` SHALL include `extra.version.provider: mike` to enable the Material theme's built-in version selector.

#### Scenario: Version provider configured
- **WHEN** `mkdocs.yml` is parsed
- **THEN** `extra.version.provider` is set to `mike`
