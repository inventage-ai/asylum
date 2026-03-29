## ADDED Requirements

### Requirement: MkDocs configuration
The project root SHALL contain a `mkdocs.yml` that configures MkDocs Material with site name, theme, navigation structure, and search. The navigation SHALL include sections for Getting Started, Commands, Configuration, Kits, Concepts, and Development.

#### Scenario: Valid MkDocs config
- **WHEN** `mkdocs build` is run from the project root
- **THEN** the site builds successfully with no errors

#### Scenario: Navigation structure
- **WHEN** the docs site is loaded
- **THEN** the sidebar shows sections: Getting Started, Commands, Configuration, Kits, Concepts, Development

### Requirement: Docs content pages
The `docs/` directory SHALL contain markdown pages for each section. Each command, kit, and concept SHALL have its own page.

#### Scenario: Command pages
- **WHEN** a user navigates to the Commands section
- **THEN** pages exist for shell, run, cleanup, version, ssh-init, and self-update

#### Scenario: Kit pages
- **WHEN** a user navigates to the Kits section
- **THEN** pages exist for node, python, java, docker, github, openspec, shell, and apt

#### Scenario: Concept pages
- **WHEN** a user navigates to the Concepts section
- **THEN** pages exist for images, mounts, sessions, and agents

#### Scenario: Kit page content
- **WHEN** a user reads a kit page
- **THEN** it SHALL include: what the kit provides, how to enable/configure it, any sub-kits, config options, and relevant examples

### Requirement: GitHub Pages deployment
A GitHub Actions workflow at `.github/workflows/docs.yml` SHALL build and deploy the docs site to GitHub Pages on push to main.

#### Scenario: Docs deploy on push
- **WHEN** a commit is pushed to main that changes files in `docs/` or `mkdocs.yml`
- **THEN** the workflow builds the site and deploys it to the `gh-pages` branch

#### Scenario: No deploy on unrelated changes
- **WHEN** a commit is pushed to main that does not change `docs/` or `mkdocs.yml`
- **THEN** the docs workflow does not run

### Requirement: Getting started page
The `docs/getting-started.md` page SHALL walk through first run: what happens when Asylum builds images, how agent config is seeded, and how to verify the setup is working.

#### Scenario: First run explained
- **WHEN** a user reads the getting started page
- **THEN** they understand the base image build, agent config seeding, and how to start their first session
