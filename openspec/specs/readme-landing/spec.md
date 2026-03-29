## ADDED Requirements

### Requirement: README structure
The README SHALL be restructured as a concise landing page (~120-150 lines) with sections: pitch, comparison table, what's included, install, quick start, configuration overview, commands overview, kits overview, building from source, and license. Each overview section SHALL link to the corresponding docs site page.

#### Scenario: README length
- **WHEN** the README is viewed
- **THEN** it is approximately 120-150 lines and self-contained enough to understand what Asylum does without visiting the docs site

#### Scenario: Docs links
- **WHEN** a user reads the configuration, commands, or kits sections
- **THEN** each section includes a link to the full docs site for detailed reference

### Requirement: Comparison table
The README SHALL include a feature comparison table covering Asylum, Claudebox, AgentBox, and Agent Safehouse. The table SHALL compare: approach, supported agents, platform support, language/runtime support, Docker-in-Docker, config system, and distribution method.

#### Scenario: Comparison present
- **WHEN** a user reads the README
- **THEN** a comparison table is visible near the top, after the pitch

#### Scenario: Neutral framing
- **WHEN** a user reads the comparison table
- **THEN** it presents a factual feature matrix with links to each tool's repository

### Requirement: What's included section
The README SHALL list out-of-the-box supported languages (Python, Node.js, Java), tools (Docker-in-Docker, gh, tmux, ripgrep, etc.), and plugins/kits (OpenSpec, GitHub CLI) so users can quickly assess whether Asylum covers their stack.

#### Scenario: Language and tool overview
- **WHEN** a user scans the README
- **THEN** they can see all supported languages, tools, and default-on kits without clicking through to the docs site
