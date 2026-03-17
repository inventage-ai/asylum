## Context

Asylum's CLI already has a `version` variable (set to `"dev"` by default, overridden via `-ldflags` during release builds). The variable is used in image tagging but never exposed to the user. The flag parsing is manual (no framework), handling known flags in `parseArgs`.

## Goals / Non-Goals

**Goals:**
- Expose the existing `version` variable to users via `--version`
- Follow the same pattern as `--help` (print and exit, no container setup)

**Non-Goals:**
- Verbose version output (build date, commit hash, Go version) — keep it minimal
- Short `-V` alias — not standard enough to warrant adding

## Decisions

**Add `--version` to the `Flags` struct and `parseArgs`**: Same pattern as `--help`. Dispatch in `main()` right after the help check. Prints `asylum <version>` to stdout and exits.

**Output format**: Plain `asylum <version>` (e.g., `asylum v0.5.0` or `asylum dev`). No extra metadata — consistent with the project's minimalist approach.

## Risks / Trade-offs

No meaningful risks. This is a trivial, self-contained addition.
