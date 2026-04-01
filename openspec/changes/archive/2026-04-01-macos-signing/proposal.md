## Why

macOS users downloading asylum binaries from GitHub Releases get Gatekeeper warnings ("unidentified developer") because the darwin binaries are unsigned and unnotarized. This erodes trust and creates friction for new users. Additionally, releases lack checksums and provenance attestation, leaving no way to verify binary integrity or build origin.

## What Changes

- Split release CI jobs: linux targets build on ubuntu, darwin targets build on macOS runner
- Sign darwin binaries with Developer ID Application certificate using ephemeral keychain
- Notarize darwin binaries via `notarytool` with App Store Connect API key auth
- Generate SHA256 checksums file for all release binaries
- Add GitHub build provenance attestation via `actions/attest-build-provenance`
- Update install script to support optional checksum verification
- Apply same signing pipeline to dev-release workflow

## Capabilities

### New Capabilities
- `binary-signing`: macOS code signing and notarization for darwin release binaries, SHA256 checksums, and GitHub attestation for all binaries

### Modified Capabilities
- `dev-release`: Add darwin signing/notarization and checksums to dev release workflow

## Impact

- `.github/workflows/release.yml` — major restructure: split into multi-job workflow with signing and notarization steps
- `.github/workflows/dev-release.yml` — same split and signing additions
- `install.sh` — add checksum verification support
- `Makefile` — may need per-platform build targets
- GitHub Secrets — five new secrets required: `APPLE_CERTIFICATE_P12`, `APPLE_CERTIFICATE_PASSWORD`, `APPLE_API_KEY`, `APPLE_API_KEY_ID`, `APPLE_API_ISSUER_ID`
- CI costs — macOS runner minutes are 10x Linux rate
- `permissions` in workflows need `id-token: write` and `attestations: write` for attestation
