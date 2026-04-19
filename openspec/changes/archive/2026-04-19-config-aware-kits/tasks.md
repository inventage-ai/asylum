## 1. Kit struct and assembly

- [x] 1.1 Add `DockerSnippetFunc`, `RulesSnippetFunc`, `EnvFunc`, `ProjectSnippetFunc` fields to Kit struct in `kit.go`
- [x] 1.2 Remove `ConfigureFunc` field from Kit struct
- [x] 1.3 Update `AssembleDockerSnippets` to accept a `func(string) *SnippetConfig` accessor and call `DockerSnippetFunc` when present (fall back to static string)
- [x] 1.4 Update `AssembleRulesSnippets` to use `RulesSnippetFunc` with same fallback pattern
- [x] 1.5 Add `AssembleProjectSnippets` and `AssembleEnvVars` functions
- [x] 1.6 Update all assembly call sites in `main.go`, `image.go`, `order.go`, `container.go`, and integration tests to pass the config accessor

## 2. Java kit rewrite

- [x] 2.1 Replace static `DockerSnippet` with `DockerSnippetFunc` that generates mise install from `SnippetConfig.Versions` (default: 17, 21, 25)
- [x] 2.2 Replace static `RulesSnippet` with `RulesSnippetFunc` that generates rules from configured versions
- [x] 2.3 Add `EnvFunc` that returns `ASYLUM_JAVA_VERSION` from `SnippetConfig.DefaultVersion`
- [x] 2.4 Add `ProjectSnippetFunc` that returns mise install command when `DefaultVersion` is not in `Versions`

## 3. Remove java leaks from generic code

- [x] 3.1 Remove `JavaVersion()` from `config.go` (`setJavaVersion` stays for `.tool-versions` parsing; `JavaVersions()` didn't exist on main)
- [x] 3.2 Remove `javaVersion` params from `EnsureProject` in `image.go`; use `AssembleProjectSnippets` instead
- [x] 3.3 Remove hardcoded `ASYLUM_JAVA_VERSION` env from `container.go`; add generic kit `EnvFunc` collection loop
- [x] 3.4 Remove `ConfigureFunc` loop from `main.go` (didn't exist on main — already clean)

## 4. Tests

- [x] 4.1 Update `image_test.go`, `assembly_test.go`, `order_test.go` for new signatures
- [x] 4.2 Add tests for `DockerSnippetFunc` fallback behavior, `AssembleProjectSnippets`, `AssembleEnvVars`
- [x] 4.3 Add tests for java kit snippet generation (default versions, custom versions, project snippet for non-preinstalled version)
- [x] 4.4 Update container tests: `ASYLUM_JAVA_VERSION` now comes from kit `EnvFunc`, not hardcoded
- [x] 4.5 Run full test suite — 491 tests passing
