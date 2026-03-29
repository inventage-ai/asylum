# GitHub CLI Kit

GitHub CLI (`gh`) for interacting with GitHub from within containers.

**Activation: Default** — added to config on first detection; active when present.

## What's Included

- **[gh](https://cli.github.com/)** — GitHub's official CLI for repos, PRs, issues, actions, and more

## Configuration

```yaml
kits:
  github:
    credentials: auto   # share gh authentication from host
    disabled: true       # disable this default-on kit
```

## Authentication

Enable `credentials: auto` to extract your `gh` auth token from the host (including system keyrings) and generate a `hosts.yml` inside the container. This lets `gh` authenticate without requiring `gh auth login` in every new container.

Without credentials enabled, `gh` auth is part of your agent config, which is seeded from your host on first run and persisted in `~/.asylum/agents/<agent>/`.

To authenticate manually inside a container:

```sh
asylum shell
gh auth login
```
