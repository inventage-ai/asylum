# GitHub CLI Kit

GitHub CLI (`gh`) for interacting with GitHub from within containers.

**Default: On** — active unless explicitly disabled.

## What's Included

- **[gh](https://cli.github.com/)** — GitHub's official CLI for repos, PRs, issues, actions, and more

## Configuration

```yaml
kits:
  github:
    disabled: true    # disable this default-on kit
```

## Authentication

`gh` auth is part of your agent config, which is seeded from your host on first run and persisted in `~/.asylum/agents/<agent>/`. If you're already authenticated with `gh` on your host, it should work inside the container.

To authenticate manually inside a container:

```sh
asylum shell
gh auth login
```
