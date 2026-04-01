## 1. Release Workflow Restructure

- [x] 1.1 Split `release.yml` into `build-linux` (ubuntu-latest) and `build-darwin` (macos-latest) jobs that upload artifacts
- [x] 1.2 Add `release` job that downloads artifacts from both build jobs, generates checksums, creates GitHub release, and adds attestation

## 2. macOS Signing

- [x] 2.1 Add ephemeral keychain setup step to `build-darwin` job: decode .p12 from secret, create temp keychain, import cert, configure partition list
- [x] 2.2 Add codesign step for both darwin binaries with `--sign`, `--options runtime`, `--timestamp`
- [x] 2.3 Add keychain cleanup as a post/always step

## 3. Notarization

- [x] 3.1 Add step to zip each signed darwin binary for notarytool submission
- [x] 3.2 Add notarytool submit + wait steps using API key auth from secrets
- [x] 3.3 Ensure job fails if notarization status is not "Accepted"

## 4. Checksums and Attestation

- [x] 4.1 Add step to generate `checksums.txt` with SHA256 hashes of all binaries
- [x] 4.2 Add `actions/attest-build-provenance@v2` step for all binaries
- [x] 4.3 Update workflow permissions to include `id-token: write` and `attestations: write`

## 5. Dev Release Workflow

- [x] 5.1 Apply the same build-linux/build-darwin split to `dev-release.yml`
- [x] 5.2 Add signing, notarization, and checksums to dev-release darwin job
- [x] 5.3 Update dev release creation to include checksums.txt

## 6. Install Script

- [x] 6.1 Add checksum verification to `install.sh`: download checksums.txt, verify binary hash
- [x] 6.2 Handle missing sha256sum/shasum gracefully with warning
