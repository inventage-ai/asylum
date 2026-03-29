## 1. MkDocs Setup

- [x] 1.1 Create `mkdocs.yml` at project root with Material theme, navigation structure, and search config
- [x] 1.2 Create `docs/` directory structure matching the design (index, getting-started, commands/, configuration/, kits/, concepts/, development/)

## 2. Docs Content — Getting Started

- [x] 2.1 Write `docs/index.md` — elevator pitch, install, quick start (concise landing for the docs site)
- [x] 2.2 Write `docs/getting-started.md` — first run walkthrough: image build, agent seeding, starting a session

## 3. Docs Content — Commands

- [x] 3.1 Write `docs/commands/index.md` — overview table of all commands
- [x] 3.2 Write individual command pages: shell, run, cleanup, version, ssh-init, self-update

## 4. Docs Content — Configuration

- [x] 4.1 Write `docs/configuration/index.md` — layered config, merge rules, .asylum/.asylum.local, example config
- [x] 4.2 Write `docs/configuration/flags.md` — CLI flags reference
- [x] 4.3 Write `docs/configuration/packages.md` — apt/npm/pip/run package installation

## 5. Docs Content — Kits

- [x] 5.1 Write `docs/kits/index.md` — what kits are, enable/disable, resolution logic, default-on kits, kit table
- [x] 5.2 Write kit pages: node, python, java, docker, github, openspec, shell, apt

## 6. Docs Content — Concepts & Development

- [x] 6.1 Write concept pages: images, mounts, sessions, agents
- [x] 6.2 Write `docs/development/building-from-source.md`

## 7. README Restructure

- [x] 7.1 Rewrite README.md as landing page: pitch, comparison table, what's included, install, quick start, config/commands/kits overviews with docs links, build from source, license

## 8. GitHub Pages Deployment

- [x] 8.1 Create `.github/workflows/docs.yml` — build and deploy on push to main (path-filtered to docs/ and mkdocs.yml)

## 9. Verification

- [x] 9.1 Run `mkdocs build` locally to verify the site builds with no errors
- [x] 9.2 Verify all navigation links resolve and no pages are orphaned
