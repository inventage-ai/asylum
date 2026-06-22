## Why

The `openspec` kit installs the OpenSpec CLI into containers but never initializes a project — the in-container agent has zero guidance, so it improvises `openspec init` with guessed defaults. Worse, our preferred workflow set (the experimental `custom` profile with `verify` instead of `sync`) lives in a global `config.json` whose location and effect are opaque even to experienced users. Users have complained that getting OpenSpec set up correctly is unclear and undocumented.

## What Changes

- Add an `asylum-openspec-init` script on the container PATH that initializes OpenSpec non-interactively with our preferred settings: it maps the active agent to the OpenSpec `--tools` id, runs `openspec init` for a fresh project or `openspec update --force` for an existing one.
- Seed the preferred OpenSpec global config (`{"profile":"custom","workflows":["propose","explore","apply","verify","archive"]}`) into the image at build time, so `openspec init` materializes the `verify` workflow command/skill files (and omits `sync`) deterministically.
- Set an `ASYLUM_AGENT` environment variable in every container, carrying the active agent's name, so the script can resolve the correct `--tools` id at runtime (the image is shared but the agent is chosen per run).
- Add a `RulesSnippet` to the `openspec` kit telling the agent to run `asylum-openspec-init` when the user wants OpenSpec and `openspec/` is absent.
- Make the `openspec` kit default-on (`TierDefault`), aligning the code with the existing spec and the docs.
- Document the setup recipe (`! asylum-openspec-init`) in `docs/kits/openspec.md`.

## Capabilities

### New Capabilities
- `openspec-init-script`: a container script that initializes/updates OpenSpec non-interactively, mapping the active agent to the OpenSpec `--tools` id and choosing init vs update based on whether `openspec/` already exists.

### Modified Capabilities
- `openspec-kit`: the kit additionally seeds the preferred OpenSpec global config, ships the `asylum-openspec-init` script, provides a `RulesSnippet`, and is default-on.
- `container-assembly`: the container sets an `ASYLUM_AGENT` env var carrying the active agent's name.

## Impact

- `internal/kit/openspec.go` — DockerSnippet (global config seed + script install), RulesSnippet, Tier.
- `internal/container/container.go` — set `ASYLUM_AGENT` from the active agent's name.
- `docs/kits/openspec.md` — setup recipe.
- `CHANGELOG.md` — Added/Changed entries.
- No new dependencies. The script is plain bash; agent→tools mapping only needs to translate `copilot` → `github-copilot`.
