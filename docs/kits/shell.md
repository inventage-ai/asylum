# Shell Kit

Shell configuration: oh-my-zsh, tmux, and direnv.

**Default: On** — active unless explicitly disabled.

## What's Included

- **[oh-my-zsh](https://ohmyz.sh/)** with the `robbyrussell` theme
- **[tmux](https://github.com/tmux/tmux)** with a pre-configured setup
- **[direnv](https://direnv.net/)** hooks in both bash and zsh
- PATH configuration for fnm and mise in both shells

## Configuration

```yaml
kits:
  shell:
    disabled: true           # disable this default-on kit
    build:                   # custom commands to run at image build time
      - "curl -fsSL https://deno.land/install.sh | sh"
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `disabled` | bool | `false` | Disable this kit |
| `build` | list | `[]` | Custom shell commands to run during project image build |

## tmux Configuration

The pre-installed tmux config includes:

- Mouse support enabled
- 256-color terminal
- 50,000 line scroll history
- `|` to split horizontally, `-` to split vertically
- Status bar: hostname (left), date/time (right)

## Custom Build Commands

The `build` option runs arbitrary commands during the project image build. This is the escape hatch for installing tools that don't fit into other kit categories:

```yaml
kits:
  shell:
    build:
      - "curl -fsSL https://deno.land/install.sh | sh"
      - "go install golang.org/x/tools/gopls@latest"
```

Commands run as the `claude` user. See [Packages](../configuration/packages.md) for more details.
