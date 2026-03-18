## Context

`.tool-versions` is a simple line-based format: `<tool> <version>`. It's used by mise and asdf. When a project has `.tool-versions` with `java 21.0.2`, asylum should treat that as `versions.java: 21.0.2` so the project image gets the exact version installed.

## Goals / Non-Goals

**Goals:**
- Read Java version from `.tool-versions` and use it as `versions.java`
- Lowest priority: asylum config files and `--java` flag override it
- No new dependencies — `.tool-versions` is trivial to parse

**Non-Goals:**
- Reading other tools from `.tool-versions` (Node.js is managed by fnm, Python by uv)
- Reading `.mise.toml` — different format, can add later if needed
- Supporting `.tool-versions` outside the project directory

## Decisions

### Parse in config.Load after all asylum config layers

Read `.tool-versions` from the project directory after merging all asylum config layers but before applying CLI flags. If `versions.java` is already set by any asylum config layer, `.tool-versions` doesn't override it. CLI flags still win over everything.

The priority order becomes: `.tool-versions` < `~/.asylum/config.yaml` < `.asylum` < `.asylum.local` < CLI flags.

### Only extract Java for now

Only the `java` line is relevant — Node.js and Python have their own version managers (fnm, uv) that asylum already handles separately. This keeps the scope minimal.

### Strip distribution prefix if present

`.tool-versions` may say `java temurin-21.0.2` (mise format) or `java 21.0.2` (plain). We pass whatever is there to `versions.java` and let the existing project Dockerfile generation handle it with `mise install java@<version>`.

## Risks / Trade-offs

- **Version format mismatch**: `.tool-versions` might use `temurin-21.0.2` while the pre-installed check expects `21`. This is fine — a specific patch version like `21.0.2` won't match `preinstalledJava["21"]`, so it triggers a project image build with the exact version. That's the correct behavior.
