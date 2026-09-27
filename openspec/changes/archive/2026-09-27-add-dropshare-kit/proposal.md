# Proposal

## Why

Agents in the sandbox often produce files worth sharing, such as screenshots for GitHub issues or build artifacts. Dropshare on the host can publish them, but today the agent only hands a URL to the host opener, and the share link lands on the host clipboard, out of the agent's reach. The agent cannot put the link into an issue or a comment without the user pasting it back. Dropshare can report the link through a callback URL on a custom scheme, and a proof of concept confirmed that such a callback reaches a host applet within a few seconds.

## What Changes

- Add an opt-in `dropshare` kit for macOS hosts.
- The kit ships a container command and a Claude skill. The command uploads a file from the container and prints the share URL.
- The kit contributes a broker route. The route copies the file out of the container with `docker cp`, starts the Dropshare upload of that copy with a per-request callback, waits for the callback, and returns the share URL.
- Asylum builds and registers one global callback applet on the host at `~/.asylum/dropshare/`, on first use. The applet receives Dropshare callbacks and hands them to the waiting broker through an inbox directory.

## Capabilities

### New Capabilities

- `dropshare-kit`: Upload a file from inside the container to Dropshare on the host and return the share URL to the caller.

### Modified Capabilities

None.

## Impact

- New code: `internal/kit/dropshare.go` (kit, route, callback parsing, applet management) and tests.
- `internal/docker`: `CopyOut` streams a container path as a tar archive.
- New host state: `~/.asylum/dropshare/` holds the applet, the callback inbox, and staged copies of files being uploaded. LaunchServices gains one registration for the `asylum-dropshare:` scheme.
- Host tools used: `osacompile`, `plutil`, `codesign`, `lsregister`, and `open`. All ship with macOS. No new Go dependencies.
- Users with a hand-written `dropshare` skill in `~/.claude/skills/` get a name clash with the kit's skill. They should delete theirs, and Asylum does not touch `~/.claude`.
- Users who allowlisted `dropshare5` in `kits.browser-open.schemes` only for uploads can remove it.
- Docs: `docs/kits/dropshare.md`, CHANGELOG entry.
