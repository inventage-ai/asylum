# Spec Delta

## Purpose

Lets a kit declare the one host operating system it works on, so Asylum only offers it on that OS while still honouring it wherever a user enables it by hand.

## ADDED Requirements

### Requirement: Kits declare a host OS
A kit SHALL be able to declare the host operating system it works on. A kit that declares none SHALL count as available on every host. A kit counts as available on a host when it declares no OS or declares the host's OS. The `iterm` and `dropshare` kits SHALL declare macOS.

#### Scenario: Kit without a host OS
- **WHEN** a kit declares no host OS
- **THEN** it is available on macOS and Linux hosts alike

#### Scenario: macOS kit on a Linux host
- **WHEN** Asylum runs on a Linux host
- **THEN** `iterm` and `dropshare` are not available

#### Scenario: macOS kit on a macOS host
- **WHEN** Asylum runs on a macOS host
- **THEN** `iterm` and `dropshare` are available

### Requirement: Unavailable kits are not offered
The system SHALL leave a kit that is not available on the host out of every surface that offers kits. It SHALL NOT appear in the first-run wizard's kit step, the kit-sync prompt, or the Kits tab of `asylum config`. It SHALL NOT be written to a generated config, neither as an active entry nor as a comment, whether by the first-run wizard, the non-interactive default config, or the kit-sync flow. It SHALL NOT appear in the "Disabled Kits" section of the sandbox rules.

#### Scenario: First run on Linux
- **WHEN** a user runs Asylum for the first time on a Linux host
- **THEN** neither the wizard nor the generated `~/.asylum/config.yaml` mentions `iterm` or `dropshare`

#### Scenario: Kit sync on Linux
- **WHEN** an Asylum update adds a macOS-only kit and the user starts Asylum on a Linux host
- **THEN** the kit-sync prompt does not offer it and no comment line for it is written to the config

#### Scenario: Sandbox rules on Linux
- **WHEN** a container starts on a Linux host with neither macOS-only kit enabled
- **THEN** the "Disabled Kits" section of the sandbox rules lists neither `iterm` nor `dropshare`

#### Scenario: Offered on macOS
- **WHEN** the same situations occur on a macOS host
- **THEN** `iterm` and `dropshare` are offered and listed as before

### Requirement: Enabled unavailable kits behave as before
A kit that is not available on the host but is enabled in the config SHALL resolve, build, and run exactly as it would on a host it declares. The system SHALL NOT block, drop, or warn about it.

#### Scenario: Shared config on Linux
- **WHEN** `~/.asylum/config.yaml` enables `dropshare` and Asylum starts on a Linux host
- **THEN** the kit is active as configured and appears under "Active Kits" in the sandbox rules

#### Scenario: Config command keeps the entry
- **WHEN** the config above is open in `asylum config` on a Linux host and the user saves the Kits tab
- **THEN** the `dropshare` entry in the config is unchanged
