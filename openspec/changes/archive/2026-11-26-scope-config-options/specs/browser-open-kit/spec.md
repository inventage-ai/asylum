# Spec Delta

## MODIFIED Requirements

### Requirement: Configurable extra URL schemes
The `browser-open` kit SHALL accept a `schemes` list in its kit configuration naming additional URL schemes the open path may hand to the host. Entries SHALL be normalized case-insensitively and SHALL tolerate the forms users copy from a tool's documentation (`dropshare5`, `dropshare5:`, `dropshare5://`, `dropshare5:///`). An entry that is not a syntactically valid URL scheme (a letter followed by letters, digits, `+`, `-`, or `.`) SHALL be reported as a warning and ignored, leaving the remaining entries in effect. The list SHALL have scope `personal` (see `config-layer-scopes`) and SHALL accumulate across the global and local layers, so `.asylum.local` adds to — rather than replaces — schemes allowed in the user's global configuration. Entries in the project layer (`.asylum`) SHALL be ignored and reported as rejected options. `http` and `https` SHALL always be allowed and need not be listed.

#### Scenario: Extra scheme configured
- **WHEN** the configuration contains `kits: {browser-open: {schemes: ["dropshare5"]}}`
- **THEN** the effective allowlist is `http`, `https`, and `dropshare5`

#### Scenario: Entry forms normalized
- **WHEN** an entry is written as `Dropshare5:///`, `DROPSHARE5:`, or `dropshare5://`
- **THEN** it is treated as the scheme `dropshare5`

#### Scenario: Layers accumulate
- **WHEN** the global configuration allows `dropshare5` and `.asylum.local` allows `vscode`
- **THEN** the effective allowlist contains both, plus `http` and `https`

#### Scenario: Project layer ignored
- **WHEN** the global configuration allows `dropshare5` and `.asylum` allows `vscode`
- **THEN** the effective allowlist is `http`, `https` and `dropshare5`, and `vscode` is reported as a rejected option

#### Scenario: Invalid entry ignored with a warning
- **WHEN** the list contains an entry that is not a valid scheme (e.g. `"1foo"` or `"a b"`) alongside a valid one
- **THEN** a warning names the rejected entry, the valid entry stays in effect, and the invalid entry is never accepted by the open path

#### Scenario: No configuration means no extra schemes
- **WHEN** no `schemes` list is configured
- **THEN** the effective allowlist is exactly `http` and `https`
