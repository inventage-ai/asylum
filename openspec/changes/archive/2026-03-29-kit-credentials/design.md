## Context

Credentials in Asylum are currently handled by a first-run hack: `internal/firstrun/firstrun.go` checks for `~/.m2/settings.xml`, prompts Y/n, and mounts the entire file read-only via a `volumes:` entry. This has no kit awareness, no filtering, and no path for other ecosystems.

Kits today are purely static — they declare strings (DockerSnippet, EntrypointSnippet, CacheDirs, etc.) that get aggregated at build/launch time. Credentials require the first dynamic kit behavior: a function that inspects the project, reads host files, and produces scoped output.

The Maven `~/.m2` directory is already a named cache volume (`CacheDirs: {"maven": "~/.m2"}`). The generated settings.xml must be bind-mounted as a single file into this named volume.

## Goals / Non-Goals

**Goals:**
- Kits own their credential logic end-to-end (discovery, extraction, filtering, file generation)
- Maven credentials are scoped to the project: only server entries matching repositories in `pom.xml` are mounted
- Users control credentials per-kit via config (`auto`, explicit list, or off)
- First-run onboarding presents credential-capable kits via TUI multiselect
- Interface supports future kit credential implementations without changes

**Non-Goals:**
- Credential support for node, docker, or github (interface only, no implementation)
- Multi-module pom.xml walking or parent POM resolution
- Maven profile activation evaluation (all profiles' repos are included regardless)
- Migration of existing `volumes: [~/.m2/settings.xml:ro]` config entries
- Credential encryption or secret management

## Decisions

### 1. CredentialFunc as a function field on Kit, not an interface

Add `CredentialFunc func(CredentialOpts) ([]CredentialMount, error)` to the Kit struct. Kits that don't handle credentials leave it nil.

**Why not an interface?** The project convention is "no unnecessary interfaces — don't create an interface until there are two implementations." A function field is simpler, consistent with how kits are already structured (concrete structs, not interface implementations), and avoids ceremony for kits that don't need credentials.

### 2. Credential generation at container launch, not image build

CredentialFunc runs in `appendVolumes()` during `container.RunArgs()`, not during image build. Generated files are written to `~/.asylum/projects/<cname>/credentials/` and bind-mounted read-only.

**Why?** Credentials depend on both the project (pom.xml) and the host (~/.m2/settings.xml), both of which can change between runs. Baking them into the image would require rebuilds on credential changes and would persist secrets in Docker image layers.

### 3. Generated file in project credentials dir, not a temp file

Write to `~/.asylum/projects/<cname>/credentials/settings.xml` rather than `os.CreateTemp`. The file is overwritten fresh every launch.

**Why?** Temp files risk accumulation. A fixed path under the project dir is predictable, inspectable for debugging, and automatically scoped to the project.

### 4. Credentials field supports string, bool, and list via custom YAML unmarshaling

The `Credentials` field on KitConfig uses a custom type that accepts `auto` (string), `false` (bool), or a string list. This mirrors how `shadow-node-modules` handles bool, but extended for the polymorphic case.

**Alternatives considered:** Separate fields (`credentials-mode` + `credentials-ids`) — rejected as more complex config surface for the user. Single field with custom unmarshal is cleaner.

### 5. CredentialFunc on the sub-kit (java/maven), not the parent (java)

Maven credential handling lives on the `java/maven` sub-kit, not on `java`. The parent kit has no credential behavior of its own.

**Why?** Gradle would have different credential logic. Putting it on the sub-kit keeps concerns separated and avoids the parent needing to dispatch to children.

### 6. XML parsing with encoding/xml, no new dependencies

Parse pom.xml and settings.xml using Go's standard `encoding/xml`. The parsing is targeted — only extract `<repository>/<id>`, `<pluginRepository>/<id>`, `<distributionManagement>/*/<id>`, and `<server>` entries. No need for a full Maven model.

### 7. First-run uses TUI multiselect for all credential-capable kits

Replace the current file-detection + Y/n prompt with `tui.MultiSelect` showing every kit that has a non-nil CredentialFunc. No host file detection — the prompt is about enabling the capability, not checking if files exist. Missing host files are handled gracefully at launch (CredentialFunc returns empty or warns).

**Why no detection?** The user might not have credentials now but will later. The onboarding question is "do you want this kit to handle credentials?" not "do you have credentials right now?"

## Risks / Trade-offs

**[pom.xml parsing is fragile for edge cases]** → Mitigation: only parse the root pom.xml, don't attempt POM inheritance or property interpolation. Explicit mode (`credentials: [id1, id2]`) is the escape hatch for complex setups.

**[Bind-mounting a file into a named volume has ordering sensitivity]** → Mitigation: credential mounts are added after cache volume mounts in `appendVolumes`, which is the correct Docker ordering. Add a comment documenting this dependency.

**[CredentialFunc runs every launch, adding latency]** → Mitigation: the work is trivial (read two small XML files, filter, write one). No measurable impact.

**[Generated settings.xml may lack mirror/proxy config the user expects]** → Mitigation: the generated file only contains `<servers>`. If users need mirrors or proxies, they can add the full file as a custom volume mount (existing behavior still works).
