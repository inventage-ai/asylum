# self-update

Update Asylum to a new version.

## Usage

```
asylum self-update              # latest stable release
asylum self-update 0.5.0        # specific version
asylum self-update --dev         # latest dev build from main
asylum self-update --safe        # emergency update (always dev, no checks)
asylum selfupdate               # alias
```

## Description

Downloads and replaces the Asylum binary with a new version. The update is atomic — the new binary is downloaded to a temp file and renamed into place.

## Flags

| Flag | Description |
|------|-------------|
| `--dev` | Update to the latest dev build (rolling release from `main`) |
| `--safe` | Emergency update: always pulls dev, skips all version checks |

## Channels

Asylum supports two release channels:

| Channel | Source | Use case |
|---------|--------|----------|
| `stable` | Latest GitHub release | Default, recommended for most users |
| `dev` | Rolling `dev` pre-release from `main` | Latest features, may be unstable |

To always track dev builds, set the release channel in your global config:

```yaml
# ~/.asylum/config.yaml
release-channel: dev
```

With this setting, `asylum self-update` (without `--dev`) will pull from the dev channel.

## Examples

```sh
# Update to latest stable
asylum self-update

# Install a specific version
asylum self-update 0.4.0

# Track dev builds
asylum self-update --dev

# Emergency: binary is broken, just get something working
asylum self-update --safe
```

## Notes

- On the stable channel, `self-update` skips the download if you're already on the latest version.
- On the dev channel, it shows a changelog of recent commits when the build changes.
- If the binary is in a system directory, you may need to run with `sudo`.
