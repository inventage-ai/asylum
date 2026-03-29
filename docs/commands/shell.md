# shell

Start an interactive zsh shell inside the container.

## Usage

```
asylum shell
asylum shell --admin
```

## Description

Opens a zsh shell in the running container (or starts a new container first). Useful for manual debugging, running tests, or exploring the container environment.

The `--admin` flag opens a root shell with a notice that you're running as root. This is needed for installing system packages or modifying files owned by root.

## Flags

| Flag | Description |
|------|-------------|
| `--admin` | Open a root shell instead of the default user shell |

## Examples

```sh
# Open a regular shell
asylum shell

# Open a root shell to install a system package
asylum shell --admin
apt-get install -y some-package

# Shell with port forwarding
asylum -p 3000 shell
```

## Notes

- Multiple shell sessions can share the same container. The container is removed when the last session exits.
- Global flags like `-p` and `-v` are applied when the container is first created, not when exec-ing into an existing one.
