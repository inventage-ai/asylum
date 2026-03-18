## Tasks

- [x] **Rewrite parseArgs** — Replace heuristic passthrough with strict parsing. Return `(cliFlags, subcommand string, extraArgs []string, error)`. Unknown flags produce errors. `--` collects remaining args. `run` subcommand swallows everything after it. `shell` explicitly handles `--admin`.
- [x] **Simplify resolveMode** — Receive subcommand string directly instead of inferring from positional args. Map subcommands to container modes.
- [x] **Update printUsage** — Reflect new grammar with `run` subcommand and `--` separator.
- [x] **Update parseArgs tests** — Rewrite test cases for strict parsing: unknown flags error, `--` passthrough, `run` subcommand, `shell --admin`, and edge cases.
