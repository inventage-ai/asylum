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
