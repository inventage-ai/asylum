# Proposal

## Why

Asylum exposes the values of environment variables passed into the container (GitHub issue #40). `--debug` prints every `-e KEY=value` in plain text, including user secrets from `env:`/`-e` and asylum's own `ASYLUM_BROKER_TOKEN`. `docker run` also receives them as arguments, where other users on a Linux host and process-logging tools can read them. `env-from-host` (#42) already avoided the command line for its own values, but with a special case for `DOCKER_*` names.

## What Changes

- All `-e NAME=value` args of `docker run` are passed through an env file on stdin (`--env-file /dev/stdin`) instead. It comes after the project's `.env` file, so precedence is unchanged. Values an env file cannot hold (newlines, invalid UTF-8, 64 KiB or more, names with whitespace or a leading `#`) stay inline.
- `--debug` masks env values as `-e NAME=***`, in both the argument table and the overrides section. So does the conflict error for two args of the same priority.
- `env-from-host` passes its values like any other env var. Its bare `-e NAME` mechanism and the `DOCKER_*` exception are removed, since the docker CLI does not read env-file values into its own environment. Mapped values the env file cannot hold abort container creation instead of falling back to the command line.

## Capabilities

### New Capabilities

### Modified Capabilities
- `container-assembly`: environment values are passed to `docker run` via an env file on stdin.
- `runarg-pipeline`: debug output and conflict errors mask env values.
- `env-from-host`: the separate "values stay off the docker command line" requirement is removed, because container-assembly now covers it for all variables; values the env file cannot hold are rejected.

## Impact

- `internal/container`: `RunArgs` returns the env file content; `splitEnvArgs`; `HostEnv` is removed.
- `internal/docker`: `RunDetached(args, stdin)`.
- `internal/kit`: `RunArg.String()` masks env values.
- `cmd/asylum/main.go`, docs, CHANGELOG.
- Not addressed: `docker inspect` still shows the values, and `docker exec -e` session env (iTerm metadata) stays inline.
