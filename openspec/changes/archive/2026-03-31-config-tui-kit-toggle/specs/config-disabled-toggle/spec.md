## ADDED Requirements

### Requirement: Set disabled flag on kit entry
The system SHALL provide a function to add `disabled: true` as the first property under a kit's YAML entry in the config file, using text-based editing that preserves comments and formatting.

#### Scenario: Disable a kit with no existing disabled field
- **WHEN** `SetKitDisabled` is called for kit `ast-grep` which has an entry `ast-grep:` with no `disabled` field
- **THEN** `disabled: true` is inserted as the first line after the `ast-grep:` key, indented 2 spaces deeper

#### Scenario: Disable a kit that already has disabled: true
- **WHEN** `SetKitDisabled` is called for kit `ast-grep` which already has `disabled: true`
- **THEN** the file is unchanged

#### Scenario: Disable a kit with existing properties
- **WHEN** `SetKitDisabled` is called for kit `apt` which has `packages:` as its first property
- **THEN** `disabled: true` is inserted before `packages:`, becoming the first property

### Requirement: Remove disabled flag from kit entry
The system SHALL provide a function to remove the `disabled: true` line from a kit's YAML entry in the config file, using text-based editing that preserves comments and formatting.

#### Scenario: Remove disabled from a disabled kit
- **WHEN** `RemoveKitDisabled` is called for kit `ast-grep` which has `disabled: true`
- **THEN** the `disabled: true` line is removed and the kit entry remains with any other properties intact

#### Scenario: Remove disabled from a kit without disabled field
- **WHEN** `RemoveKitDisabled` is called for kit `ast-grep` which has no `disabled` field
- **THEN** the file is unchanged

#### Scenario: Remove disabled leaves empty kit entry clean
- **WHEN** `RemoveKitDisabled` is called for kit `ast-grep` whose only property is `disabled: true`
- **THEN** the `disabled: true` line is removed, leaving just the `ast-grep:` key line
