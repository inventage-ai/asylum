## ADDED Requirements

### Requirement: browser kit registration
The system SHALL register a `browser` kit via `init()` in `internal/kit/browser.go` with name `"browser"`, TierOptIn, and a dependency on the `node` kit.

#### Scenario: Kit is registered at startup
- **WHEN** the application starts
- **THEN** the kit registry contains a `"browser"` entry with Tier set to TierOptIn and Deps containing `"node"`

### Requirement: Chromium installation via Playwright
The kit SHALL provide a DockerSnippet that installs the Playwright npm package and then runs `npx playwright install --with-deps chromium` to install Chromium along with all required system libraries.

#### Scenario: Chromium installed in image
- **WHEN** the browser kit is active and the Docker image is built
- **THEN** Chromium is installed and launchable via Playwright inside the container

#### Scenario: System dependencies installed
- **WHEN** the DockerSnippet executes the Playwright install command with `--with-deps`
- **THEN** all required system libraries (fonts, graphics, etc.) are installed automatically

### Requirement: Playwright cache directory
The kit SHALL declare a CacheDirs entry mapping `"playwright"` to `/home/claude/.cache/ms-playwright` so the Chromium binary is persisted in a named Docker volume.

#### Scenario: Cache volume mounted
- **WHEN** the container is created with the browser kit active
- **THEN** a named volume is mounted at `/home/claude/.cache/ms-playwright`

### Requirement: browser config snippet
The kit SHALL provide a ConfigSnippet and ConfigNodes so that kit sync can add a `browser` entry to the user's config file.

#### Scenario: Config entry added during kit sync
- **WHEN** kit sync detects browser as a new kit
- **THEN** a `browser:` entry with a descriptive comment is added to the kits section of `config.yaml`

### Requirement: browser tools metadata
The kit SHALL declare `Tools: []string{"playwright"}` so the tool is listed in aggregated tool output.

#### Scenario: Tool listed in aggregated tools
- **WHEN** `AggregateTools` is called with active kits including browser
- **THEN** the result contains `"playwright (browser)"`

### Requirement: browser banner line
The kit SHALL provide a BannerLines entry that prints Chromium version information in the welcome banner.

#### Scenario: Version shown in banner
- **WHEN** the container starts with browser kit active
- **THEN** the welcome banner includes a line showing the Chromium version

### Requirement: browser rules snippet
The kit SHALL provide a RulesSnippet describing browser capabilities for agents, including that Chromium is available via Playwright.

#### Scenario: Rules file contains browser section
- **WHEN** sandbox rules are assembled with browser kit active
- **THEN** the rules file contains a section describing Playwright browser automation capabilities
