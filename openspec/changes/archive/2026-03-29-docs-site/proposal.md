## Why

The README is currently the only documentation and is growing with every new kit. At 257 lines it's manageable today, but as kits multiply (9 kits with sub-kits already, each with config options, onboarding tasks, and behaviors), cramming everything into a single file is unsustainable. A structured docs site lets each kit, command, and concept have its own page while keeping the README focused as a landing page that sells the tool.

## What Changes

- **New MkDocs Material docs site** hosted on GitHub Pages with sections for getting started, commands, configuration, kits, concepts, and development.
- **README restructured** to a concise landing page: elevator pitch, comparison table (vs Claudebox, AgentBox, Safehouse), supported languages/tools overview, install, quick start, and links into the docs site.
- **GitHub Actions workflow** to build and deploy the docs site on push to main.
- **Detailed reference content** (commands, config, kits, concepts) moved from README into structured docs pages.

## Capabilities

### New Capabilities
- `docs-site`: MkDocs Material documentation site structure, configuration, GitHub Pages deployment workflow, and all content pages (getting started, commands, configuration, kits, concepts, development).
- `readme-landing`: Restructured README as a concise landing page with comparison table, feature highlights, and docs site links.

### Modified Capabilities

## Impact

- New `docs/` directory with MkDocs config and markdown pages
- New `mkdocs.yml` at project root
- New `.github/workflows/docs.yml` for GitHub Pages deployment
- `README.md` rewritten (shorter, with links to docs)
- New Python dev dependency (MkDocs Material) for building docs locally — not a runtime dependency of Asylum itself
