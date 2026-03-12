## Context

PLAN.md sections 4.1–4.5 define the config format, layering, and merge semantics. The config struct mirrors the YAML schema exactly.

## Goals / Non-Goals

**Goals:**
- Parse the YAML config format from PLAN.md section 4.3
- Load and merge three config files in priority order
- Provide a `Config` struct that downstream packages consume
- Volume shorthand parsing as a standalone function (used by container package)
- CLI flags struct that overlays on merged config

**Non-Goals:**
- No CLI flag parsing here — that belongs in `cmd/asylum/main.go`
- No validation of agent names or version values — consumers validate their own inputs

## Decisions

- **Config struct**: Flat struct with `Agent`, `Ports`, `Volumes`, `Versions`, `Packages` fields. `Versions` is `map[string]string`, `Packages` is `map[string][]string`.
- **CLIFlags struct**: Separate struct for CLI-only overrides. `Load()` accepts both the project directory and CLI flags, applying flags last.
- **Merge function**: A single `merge(base, overlay)` function that applies overlay on top of base following the three merge rules. Called twice (global → project → local).
- **Volume parsing**: `ParseVolume(raw, homeDir)` returns a `Volume` struct with `Host`, `Container`, `Options` fields. Handles tilde expansion, shorthand detection, and mount options.
- **File loading**: Missing config files are silently skipped (not errors). Only YAML parse errors are fatal.

## Risks / Trade-offs

- The merge logic is the trickiest part. Table-driven tests cover all the edge cases documented in PLAN.md section 4.4.
