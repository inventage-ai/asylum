## Context

Asylum currently has a single README.md (257 lines) as its only documentation. The project has 9 kits (with sub-kits), 6 commands, a layered config system, and several architectural concepts — all growing steadily. The README tries to be both a landing page and a reference manual, which will become untenable as kits multiply.

## Goals / Non-Goals

**Goals:**
- Structured documentation site with per-kit, per-command, and per-concept pages
- README that works as a landing page: pitch, comparison, quick start, links to docs
- Automated deployment to GitHub Pages on push to main
- Easy to add a new page when adding a new kit or command

**Non-Goals:**
- Auto-generating docs from Go code or kit structs (future improvement)
- Versioned docs (tracking multiple Asylum releases) — just track main for now
- API documentation (Asylum has no programmatic API)
- Localization / i18n

## Decisions

### MkDocs Material over Hugo, Docusaurus, or mdBook

MkDocs Material is the closest match to the "readthedocs style" goal. It has built-in search, excellent sidebar navigation, and is the standard for CLI tool documentation (used by FastAPI, Pydantic, Poetry). The GitHub Actions deployment is ~15 lines of YAML via `mkdocs-gh-deploy`.

Alternatives considered:
- **Hugo**: Go-based (fits project ecosystem), but fewer docs-first themes and more config overhead.
- **Docusaurus**: Excellent docs UX, but requires Node.js and is heavier than needed.
- **mdBook**: Too sparse — limited navigation, no search, minimal styling.

MkDocs introduces a Python build dependency, but only for docs builds — not for Asylum itself. The CI workflow installs it in isolation.

### Site structure: kits/ as the growth directory

```
docs/
├── index.md                  ← Elevator pitch + quick start
├── getting-started.md        ← First run walkthrough
├── commands/
│   ├── index.md              ← Command overview table
│   ├── shell.md
│   ├── run.md
│   ├── cleanup.md
│   ├── version.md
│   ├── ssh-init.md
│   └── self-update.md
├── configuration/
│   ├── index.md              ← Layered config, merge rules
│   ├── flags.md              ← CLI flags reference
│   └── packages.md           ← apt/npm/pip/run packages
├── kits/
│   ├── index.md              ← What kits are, resolution, defaults
│   ├── node.md
│   ├── python.md
│   ├── java.md
│   ├── docker.md
│   ├── github.md
│   ├── openspec.md
│   ├── shell.md
│   └── apt.md
├── concepts/
│   ├── images.md             ← Two-tier image strategy
│   ├── mounts.md             ← What's mounted where
│   ├── sessions.md           ← Multi-session, resume, seeding
│   └── agents.md             ← Claude/Gemini/Codex specifics
└── development/
    └── building-from-source.md
```

Key principle: adding a new kit = adding one markdown file to `docs/kits/` and one line to the nav in `mkdocs.yml`. Everything else is stable.

### README as landing page with comparison table

The README will be restructured to ~120-150 lines covering:
1. One-paragraph pitch
2. Comparison table (vs Claudebox, AgentBox, Safehouse)
3. What's included (languages, tools, features)
4. Install
5. Quick start
6. Brief config overview with link to docs
7. Commands table with link to docs
8. Kits table with link to docs
9. Building from source
10. License

The comparison table highlights Asylum's key differentiators: multi-agent support, Go binary distribution, Docker-in-Docker, real host path mounts, layered YAML config, and the kit system.

### GitHub Actions workflow for docs deployment

A dedicated `.github/workflows/docs.yml` workflow triggers on push to main (when docs/ or mkdocs.yml change) and uses the official `squidfunk/mkdocs-material` action to build and deploy to the `gh-pages` branch. This avoids needing Python installed in CI.

## Risks / Trade-offs

- **Content duplication**: Some README content will overlap with docs pages (e.g., install instructions). Mitigation: README has the minimal version, docs have the detailed version. Accept minor duplication for README self-containedness.
- **Stale comparison table**: Competitor tools evolve. Mitigation: frame the comparison as a neutral feature matrix, not claims about competitors. Include links to each tool so readers can verify.
- **MkDocs Material updates**: The theme is actively maintained but could change. Mitigation: pin the version in the workflow. Low risk — it's been stable for years.
- **Docs falling behind code**: New kits or commands could be added without docs. Mitigation: the CLAUDE.md or PR template could remind contributors, but enforcement is manual for now.
