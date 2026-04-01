## Context

Release binaries are currently built on a single `ubuntu-latest` runner and uploaded as raw files to GitHub Releases. macOS users get Gatekeeper warnings on first run. There are no checksums or provenance attestation.

The team has a Developer ID Application certificate (.p12) and App Store Connect API key (.p8) for notarization. Five GitHub Secrets will be configured: `APPLE_CERTIFICATE_P12`, `APPLE_CERTIFICATE_PASSWORD`, `APPLE_API_KEY`, `APPLE_API_KEY_ID`, `APPLE_API_ISSUER_ID`.

## Goals / Non-Goals

**Goals:**
- Darwin binaries pass Gatekeeper without warnings
- All binaries have SHA256 checksums for integrity verification
- All binaries have GitHub provenance attestation
- Signing works on any macOS runner (not tied to specific hardware)
- Both release and dev-release workflows get the same signing treatment

**Non-Goals:**
- DMG/PKG packaging — bare signed binaries are sufficient
- Linux binary signing (GPG or Sigstore — future work)
- Automatic checksum verification in install.sh (add support, keep it optional)
- Signing the install script itself

## Decisions

### Split build into platform-specific jobs
Build linux targets on `ubuntu-latest` and darwin targets on `macos-latest`. A final `release` job collects artifacts and creates the GitHub release.

*Alternative: use `rcodesign` (Rust reimplementation) to sign on Linux.* Rejected — `rcodesign` can sign but notarization still requires Apple's toolchain. Adds complexity for partial benefit.

*Alternative: build everything on macOS.* Rejected — unnecessary cost (10x minute rate) for linux targets that don't need signing.

### Ephemeral keychain for signing
Each CI run creates a temporary keychain, imports the .p12 from GitHub Secrets, signs, then deletes the keychain. This avoids depending on persistent state on any specific runner.

Steps:
1. Decode `APPLE_CERTIFICATE_P12` from base64 to a temp file
2. `security create-keychain -p <random> build.keychain`
3. `security import <p12> -k build.keychain -P $PASSWORD -T /usr/bin/codesign`
4. `security set-key-partition-list -S apple-tool:,apple: -k <random> build.keychain`
5. Sign with `codesign --sign "Developer ID Application: Inventage AG" --options runtime --timestamp`
6. Delete keychain in post-step

### Notarization via zip submission
`notarytool` requires a zip/dmg/pkg container. Wrap each binary in a zip, submit, wait for approval, then distribute the bare binary (not the zip). Gatekeeper verifies notarization status online for bare binaries.

Auth via App Store Connect API key (stored as secrets), not app-specific password — API keys are org-level and don't depend on a personal Apple ID.

### Checksums as a release asset
Generate `checksums.txt` containing SHA256 hashes of all binaries. Upload alongside binaries. Update `install.sh` to optionally verify the checksum after download (requires `sha256sum` or `shasum`).

### GitHub attestation
Add `actions/attest-build-provenance@v2` for all binaries. Requires `id-token: write` and `attestations: write` permissions. Users verify with `gh attestation verify`.

## Risks / Trade-offs

- **macOS CI cost** → Darwin builds on `macos-latest` cost 10x per minute. Mitigated by only building darwin targets there (Go cross-compile is fast, ~1-2 min). Can switch to self-hosted macOS runner later if costs matter.
- **Notarization latency** → Apple's notarization typically takes 1-5 minutes but can occasionally be slow. CI jobs need a reasonable timeout. `notarytool wait` handles polling.
- **Bare binary notarization** → Can't staple tickets to bare Mach-O binaries. First run needs internet for Gatekeeper to verify online. Airgapped environments won't benefit. This is standard for Go CLI tools.
- **Certificate expiry** → Developer ID certs are valid for 5 years. Need to track renewal (expires ~2031). Notarization will fail when cert expires.
- **Secret rotation** → If the API key or cert needs replacing, someone with Apple Developer admin access must generate new credentials and update GitHub Secrets.
- **Self-hosted runner on public repo** → Forks can modify workflow YAML to target self-hosted runners. Mitigated by: (1) repo setting "Require approval for all outside collaborators" is enabled — no fork workflow runs without maintainer approval, (2) self-hosted runners are only referenced in `push`-triggered workflows (`release.yml`, `dev-release.yml`), never in `pull_request`-triggered workflows. The `ci.yml` workflow MUST only use GitHub-hosted runners.
