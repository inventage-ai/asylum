# OpenSpec Kit

[OpenSpec](https://openspec.dev) CLI for structured change management.

**Activation: Default** — added to config on first detection; active when present. Depends on the [Node.js](node.md) kit.

## What's Included

- **openspec** — CLI for managing specs, changes, and proposals

## Configuration

```yaml
kits:
  openspec:
    disabled: true    # disable this default-on kit
```

## Dependencies

The OpenSpec kit depends on `node` because it's installed via npm. If the Node.js kit is not active, Asylum emits a warning.

## Setup

To set up OpenSpec in a project that doesn't have it yet, run the bundled init script:

```sh
! asylum-openspec-init
```

This initializes OpenSpec non-interactively with Asylum's preferred settings: the `custom` workflow profile (`propose`, `explore`, `apply`, `verify`, `archive` — `verify` instead of `sync`), wired up for whichever agent is running. It is safe to re-run — on an already-initialized project it refreshes the instruction files instead. The agent will also run it automatically when you ask to use OpenSpec in an uninitialized project.

## Usage

OpenSpec is used inside containers for structured change management:

```sh
openspec new change "add-feature"
openspec status --change "add-feature"
openspec list
```

See the [OpenSpec documentation](https://openspec.dev) for full usage details.

## Notes

- OpenSpec telemetry is disabled inside containers (`OPENSPEC_TELEMETRY=0`).
