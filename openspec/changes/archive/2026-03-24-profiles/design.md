## Context

Asylum currently installs all languages (Java, Python, Node.js) and their toolchains unconditionally in a monolithic Dockerfile. Language-specific concerns (cache directories, environment variables, entrypoint setup, onboarding tasks) are scattered across `container.go`, `entrypoint.sh`, `image.go`, and `main.go`. Adding a new language requires touching 5+ files with no shared abstraction.

The profile system introduces a single Go type that owns all aspects of a language's lifecycle within the container: what gets installed, how the environment is configured, what caches are mounted, what onboarding runs, and what entrypoint setup is needed.

## Goals / Non-Goals

**Goals:**
- Cohesive abstraction: all language-specific concerns live in one profile definition
- Hierarchical sub-profiles: `java` contains `maven` and `gradle`; users can select the whole tree or specific sub-profiles
- Two-tier integration: global profiles build into base image, project profiles into project image
- Full backwards compatibility: zero config changes needed for existing users
- Profiles hook into: image build, container setup (cache dirs, volumes, env), onboarding, entrypoint, config defaults

**Non-Goals:**
- User-defined profiles (filesystem-based, `~/.asylum/profiles/`) — future work
- Profile marketplace or sharing mechanism
- Dynamic profile discovery at runtime
- Changing the agent system (Claude, Gemini, Codex remain separate)
- Adding new languages beyond the current three (java, python, node) — the system supports it, but this change only migrates existing functionality

## Decisions

### 1. Profile as a Go struct, not an interface

Profiles are data-heavy (snippets, config defaults, cache maps) with no polymorphic behavior. A struct with fields is simpler than an interface with methods. The registry is a flat map of `*Profile` keyed by name.

```go
type Profile struct {
    Name              string
    Description       string
    DockerSnippet     string              // Dockerfile RUN instructions
    EntrypointSnippet string              // entrypoint.sh shell snippet
    CacheDirs         map[string]string   // name → container path
    Config            config.Config       // default config values
    OnboardingTasks   []onboarding.Task   // tasks to register
    SubProfiles       map[string]*Profile // child profiles
}
```

**Alternative considered**: Interface with methods (`DockerSnippet() string`, etc.). Rejected because there's no behavioral variation — every profile is the same shape with different data. A struct avoids boilerplate method implementations.

### 2. Hierarchical activation semantics

- `profiles: [java]` → activates `java` and all its sub-profiles (`maven`, `gradle`)
- `profiles: [java/maven]` → activates `java` (parent implied) + `maven` only
- `profiles: [java/maven, java/gradle]` → equivalent to `profiles: [java]`

Resolution algorithm:
1. Parse each profile string: split on `/` to get `(parent, child?)`
2. Look up parent in registry
3. If no child specified, activate parent + all sub-profiles
4. If child specified, activate parent + that sub-profile only
5. Deduplicate: a profile activated multiple ways is included once
6. Return a flat list of `*Profile` in deterministic order (parent before children)

**Alternative considered**: Flat profiles with dependency resolution (claudebox model). Rejected because hierarchical grouping is more intuitive ("Java with Maven") and avoids the complexity of a dependency solver.

### 3. Dockerfile decomposition: core + snippets + tail

The monolithic `assets/Dockerfile` splits into three embedded parts:

- **`assets/Dockerfile.core`**: OS packages, Docker, GitHub/GitLab CLIs, user creation, mise/fnm/uv installation (managers only, no languages), agent CLIs
- **`assets/Dockerfile.tail`**: oh-my-zsh, shell config, git config, tmux, entrypoint COPY, final USER/WORKDIR
- Profile `DockerSnippet` fields: language-specific RUN instructions inserted between core and tail

The image builder (`image.go`) assembles:
```
Dockerfile.core
+ java.DockerSnippet (if active)
+ python.DockerSnippet (if active)
+ node.DockerSnippet (if active)
+ Dockerfile.tail
```

For the project image, the existing `generateProjectDockerfile` logic is extended: it starts from the base image (which already has global profiles baked in) and appends project-level profile snippets + user `packages`.

**Alternative considered**: Build args to conditionally skip sections in a single Dockerfile. Rejected because it doesn't scale and prevents profiles from contributing arbitrary instructions.

### 4. Entrypoint decomposition

Same pattern as Dockerfile: `entrypoint.core` + profile snippets + `entrypoint.tail`.

- **Core**: PATH setup, git config from host, SSH, direnv, Docker daemon, welcome banner header
- **Profile snippets**: fnm env setup (node), mise activation (java), venv auto-creation (python/uv)
- **Tail**: Welcome banner tool versions (dynamically assembled from active profiles), `exec "$@"`

The entrypoint is assembled at image build time (written to a temp file and COPYed), not at container start. This means entrypoint content is determined by which profiles are active when the image is built.

### 5. nil means all, empty means none

When `profiles` is not specified in any config layer, it defaults to all built-in profiles (`[java, python, node]`). This preserves backwards compatibility — existing users with no `profiles` key get exactly the same image as today.

When `profiles: []` is explicitly set, no language profiles are activated — only the core OS/shell/agents are installed.

In the merge chain, `profiles` follows last-wins semantics like other list fields: if a project config specifies `profiles: [java]`, it replaces the global default entirely (no merging of profile lists between layers).

**Alternative considered**: Additive merge (project profiles added to global profiles). Rejected because there's no intuitive way to remove a profile, and "I said java, I meant only java" is the expected behavior.

### 6. Config defaults from profiles merge early

Profile config defaults (e.g., java profile setting `versions.java: 21`) are injected into the merge chain after the config layer that activated them but before subsequent layers:

```
Global config  →  global profiles' Config()  →  Project config  →  project profiles' Config()  →  Local config  →  CLI flags
```

This means project config can override global profile defaults, and local config can override project profile defaults. The existing merge logic (scalars: last wins, maps: merge per key, lists: concatenate) applies unchanged.

### 7. Dynamic CacheDirs

The hardcoded `CacheDirs` map in `container.go` is replaced by aggregation from active profiles:

```go
func AggregateCacheDirs(profiles []*profile.Profile) map[string]string {
    dirs := map[string]string{}
    for _, p := range profiles {
        maps.Copy(dirs, p.CacheDirs)
    }
    return dirs
}
```

`container.RunArgs` receives the resolved profile list and uses the aggregated cache dirs for volume mounts. The migration code in `main.go` that moves old bind-mounted caches also uses this dynamic map.

### 8. Hash computation includes profiles

Image cache invalidation must account for which profiles are active and their content. The base image hash includes:
- `Dockerfile.core` + `Dockerfile.tail` content (as today for the monolithic file)
- Entrypoint core + tail content
- Sorted concatenation of active global profile snippets (Docker + entrypoint)

The project image hash includes:
- Active project profile snippets
- User `packages` (as today)
- Custom java version (as today)

This ensures that changing which profiles are active triggers a rebuild.

## Risks / Trade-offs

**Dockerfile split increases build-time complexity** → Mitigated by keeping the assembly logic simple (string concatenation in deterministic order) and testing that the assembled output matches the current monolithic Dockerfile when all profiles are active.

**Entrypoint assembly at build time locks profile selection to image build** → Acceptable because profiles are a build-time concept (they determine what's installed). Runtime profile switching would require a fundamentally different approach and isn't needed.

**`profiles` last-wins semantics could surprise users who expect additive behavior** → Mitigated by documenting clearly and by the fact that the most common case (no profiles specified = all active) requires zero config changes.

**Three embedded file pairs (core/tail for Dockerfile and entrypoint) instead of one each** → More files to manage, but each is smaller and has a single concern. The go:embed declarations grow but remain straightforward.
