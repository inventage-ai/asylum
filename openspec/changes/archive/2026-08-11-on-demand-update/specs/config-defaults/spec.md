## MODIFIED Requirements

### Requirement: Default config content
The default config SHALL be assembled from a header template, kit-provided ConfigSnippets (for active kits) and comments (for opt-in kits), and a footer template. The header SHALL document the `version-check-interval` field as a commented example, showing the default of `24h`, so users can discover and adjust the background version-refresh window.

#### Scenario: Default kits present
- **WHEN** the default config is examined
- **THEN** it contains active kits for java (versions 17, 21, 25; default 21), python, and node with their default settings, assembled from each kit's ConfigSnippet

#### Scenario: Optional sections commented out
- **WHEN** the default config is examined
- **THEN** optional agents (gemini, codex, opencode), optional kits (apt, shell, title), ports, volumes, and env sections are present but commented out with explanatory comments

#### Scenario: Version check interval documented
- **WHEN** the default config is examined
- **THEN** a commented `version-check-interval` entry is present with an explanatory comment and the `24h` default
