# Design

## Context

`RunArgs` resolves all RunArgs and flattens them into the `docker run` argument list. `-e` args come from core, kits, agents and user config. #42 introduced `HostEnv` and `RunDetached(args, env)` to keep `env-from-host` values out of the arguments by placing them in the docker process environment.

## Goals / Non-Goals

**Goals:** no env value in `docker run` arguments or `--debug` output, for every source; unchanged precedence.

**Non-Goals:** `docker inspect` (inherent to Docker); `docker exec -e` session env (iTerm session metadata, not secrets).

## Decisions

- **Env file on stdin over process environment.** The #42 approach put values into the docker CLI's own environment, where the CLI acts on some names (`DOCKER_HOST`, proxies, `HOME`, ...); that needed a `DOCKER_*` exception and still left edge cases. Generalizing it to core/kit vars such as `XDG_RUNTIME_DIR` would make this worse. An env file read from `/dev/stdin` reaches only the container, never the CLI, and is not visible in `ps -E` either. A temporary file was rejected because it writes secrets to disk.
- **Split after resolving.** Dedup, priorities and overrides stay on RunArgs as before. Only flattening moves `-e NAME=value` args into the env file, which is placed after all other options, so it follows the project's `.env` `--env-file`. Docker lets a later env file override an earlier one, and the dedup already guarantees unique names. This was verified with the Docker CLI, for both `--rm` and `-d`.
- **Unrepresentable values stay inline.** Docker's env file parser reads lines of at most 64 KiB, requires UTF-8, rejects names with whitespace and treats a leading `#` as a comment. Such variables keep their `-e` argument rather than failing the start or leaking the value in a parser error. `env-from-host` is the exception: its values come from the host environment and exist to be kept off the command line, so an unrepresentable value aborts with an error naming only the variable. Supporting such values (e.g. via the docker process environment) is not worth the code for secrets that are 64 KiB or larger or not UTF-8.
- **`RunArgs` returns the env file content** as an extra value. The alternative, splitting the flat list in `main`, would have to locate the image position in a positional list.
- **Masking in `RunArg.String()`.** It is used only for display (debug table, overrides, conflict error), so masking there covers all three.

## Risks / Trade-offs

- [`/dev/stdin` must be readable by the docker CLI] → It works on Linux and macOS, the only supported hosts; `RunDetached` always supplies stdin.
- [Debug output no longer shows values, which makes value-related debugging harder] → Names and sources stay visible; values can be checked inside the container.
