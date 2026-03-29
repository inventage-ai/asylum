---
name: commit
description: Create a commit from current changes with an appropriate message. Handles OpenSpec change lifecycle (verify, archive) and CHANGELOG updates automatically.
disable-model-invocation: true
---

Create a commit from the current working tree changes.

**Steps**

1. **Assess current changes**

   Run `git status` (excluding untracked files) and `git diff --stat` to understand what's changed.

   If there are no staged or unstaged changes, abort with: "Nothing to commit."

2. **Check for active OpenSpec changes**

   Run `openspec list --json` to check for active (non-archived) changes.

   For each active change:
   - Read its `tasks.md` to check task completion status
   - Determine if the current changes relate to this change (check if modified files overlap with what the change's tasks describe)

   **If changes relate to an active OpenSpec change:**

   a. **Check task completion**:
      - Parse `tasks.md` for `- [ ]` (incomplete) vs `- [x]` (complete)
      - If there are incomplete tasks: ask the user "Change '<name>' has X incomplete tasks. Commit anyway?"
        - If user declines: abort
        - If user confirms: proceed without archiving
      - If all tasks are complete: continue to verification

   b. **Run verification if not already done this session**:
      - Use the Skill tool to invoke `opsx:verify` for the change
      - If CRITICAL issues are found: alert the user and ask whether to proceed
      - If no critical issues (or verify was already run): continue to archive

   c. **Archive the change**:
      - Use the Skill tool to invoke `opsx:archive` for the change

3. **Update documentation**

   If the changes being committed affect user-facing behavior (new features, new kits, changed config options, new commands, changed behavior), check whether the `docs/` directory needs updating.

   Relevant docs areas:
   - `docs/kits/` — one page per kit, plus `index.md` with the kit table
   - `docs/configuration/` — config file format, options, flags
   - `docs/commands/` — CLI commands and subcommands
   - `docs/concepts/` — architecture concepts (agents, images, sessions, mounts)

   Read the relevant existing doc pages and update them to reflect the changes. For new kits, create a new kit page following the style of existing ones and add a row to `docs/kits/index.md`. Stage any doc changes.

   Skip doc updates for:
   - Pure refactoring with no behavior change
   - Test-only changes
   - Internal implementation details not visible to users
   - OpenSpec artifact files

4. **Update CHANGELOG.md**

   Read `CHANGELOG.md` and the `## Unreleased` section.

   Review the changes being committed. If they represent a user-facing change (new feature, bug fix, behavior change), update the Unreleased section following these rules:

   - **Merge, don't duplicate**: If an existing entry covers the same feature or area, update that entry to reflect the new state rather than adding a separate line. For example, if "Added: Kit credential system" already exists and this commit enhances it, update that entry — don't add a second one.
   - **Don't log fixes for unreleased features**: If the commit fixes a bug in something that was added since the last release (i.e., users on the last release never saw the bug), don't add a Fixed entry. Instead, update the relevant Added/Changed entry to describe the correct behavior, or simply omit it if the Added entry already implies correct behavior.
   - **Don't log intermediate changes**: If the commit changes how an unreleased feature works, update the original Added entry rather than adding a Changed entry. The changelog should describe the final state, not the development history.
   - **Order by importance**: When adding new entries, place them by user impact — breaking changes and major features first within each category.

   If a significant change is missing:
   - Add a concise entry under the appropriate category (Added/Changed/Fixed/Removed)
   - Show the user what was added or updated
   - Stage `CHANGELOG.md`

   Skip CHANGELOG updates for:
   - Pure refactoring with no behavior change
   - Test-only changes
   - Documentation/comment updates
   - OpenSpec artifact files

5. **Stage and commit**

   - Stage all relevant changed files (modified + new files related to the work)
   - Do NOT stage files that look like secrets (`.env`, credentials, etc.)
   - Do NOT stage unrelated changes that happen to be in the working tree — ask the user if unclear
   - Draft a concise commit message (1-2 sentences) that focuses on "why" not "what"
   - Create the commit

6. **Show summary**

   Display:
   - Commit hash and message
   - Files included
   - Whether a change was archived
   - Whether documentation was updated
   - Whether CHANGELOG was updated
   - `git status` after commit to confirm clean state

**Commit Message Guidelines**

- First line: imperative mood, under 72 chars (e.g., "Add worktree volume mounting")
- Focus on the user-facing change, not implementation details
- For bug fixes: describe what was broken
- For features: describe the capability added
- For refactors: "refactor:" prefix, describe what was simplified

**Guardrails**
- Never commit files that likely contain secrets
- Never amend existing commits — always create new ones
- If the working tree has changes to many unrelated files, ask the user which to include
- Always read the diff before writing the commit message
- Use HEREDOC syntax for commit messages to preserve formatting
- Do NOT push after committing
