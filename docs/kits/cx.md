# cx Kit

Semantic code navigation for AI agents via [cx](https://github.com/ind-igo/cx).

**Activation: Opt-in** — only active if explicitly enabled in your config.

## What's Included

- **cx** — CLI for semantic code navigation using tree-sitter grammars

## Configuration

```yaml
kits:
  cx:
    packages:            # tree-sitter language grammars to install
      - python
      - typescript
      - go
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `packages` | list | `[]` | Tree-sitter language grammars to install via `cx lang add` |

## Usage

```sh
# File overview (table of contents with symbols)
cx overview src/main.go

# Search for symbols across the project
cx symbols --query "parse"

# Jump to a function definition
cx definition --name parseArgs

# Find all references to a symbol
cx references --name Config

# Add support for a language
cx lang add python
```

cx uses tree-sitter for parsing, so language grammars must be installed for each language you want to navigate. Configure them in the `packages` list to have them installed automatically at image build time.
