## ADDED Requirements

### Requirement: Config-driven Docker snippet generation
A kit MAY provide a `DockerSnippetFunc` that receives the kit's `*SnippetConfig` and returns a Dockerfile snippet. When present, the assembly layer SHALL call it instead of using the static `DockerSnippet` string. When `*SnippetConfig` is nil (kit has no user config), the func SHALL still be called and MUST return a sensible default.

#### Scenario: Kit with DockerSnippetFunc and config
- **WHEN** a kit has `DockerSnippetFunc` set and the user has configured that kit
- **THEN** `AssembleDockerSnippets` SHALL call the func with the kit's `*SnippetConfig` and use the returned string

#### Scenario: Kit with DockerSnippetFunc and no config
- **WHEN** a kit has `DockerSnippetFunc` set but the user has no config for that kit
- **THEN** `AssembleDockerSnippets` SHALL call the func with nil `*SnippetConfig` and use the returned string

#### Scenario: Kit with static DockerSnippet only
- **WHEN** a kit has `DockerSnippet` set but no `DockerSnippetFunc`
- **THEN** `AssembleDockerSnippets` SHALL use the static string unchanged

#### Scenario: Kit with both func and static string
- **WHEN** a kit has both `DockerSnippetFunc` and `DockerSnippet` set
- **THEN** the func SHALL take precedence

### Requirement: Config-driven rules snippet generation
A kit MAY provide a `RulesSnippetFunc` that receives the kit's `*SnippetConfig` and returns a rules markdown fragment. When present, the assembly layer SHALL call it instead of using the static `RulesSnippet` string, following the same fallback semantics as `DockerSnippetFunc`.

#### Scenario: Kit with RulesSnippetFunc
- **WHEN** a kit has `RulesSnippetFunc` set
- **THEN** the rules assembly SHALL call the func and use the returned string

#### Scenario: Kit with static RulesSnippet only
- **WHEN** a kit has no `RulesSnippetFunc`
- **THEN** the static `RulesSnippet` string SHALL be used

### Requirement: Kit-contributed environment variables
A kit MAY provide an `EnvFunc` that receives the kit's `*SnippetConfig` and returns a `map[string]string` of environment variables to set on the container. The container assembly layer SHALL collect env vars from all kits with an `EnvFunc` and include them as run args.

#### Scenario: Kit provides env vars
- **WHEN** a kit has `EnvFunc` and returns `{"ASYLUM_JAVA_VERSION": "21"}`
- **THEN** the container run args SHALL include `-e ASYLUM_JAVA_VERSION=21`

#### Scenario: Kit provides no env vars
- **WHEN** a kit has `EnvFunc` that returns nil or an empty map
- **THEN** no additional env vars SHALL be added for that kit

#### Scenario: Kit has no EnvFunc
- **WHEN** a kit has no `EnvFunc` set
- **THEN** no env var collection is attempted for that kit

### Requirement: Kit-contributed project Dockerfile snippets
A kit MAY provide a `ProjectSnippetFunc` that receives the kit's `*SnippetConfig` and returns Dockerfile commands for the project image. The image assembly layer SHALL collect project snippets from all kits and include them in the generated project Dockerfile.

#### Scenario: Kit contributes project snippet
- **WHEN** a kit's `ProjectSnippetFunc` returns a non-empty string
- **THEN** the project Dockerfile SHALL include that snippet

#### Scenario: Kit contributes empty project snippet
- **WHEN** a kit's `ProjectSnippetFunc` returns an empty string
- **THEN** no commands are added for that kit

#### Scenario: No kits have ProjectSnippetFunc
- **WHEN** no active kits have `ProjectSnippetFunc`
- **THEN** the project Dockerfile generation SHALL behave as before (packages only)
