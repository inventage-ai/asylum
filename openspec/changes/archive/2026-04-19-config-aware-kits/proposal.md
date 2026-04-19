## Why

Java-specific logic (version selection, `ASYLUM_JAVA_VERSION` env, `preinstalledJava` map) is scattered across `config.go`, `container.go`, `image.go`, and `main.go` — all generic infrastructure. When the java kit's configured versions were ignored (issue #26), the fix required changes in 5 files because the kit couldn't own its own behavior. If node or python ever need configurable versions, the same leak would repeat. Kits should be self-contained: given their config, they produce their own snippets, env vars, and image logic.

## What Changes

- Kit struct gains optional snippet generator functions (`DockerSnippetFunc`, `RulesSnippetFunc`) that receive `*KitConfig` and return a snippet. Static `DockerSnippet`/`RulesSnippet` strings remain as the fallback for kits that don't need config.
- Kit struct gains `EnvFunc(*KitConfig) map[string]string` for contributing container environment variables (replaces hardcoded `ASYLUM_JAVA_VERSION` in `container.go`).
- Kit struct gains `ProjectSnippetFunc(*KitConfig) string` for contributing to the project Dockerfile (replaces `javaVersion`/`javaVersions` params in `EnsureProject`).
- Snippet assembly functions (`AssembleDockerSnippets`, etc.) accept a config accessor so they can call snippet funcs with the right `KitConfig`.
- Java kit moves all version logic into its own file: `DockerSnippetFunc` generates the mise install command from configured versions, `EnvFunc` emits `ASYLUM_JAVA_VERSION`, `ProjectSnippetFunc` handles custom version installation.
- **BREAKING** (internal): `EnsureProject` signature drops `javaVersion` and `javaVersions` params. `ConfigureFunc` field removed from Kit.
- `config.go`: `JavaVersion()`, `JavaVersions()`, `setJavaVersion()` removed. Java version accessors move into the java kit.
- `container.go`: `ASYLUM_JAVA_VERSION` hardcoded env removed; replaced by generic `EnvFunc` collection from kits.

## Capabilities

### New Capabilities

- `kit-snippet-generation`: Kits can generate Docker/rules/entrypoint snippets dynamically from their KitConfig, replacing static strings when config-dependent behavior is needed.

### Modified Capabilities

- `image-build`: `EnsureProject` no longer takes java-specific parameters; project snippet generation is delegated to kits via `ProjectSnippetFunc`.
- `container-assembly`: Container env vars contributed by kits are collected via `EnvFunc` instead of hardcoded per-kit logic.
- `mise-java`: Java versions installed in the base image are driven by the kit's `DockerSnippetFunc` reading config, not hardcoded. `ASYLUM_JAVA_VERSION` is set via the kit's `EnvFunc`. Custom version installation in project image is handled by the kit's `ProjectSnippetFunc`.

## Impact

- `internal/kit/kit.go` — new fields on Kit struct, updated assembly functions
- `internal/kit/java.go` — rewritten to use snippet funcs, owns all java version logic
- `internal/image/image.go` — `EnsureProject` simplified, no java params
- `internal/container/container.go` — generic kit env collection replaces hardcoded java env
- `internal/config/config.go` — `JavaVersion()`, `JavaVersions()`, `setJavaVersion()` removed
- `cmd/asylum/main.go` — `ConfigureFunc` loop removed, snippet assembly calls updated
- Tests in all affected packages
