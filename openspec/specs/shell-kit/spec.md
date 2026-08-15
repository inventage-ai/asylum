# shell-kit Specification

## Purpose

Makes the container's interactive shell usable rather than bare: oh-my-zsh with a configured theme, direnv hooks, and the related shell setup an agent or a user dropping into `asylum shell` expects to find.

## Requirements

### Requirement: Shell configuration kit
The system SHALL provide a `shell` kit that installs oh-my-zsh, configures the zsh theme, sets up direnv hooks, configures terminal size handling, and provides a default tmux configuration. The kit SHALL be default-on.

#### Scenario: Shell kit active
- **WHEN** the shell kit is active
- **THEN** the container has oh-my-zsh with robbyrussell theme, direnv hooks in bash/zsh, terminal size handling via stty, and a tmux configuration

#### Scenario: Shell kit disabled
- **WHEN** the shell kit is disabled
- **THEN** the container has bare zsh with no oh-my-zsh, no tmux config, and no direnv hooks (direnv is still installed but hooks are not configured)
