# Java Kit

JDK 17, 21, and 25 via mise, with Maven and Gradle available as sub-kits.

**Activation: Default** — added to config on first detection; active when present.

## What's Included

- **JDK 17, 21, 25** installed via [mise](https://mise.jdx.dev/) (default: 21)
- **Maven** (via sub-kit) with dependency caching
- **Gradle** (via sub-kit) with dependency caching

## Configuration

```yaml
kits:
  java:
    default-version: "17"     # which JDK is the global default
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `default-version` | string | `"21"` | JDK version to set as the global default |

JDK 17, 21, and 25 are always pre-installed. You can install additional versions at runtime with `mise install java@<version>`.

## Sub-Kits

### java/maven

Installs Maven via apt. Dependency cache persisted at `~/.m2`.

#### Maven Credentials

The `java/maven` kit can inject credentials from your host's `~/.m2/settings.xml` into the container. Rather than mounting the full file, it generates a filtered `settings.xml` containing only the server entries that your project's `pom.xml` actually references — so unrelated credentials stay on the host.

Configure credentials under the `java` kit:

**Auto mode** (default when credentials are enabled): Asylum reads the `pom.xml` in your project root, extracts all referenced server IDs (from `repositories`, `pluginRepositories`, `distributionManagement`, and profiles), and injects the matching entries.

```yaml
kits:
  java:
    credentials: auto
```

**Explicit mode**: Specify server IDs directly, bypassing `pom.xml` discovery. Useful for multi-module projects or when the root `pom.xml` doesn't declare all repositories.

```yaml
kits:
  java:
    credentials:
      - nexus-releases
      - nexus-snapshots
```

**Disabling**: Set `credentials: false` to prevent any credential injection.

The filtered `settings.xml` is mounted read-only at `~/.m2/settings.xml` inside the container. If `~/.m2/settings.xml` does not exist on the host, or no referenced server IDs match, nothing is mounted. If a server ID is referenced in `pom.xml` but missing from `settings.xml`, a comment is added in its place in the generated file.

Credentials are enabled via `asylum config` (interactive) or by editing `.asylum` directly. Changes to credentials take effect on the next fresh container start — if a container is already running, use `asylum --rebuild` to apply them.

### java/gradle

Installs Gradle via mise. Dependency cache persisted at `~/.gradle`.

## Version Switching

Switch JDK version inside the container:

```sh
mise use java@17
mise use java@25
```

You can also set the version via environment variable in your config:

```yaml
kits:
  java:
    default-version: "17"
```

Or detect it automatically from `.tool-versions` in your project root.

## Auto-Detection

If your project has a `.tool-versions` file with a `java` entry, Asylum reads the version from it and sets it as the default in the container.
