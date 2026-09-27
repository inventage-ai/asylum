# Design

## Context

- Dropshare accepts `dropshare5:///action/upload?file=<path>&callback=<url>`. When the upload finishes, it opens `<callback>?file=<path>&url=<share-url>` through macOS.
- The PoC in `build/dropshare-poc/` established how the callback behaves:
  - Dropshare ignores `http(s)` callbacks. It opens custom-scheme callbacks through LaunchServices.
  - The callback keeps its path, so `asylum-dropshare:///done/<id>` arrives intact.
  - The callback arrives 2–3 s after the upload starts.
  - `file` is path-encoded, but `&`, `+`, and `=` stay raw.
- Brokers run one per container as detached host processes. They read no project config. The session passes spawn-time values as flags and environment.
- Route handlers get a `broker.Ctx`, which exposes the container name and loopback forwarding.
- Kits deliver skills by staging them under `/opt/asylum-skills/.claude/skills/<name>/` at image build.

## Goals / Non-Goals

**Goals:**

- Return the share URL to the caller with no clipboard access.
- Keep one host-side component shared by all projects.
- Keep all parsing and validation in Go, where it can be tested.

**Non-Goals:**

- Dropshare capture actions (screenshots, recordings).
- Uploading directories or static sites.
- Choosing a Dropshare connection or display name.
- Linux or Windows hosts.

## Decisions

### D1. One global applet, callbacks routed by a per-request ID

The route generates a 32-hex-character random ID and passes `callback=asylum-dropshare:///done/<id>`. The applet writes each callback to `~/.asylum/dropshare/inbox/<id>`, and the waiting route watches for its own ID.

- **Alternative: one applet per project.** LaunchServices resolves one handler per scheme, so per-project applets would each need their own scheme. That means N registrations and cleanup whenever a project is removed.
- **Alternative: key callbacks by `file`.** Two concurrent uploads of the same file would collide, and the parse quirks in D6 would decide routing.

### D2. Inbox directory instead of the applet calling a broker

The applet only writes a file. It needs no broker endpoint, no token, and no knowledge of which containers exist.

- **Alternative: the applet POSTs to a broker.** It would have to find the right broker's port or socket and its token, which lives in the container env. Every broker would also need a callback route that the container itself must not reach.

The route polls for `inbox/<id>` every 200 ms. A poll adds nothing noticeable to a 2–3 s upload and needs no dependency. `fsnotify` would be the alternative.

### D3. Applet contents

The applet is an AppleScript `on open location` handler that runs one fixed `sh` snippet via `do shell script`, with the URL passed as a quoted argument. The snippet does three things:

- It checks that the URL matches `asylum-dropshare:///done/<32 hex>?…`, and otherwise exits.
- It writes the raw URL to `inbox/.<id>.tmp`.
- It renames the temp file to `inbox/<id>`, so a reader never sees a partial file.

The applet does no decoding. Go parses the URL (D6).

Go generates the source with the absolute inbox path baked in. The build steps match the PoC:

- `osacompile`
- `plutil -replace` for `CFBundleIdentifier`, `LSUIElement`, and `CFBundleURLTypes`
- `codesign --force --sign -`
- `lsregister -f`

`~/.asylum/dropshare/.applet-sha256` records a hash of the generated source and plist values. A mismatch triggers a rebuild.

### D4. Lazy install, serialized, atomic

The route checks the stamp before each upload. A build first takes an exclusive lock on `~/.asylum/dropshare/.lock`, because two brokers can race on the first upload. It then builds into a temp directory next to the target, renames it over `AsylumDropshare.app`, registers it, and writes the stamp last.

- **Alternative: install at session start.** That would cost every session about 1 s for a feature most sessions won't use.

### D5. Stage the file through `docker cp`

The broker runs `docker cp -L <container>:<path> -` and reads the tar stream. Docker resolves the path and its symlinks inside the container's filesystem, whose mounts are exactly what the container can see. The broker writes the file to `~/.asylum/dropshare/staging/<id>/<basename>`, which is not mounted into the container, and hands Dropshare that copy.

- **The boundary is what the container can see.** Other mounted directories, `/tmp`, and the scratchpad all work. A symlink to a host-only path, such as a file in `~/Documents`, resolves inside the container, where that path does not exist or holds the container's own file.
- **No time-of-check gap.** Dropshare reads a copy the container cannot change, so swapping the source for a symlink after the request changes nothing.
- **No host path oracle.** The broker never looks at host paths, so error messages reveal nothing about the host filesystem.
- **Directories are refused cheaply.** The broker reads the first tar header and kills the copy unless it is a regular file, so a request for `/` or a project tree copies nothing.
- The staged copy keeps the basename, so the share URL keeps its extension. The broker deletes it when the callback arrives or the upload is rejected. After a timeout Dropshare may still be reading it, so the hourly prune removes it.

Uploading anything the container can read adds no exposure: the container can already send those files anywhere over its own network. The skill's rules about secrets remain the guard.

- **Alternative: check the path on the host.** An earlier version resolved the path with `EvalSymlinks` and required it inside the project directory. Dropshare then read the path by name after the check, so the container could swap the file for a symlink to any host file in between. It also rejected legitimate files in other mounts and `/tmp`.
- **Alternative: map container paths to host paths via `docker inspect` mounts.** It covers other mounts but not `/tmp`, needs its own symlink rules per mount, and still needs the staging copy to close the time-of-check gap.

### D6. URL handling

Outgoing, the handler percent-encodes every byte outside the unreserved set for both `file` and `callback`, as the PoC did. It then starts `open -W -g <url>` with `exec.Command` and no shell, and waits for the callback while `open` runs. It kills `open` when the callback arrives, or after 5 s at the latest. Dropshare names the caller by walking up the sender's process tree. A plain `open` often exits before Dropshare looks, and Dropshare then asks the user to confirm "Unknown App". `-W` keeps the process alive, but for a URL it only returns once Dropshare quits, hence the kill. The official Dropshare MCP server does the same, but waits out the full 5 s first. Killing `open` at the callback is safe, because Dropshare names the caller when the request arrives, long before the upload finishes, and it keeps small uploads at about a second.

A failure of `open` itself no longer ends the upload at once. It shows up as a timeout instead. That is rare, since the handler has already confirmed that Dropshare runs and the applet is registered.

Incoming, the handler parses the callback in five steps:

1. Take everything after `?`.
2. Require the `file=` prefix.
3. Split at the last `&url=`.
4. Decode `file` with `url.PathUnescape`, which keeps `+` intact.
5. Require the decoded `file` to equal the staged path, and require `url` to parse as an absolute `https` URL with a host.

Splitting at the last `&url=` is safe because `url` comes last and Dropshare generates it.

### D7. Dropshare not running

Before the first request, the handler checks whether Dropshare is running. If not, it launches Dropshare in the background with `open -g -b <bundle id>`, trying `net.mkswap.Dropshare5` and then `net.mkswap.Dropshare-setapp`. It then waits 3 s. The official MCP server uses the same wait. An upload sent to a Dropshare that is still starting failed during the PoC with "error 0".

### D8. Container command and exit codes

`asylum-dropshare` is a `sh` script installed by the kit's Docker snippet, following the pattern of `asylum-open`. It sends the absolute path with `curl --max-time 330` and maps HTTP status to exit codes:

| Status | Exit |
|---|---|
| 200 | `0`, body is the URL |
| 504 | `2`, timeout |
| other | `1`, body is the error |

The route stops waiting when `r.Context()` is done, so an interrupted command releases the wait (spec: "Container session ends mid-wait").

### D9. Timeout

The route waits 5 min, which covers uploads of a few hundred MB on a typical uplink. The Claude Bash tool defaults to 2 min, so the skill tells the agent to run the command with a timeout of at least 330 000 ms.

### D10. Skill

The skill lives in `internal/kit/dropshare_skill.md` and is embedded with `go:embed`. The Docker snippet stages it, and the `asylum-dropshare` command, as base64 decoded with `base64 -d`, because the Dockerfile has no heredoc support. The skill lands at `/opt/asylum-skills/.claude/skills/dropshare/SKILL.md`. It carries over the guidance from the user's hand-written skill:

- Confirm before publishing.
- Never upload secrets.
- Rename files on disk instead of setting a display name.

It states that any file the container can see works, including `/tmp` and the scratchpad. It adds usage for GitHub, for example `![alt](url)` in `gh issue create --body`.

### D11. Uploads take turns

Tests on a real Mac showed that Dropshare drops the callback of an upload that starts while another one is still transferring:

| Situation | Result |
|---|---|
| Two uploads handed over in the same moment | One callback lost in 5 of 5 rounds |
| Handoffs 1–10 s apart, first transfer already finished | All 16 callbacks arrive |
| Second upload starts during a 50 MB or 100 MB transfer | Second callback lost in 2 of 2 runs |

The route takes a non-blocking `flock` on `~/.asylum/dropshare/.upload.lock` and polls it every 200 ms until it gets it, so a canceled request stops queuing. The kernel drops the lock when its holder exits, so a crashed broker can't leave it held.

The lock is released when the callback arrives, or `10 s + size / 1 MiB/s` after the handoff, whichever comes first. The hold starts once `open` has started. The cap stops a Dropshare failure that never calls back from blocking everyone for five minutes. Measured throughput was 8–19 MB/s, so 1 MiB/s leaves a wide margin. An upload slower than that lets the next one start early, and the next one then loses its callback and exits `2`. Its link still lands on the clipboard.

The five-minute limit covers queuing too. If the lock never frees up, the route answers 503 and the command exits `1`, because nothing was uploaded.

- **Alternative: a fixed 15 s cap.** Uploads longer than 15 s would bring the drops back.
- **Alternative: hold for the full timeout.** One Dropshare failure would block every upload for five minutes.

## Risks / Trade-offs

- **Every upload copies the file once**, costing disk space and time in proportion to its size. Large uploads are rare, and a 100 MB copy takes well under a second on an SSD.

- **Dropshare trusts the terminal, not Asylum.** The caller it finds is the terminal the broker descends from, usually iTerm2. If the session that spawned the broker exits while the container keeps running, the broker loses that ancestry and every upload prompts. Sending the upload from the applet would fix both, and needs a PoC first.

- **Dropshare fails without calling back** (upload error, connection chooser left open). The caller waits out the 5 min timeout. The timeout message tells the agent to ask the user.
- **Dropshare changes its callback encoding.** The table-driven parse tests pin the observed format, and a mismatch between `file` and the requested path fails loudly instead of returning a wrong URL.
- **Another app registers `asylum-dropshare:`.** Callbacks would stop arriving and uploads would time out. `lsregister -f` runs on every rebuild, and the docs describe the manual fix.
- **Focus.** Verified on a Mac: with `open -W -g` and `LSUIElement`, the terminal keeps focus during uploads, including when Dropshare has to be launched first. Dropshare's upload notification appears without taking focus.

## Migration Plan

The kit is opt-in, so nothing changes for existing users. To remove the host side, run `lsregister -u ~/.asylum/dropshare/AsylumDropshare.app` and delete `~/.asylum/dropshare/`. The docs page lists both steps.

## Open Questions

- Does Dropshare call the callback when an upload fails? If it does, the handler can fail fast instead of timing out. This only shortens failure latency.
