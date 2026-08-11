# update

Refresh the cached agent versions and rebuild the image if any of them changed.

## Usage

```
asylum update
```

## Description

Asylum pins each coding agent's CLI to a specific version, baked into the base image as a build ARG. Those versions are cached in `~/.asylum/versions.json` and refreshed in the background at most once every 24 hours (see [`version-check-interval`](../configuration/index.md#top-level-fields)).

`asylum update` forces that refresh immediately: it fetches every agent's latest version regardless of how fresh the cache is, writes the result, and then ensures the base and project images are up to date. If no version changed, the image hashes match and nothing is rebuilt. If a version changed, the base image and its dependent project image are rebuilt.

The command never starts a container or launches an agent — it exits once the images are current.

## update vs self-update

| Command | Updates |
|---------|---------|
| `asylum update` | The agent CLIs inside the container image |
| [`asylum self-update`](self-update.md) | The Asylum binary itself |

## Examples

```sh
# Pick up a new Claude Code release without waiting for the 24h window
asylum update

# Then start a session on the fresh image
asylum
```

## Notes

- The fetch is blocking, unlike the normal background refresh, so images are rebuilt within the same invocation.
- If every version fetch fails (e.g. no network), `versions.json` is left untouched and nothing is rebuilt.
- The command takes no arguments or flags.
