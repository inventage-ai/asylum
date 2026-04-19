## Context

Kits are registered as static structs in `init()` functions. Their `DockerSnippet`, `RulesSnippet`, and `EntrypointSnippet` are string constants. This works for most kits but breaks down when snippet content depends on user config (e.g., which Java versions to install). The current workaround (`ConfigureFunc`) mutates global kit state after config is loaded — fragile, timing-dependent, and forces java-specific parameters into generic APIs (`EnsureProject`, container env assembly).

The kit struct already has function-based hooks for runtime behavior (`ContainerFunc`, `CredentialFunc`, `MountFunc`). Extending this pattern to snippet generation is natural.

## Goals / Non-Goals

**Goals:**
- Kits can generate snippets dynamically from their `KitConfig`
- All java-specific logic lives in `java.go` — zero java knowledge in generic code
- The pattern works for any future kit that needs config-driven behavior (node versions, python versions, etc.)
- Static snippet strings continue to work unchanged for kits that don't need config

**Non-Goals:**
- Changing how kits are registered or resolved
- Changing the `KitConfig` struct or config merging
- Making all kits use dynamic snippets — only kits that need it
- Changing the two-tier image build model (base + project)

## Decisions

### Snippet func fields with static fallback

Add optional function fields alongside existing string fields:

```go
DockerSnippetFunc  func(*SnippetConfig) string
RulesSnippetFunc   func(*SnippetConfig) string
```

`SnippetConfig` is defined in the `kit` package (not `config.KitConfig`) to avoid an import cycle — `config` imports `kit`, so `kit` cannot import `config`. The struct contains only the fields snippet funcs need (`Versions`, `DefaultVersion`). A `Config.KitSnippetConfig(name)` method in the config package maps from `KitConfig` to `kit.SnippetConfig`.

Assembly functions check the func first, fall back to the string. This means:
- Existing kits need zero changes (their static strings keep working)
- A kit that needs config just sets the func, and can leave the string as a default for when config is nil

**Why funcs on Kit, not a method?** Kit is a struct, not an interface. Adding methods would require an interface or embedding — overkill for an optional hook. The func field pattern is already established (`ContainerFunc`, `CredentialFunc`, `MountFunc`).

**Why not replace strings with funcs everywhere?** Most kits have constant snippets. Forcing `func(*SnippetConfig) string { return "..." }` wrappers on 15+ kits would be churn for no benefit.

### `EnvFunc` for kit-contributed environment variables

Add `EnvFunc func(*SnippetConfig) map[string]string` to Kit. Container assembly collects env vars from all kits with an `EnvFunc` and merges them into the run args. This replaces the hardcoded `ASYLUM_JAVA_VERSION` in `container.go`.

The java kit's `EnvFunc` returns `{"ASYLUM_JAVA_VERSION": defaultVersion}` when a default version is configured.

### `ProjectSnippetFunc` for project-image contributions

Add `ProjectSnippetFunc func(*SnippetConfig) string` to Kit. When set, `EnsureProject` calls it to get Dockerfile commands for the project image. The java kit uses this to install a non-preinstalled default version via mise.

This replaces the `javaVersion` and `javaVersions` parameters on `EnsureProject` and the `preinstalledJava` map.

### Config accessor passed through assembly

`AssembleDockerSnippets` (and similar assembly functions) currently take `[]*Kit`. They need access to each kit's config to call snippet funcs. Options:

1. Pass the full `Config` and let assembly look up kit config
2. Pass a `func(kitName string) *SnippetConfig` accessor
3. Pre-bind config into each Kit before assembly

Option 2 is cleanest — it avoids coupling assembly to `Config` and avoids mutating Kit state. The accessor is `cfg.KitSnippetConfig` (a new method that maps `KitConfig` → `SnippetConfig`).

### Java version accessors move into java.go

`config.JavaVersion()` and `config.JavaVersions()` are convenience methods that reach into the java kit's config. They exist because `image.go` and `container.go` needed java-specific values. Once those callers are gone, the accessors have no purpose. The java kit's snippet funcs read `sc.DefaultVersion` and `sc.Versions` directly from the `*SnippetConfig` they receive.

`setJavaVersion()` (used for `.tool-versions` parsing) stays in config but writes to `KitConfig.DefaultVersion` generically — it doesn't need a java-specific helper.

### Remove `ConfigureFunc`

`ConfigureFunc` was a single-use workaround for the java kit. With `DockerSnippetFunc` and friends, it's obsolete. The configure loop in `main.go` is removed.

## Risks / Trade-offs

**[Assembly function signatures change]** → `AssembleDockerSnippets` and similar gain a `func(string) *SnippetConfig` parameter. All callers must be updated. These are internal functions with a small number of callers.

**[`EnsureProject` signature changes]** → Drops `javaVersion` and `javaVersions`, gains a config accessor. Internal API only, no external consumers.

**[Java version defaults change path]** → Currently the hardcoded `DockerSnippet` in java.go contains the defaults (17, 21, 25). With `DockerSnippetFunc`, the defaults come from `ConfigSnippet` (the default config seeded on first run). If a user has no java config at all and `SnippetConfig` is nil, the `DockerSnippetFunc` must fall back to sane defaults. The func receives nil `*SnippetConfig` in this case and should use the same defaults as the current static snippet.

**[`.tool-versions` integration]** → Currently `readToolVersionsJava` in config.go calls `setJavaVersion`. This needs to continue working — it sets `KitConfig.DefaultVersion` which the java kit's `EnvFunc` and `ProjectSnippetFunc` will read. No change needed to this flow, but it must be verified.
