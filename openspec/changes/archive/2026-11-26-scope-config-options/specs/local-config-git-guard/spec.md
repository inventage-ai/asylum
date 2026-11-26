# Spec Delta

## Purpose

Keeps the local config layer personal: a repository must not be able to ship `.asylum.local`, because that file may set options a committed `.asylum` may not.

## ADDED Requirements

### Requirement: Refuse a tracked local config
When the local layer file (`.asylum.local` or `.asylum.local.yaml`) exists and git tracks it, asylum SHALL refuse to load config. This SHALL happen before the file is migrated or read. The error SHALL name the file and tell the user to untrack it with `git rm --cached` and keep it ignored. A file staged in the index SHALL count as tracked.

#### Scenario: Committed local config
- **WHEN** `.asylum.local` exists and is committed in the project's repository
- **THEN** asylum exits with an error naming `.asylum.local` and showing `git rm --cached .asylum.local`, without starting a container

#### Scenario: Staged local config
- **WHEN** `.asylum.local.yaml` exists, is not committed, but has been added with `git add`
- **THEN** asylum exits with the same error for `.asylum.local.yaml`

#### Scenario: Project directory inside a larger repository
- **WHEN** the project directory is a subdirectory of a repository and its `.asylum.local` is tracked
- **THEN** asylum exits with the error

#### Scenario: Committed under another case
- **WHEN** the repository commits `.ASYLUM.LOCAL` and the case-insensitive filesystem makes asylum read it as `.asylum.local`
- **THEN** asylum exits with the error, naming `.ASYLUM.LOCAL` in the `git rm --cached` command

#### Scenario: Untracked local config
- **WHEN** `.asylum.local` exists and is not tracked
- **THEN** config loads normally

### Requirement: Exclude the local config from git
On session start in a git repository, asylum SHALL append each of `.asylum.local` and `.asylum.local.yaml` that git does not already ignore to the repository's local exclude file and log one line naming the added entries. This SHALL happen whether or not the file exists. In a linked worktree, the exclude file of the shared repository SHALL be used.

#### Scenario: Not yet ignored
- **WHEN** a session starts in a repository whose `.gitignore`, global excludes and local exclude file do not match `.asylum.local`
- **THEN** both filenames are appended to the local exclude file and `git check-ignore .asylum.local` succeeds afterwards

#### Scenario: Already ignored
- **WHEN** `.gitignore` already contains `.asylum.local*`
- **THEN** the local exclude file is not modified and nothing is logged

#### Scenario: Only the alias missing
- **WHEN** `.gitignore` lists `.asylum.local` but nothing matches `.asylum.local.yaml`
- **THEN** only `.asylum.local.yaml` is appended to the local exclude file

#### Scenario: Missing info directory
- **WHEN** the repository has no `info` directory
- **THEN** it is created together with the exclude file

#### Scenario: Linked worktree
- **WHEN** the project directory is a linked worktree created with `git worktree add`
- **THEN** the entries are written to the exclude file of the shared repository, not to a file inside the worktree

### Requirement: Git checks are skipped outside git
When the project directory is not inside a git repository, or git is not installed on the host, asylum SHALL skip the tracking check and the exclude step without logging anything.

#### Scenario: Not a repository
- **WHEN** the project directory is not inside a git repository and `.asylum.local` exists
- **THEN** config loads normally and no exclude file is written

#### Scenario: Git not installed
- **WHEN** `git` is not on the host's PATH
- **THEN** asylum starts normally
