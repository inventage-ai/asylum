# Design

## Context

`workspace.Resolve(projectDir, home)` runs early in `main`, before config, agent and image resolution. It returns the project directory, or creates a fresh workspace when the directory is the home directory or `/`. `parseArgs` has already run at that point, and it forwards `--continue` and `--resume` unchanged to the agent through `extraArgs`.

## Goals / Non-Goals

**Goals:**
- `asylum --continue` and `asylum --resume` from the home directory land in the workspace the user last worked in.

**Non-Goals:**
- Reuse under `default-resume: true` without an explicit flag. That config resumes sessions within a directory and says nothing about which directory.
- Recognising Claude's short `-c` flag. Asylum doesn't parse it today, so it can't tell `-c` apart from other passthrough flags.
- Choosing a workspace by agent session state.

## Decisions

### Detect the flags from `extraArgs`
`main` sets `reuse := slices.Contains(extraArgs, "--continue") || slices.Contains(extraArgs, "--resume")` and passes it to `Resolve`. `parseArgs` stays unchanged, and the agent still receives the flags.

### "Newest" is the most recent directory modification time
- A directory's mtime changes whenever an entry directly inside it is created, removed or renamed. That tracks "last worked in" well enough for a scratch workspace, and it works on every OS.
- Creation time was rejected. Go's `os.Stat` doesn't expose birth time portably, and the name's date prefix only has day granularity.
- Asking the agent which workspace has the newest session was rejected. Agent config isn't resolved yet when the workspace is chosen, and the rule would differ per agent.

### Only generated names count
A regexp `^\d{4}-\d{2}-\d{2}(-[a-z]+){3}$`, matching what `name()` produces, filters entries. This keeps a user's own `~/asylum-workspace/notes` from being picked.

### Fall back to a fresh workspace
If nothing matches, `Resolve` creates a fresh workspace. The agent's `--continue` then fails as it does today, which is honest. A reused workspace gets `redirected=true` plus a `reused` indicator, so `main` can print "Continuing in the newest workspace" instead of "Started a fresh workspace".

## Risks / Trade-offs

- [The user worked in an older workspace most recently, but only edited files in subdirectories, so its mtime didn't change] → asylum picks a different workspace. The warning prints the path, so the user notices, and `cd` into the right workspace still works.
- [The signature of `Resolve` changes] → It has a single caller in `main`.
