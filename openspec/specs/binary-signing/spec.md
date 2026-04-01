## ADDED Requirements

### Requirement: Darwin binaries are code-signed
The release workflow SHALL sign all darwin binaries with a Developer ID Application certificate using `codesign` with hardened runtime and secure timestamp.

#### Scenario: Signed binary passes codesign verification
- **WHEN** a darwin binary is signed during CI
- **THEN** `codesign --verify --deep --strict` succeeds on the binary

#### Scenario: Hardened runtime is enabled
- **WHEN** a darwin binary is signed
- **THEN** the binary is signed with `--options runtime` and `--timestamp` flags

### Requirement: Darwin binaries are notarized
The release workflow SHALL submit signed darwin binaries to Apple's notary service and wait for approval before publishing.

#### Scenario: Notarization succeeds
- **WHEN** a signed darwin binary is submitted via `notarytool submit`
- **THEN** `notarytool wait` reports "Accepted" status

#### Scenario: Notarized binary passes Gatekeeper
- **WHEN** a user downloads a notarized darwin binary
- **THEN** macOS Gatekeeper does not show an "unidentified developer" warning

#### Scenario: Notarization failure blocks release
- **WHEN** `notarytool wait` reports a status other than "Accepted"
- **THEN** the CI job fails and no release is published

### Requirement: Signing uses ephemeral keychain
The CI workflow SHALL create a temporary keychain, import the certificate from GitHub Secrets, perform signing, and delete the keychain afterward. No persistent keychain state on the runner.

#### Scenario: Keychain cleanup after signing
- **WHEN** the signing step completes (success or failure)
- **THEN** the temporary keychain is deleted

### Requirement: SHA256 checksums for all binaries
The release workflow SHALL generate a `checksums.txt` file containing SHA256 hashes of all release binaries and upload it as a release asset.

#### Scenario: Checksums file is published
- **WHEN** a release is created
- **THEN** a `checksums.txt` file is attached containing one `sha256  filename` line per binary

#### Scenario: Checksums match binaries
- **WHEN** a user downloads a binary and `checksums.txt`
- **THEN** `sha256sum --check checksums.txt` verifies the binary's integrity

### Requirement: GitHub build provenance attestation
The release workflow SHALL generate SLSA build provenance attestation for all release binaries using `actions/attest-build-provenance`.

#### Scenario: Attestation is verifiable
- **WHEN** a user runs `gh attestation verify <binary> --repo inventage-ai/asylum`
- **THEN** the attestation confirms the binary was built from the expected commit

### Requirement: Install script supports checksum verification
The install script SHALL optionally verify the downloaded binary against checksums from the release.

#### Scenario: Checksum verification when available
- **WHEN** `checksums.txt` is available in the release
- **THEN** the install script downloads it and verifies the binary hash before installing

#### Scenario: Graceful fallback without checksum tools
- **WHEN** neither `sha256sum` nor `shasum` is available on the system
- **THEN** the install script skips verification with a warning and proceeds with installation
