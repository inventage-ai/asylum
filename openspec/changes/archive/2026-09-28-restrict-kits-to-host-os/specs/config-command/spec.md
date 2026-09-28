# Spec Delta

## MODIFIED Requirements

### Requirement: Kits tab shows multiselect of all kits
The Kits tab SHALL display a multiselect list of all registered kits (excluding always-on kits, hidden kits, and kits that are not available on the host). Kits that are currently active (present in config and not disabled) SHALL be pre-selected.

#### Scenario: Kit list population
- **WHEN** the Kits tab is displayed
- **THEN** all registered kits with tier Default or OptIn that are not hidden and are available on the host are listed, with currently active (not disabled) kits checked

#### Scenario: Toggle kit selection
- **WHEN** the user presses space on a kit entry
- **THEN** the kit's selection state is toggled (checked/unchecked)

#### Scenario: Kits for another host OS excluded
- **WHEN** the Kits tab is displayed on a Linux host
- **THEN** `iterm` and `dropshare` are not listed, and saving the tab leaves any existing config entry for them unchanged
