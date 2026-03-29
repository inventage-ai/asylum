---
name: complete-change
description: Finalize a change - archive if needed, commit if needed, rebase onto main and fast-forward main.
disable-model-invocation: true
---

Finalize a completed change by archiving, committing, and syncing with main.

**Steps**

1. **Archive change (if needed)**

   Run `openspec list --json` to check for active changes.

   - If an active change exists with all tasks complete: use the Skill tool to invoke `opsx:archive` for it. When prompted about spec sync, always choose "Sync now".
   - If no active changes exist: skip this step.
   - If an active change has incomplete tasks: abort with "Change '<name>' has incomplete tasks. Complete them first or run `/opsx:archive` manually."

2. **Commit (if needed)**

   Run `git status --short` to check for uncommitted changes.

   - If there are staged or unstaged changes (excluding untracked files that look unrelated): use the Skill tool to invoke `commit`.
   - If the working tree is clean: skip this step.

3. **Rebase onto main**

   Rebase the current branch onto `main` to pick up any changes that landed on main since the branch diverged.

   ```bash
   git rebase main
   ```

   **If conflicts occur:**
   - For each conflicted file, determine which version is correct:
     - If the conflict is in CHANGELOG.md: keep both sides (ours + theirs), resolve ordering
     - If the conflict is in code files where our branch has a full rewrite (e.g., config restructure): take ours (`git checkout --ours`)
     - If the conflict is in code files where main has a fix we should keep: merge manually
     - For OpenSpec artifacts: take ours (our branch has the latest)
   - After resolving each file: `git add <file>`
   - Continue: `git rebase --continue`
   - Repeat until rebase completes

   If rebase cannot be resolved automatically, abort with `git rebase --abort` and explain what went wrong.

4. **Verify after rebase**

   ```bash
   go build ./... && go test ./...
   ```

   If build or tests fail after rebase: abort and report the failures. Do NOT force-push broken code.

5. **Fast-forward main**

   Record the current branch name first, then bring main up to date:

   ```bash
   BRANCH=$(git rev-parse --abbrev-ref HEAD)
   git checkout main
   git merge --ff-only $BRANCH
   git checkout $BRANCH
   ```

   If fast-forward fails (main has diverged): switch back to the original branch (`git checkout $BRANCH`), then abort and explain. The user needs to decide how to proceed.

   **Always return to the original branch**, even if the fast-forward succeeds. The user should end up on the same branch they started on.

6. **Show summary**

   Display:
   - Whether a change was archived
   - Whether a commit was created
   - Whether rebase had conflicts (and how they were resolved)
   - Whether main was updated
   - Final `git log --oneline -5` showing the result

**Guardrails**
- Never force-push
- Never push to remote (this is local only)
- If any step fails, stop and explain — don't try to recover silently
- Always verify build + tests pass after rebase before updating main
- If the current branch IS main, skip the rebase and fast-forward steps
