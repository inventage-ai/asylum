# Tasks

## 1. Staging from the container

- [x] 1.1 Add `docker.CopyOut` and stage the requested file from its tar stream into `~/.asylum/dropshare/staging/<id>/` (D5); verify tests that a directory is refused without being copied, a missing file is rejected, and swapping the source for a symlink after the request leaves the staged content unchanged, and that the swap test fails when Dropshare is handed the original path
- [x] 1.2 Remove the staged copy when the callback arrives or the upload is rejected, and prune copies older than an hour; verify with handler and prune tests

## 2. Callback and URL handling

- [x] 2.1 Implement the callback parser (D6) with table-driven tests for a plain path, `a & b#1+c=ü%20.txt`, a path containing `&url=`, a missing `file=`, a non-https `url`, and a `file` mismatch; verify `go test ./internal/kit/...`
- [x] 2.2 Implement the outgoing upload URL encoder and verify a test round-trips the PoC filenames through encode and parse

## 3. Applet management

- [x] 3.1 Generate the applet source and plist values with the inbox path baked in, plus their hash; verify with a test that the hash changes when the inbox path or source changes and not otherwise
- [x] 3.2 Implement build-under-lock into a temp dir, rename, `lsregister -f`, and write the stamp last (D3, D4); verify on a Mac that a first upload builds the applet and a second upload skips the rebuild
- [x] 3.3 Implement inbox polling keyed by request ID, with removal of consumed records and records older than one hour; verify with a test that uses a temp inbox and a fake callback file

## 4. Route and kit

- [x] 4.1 Implement the upload route: reject non-darwin hosts, check the path, ensure the applet, launch Dropshare if it isn't running (D7), start `open -W -g` and kill it when the callback arrives or after 5 s, wait up to 5 min or until the request context ends, and return 200, 504, or an error; verify with handler tests that stub the opener and write callbacks into a temp inbox
- [x] 4.2 Register the opt-in `dropshare` kit with its route, the `asylum-dropshare` Docker snippet (D8), `ProvidesSkills`, the staged `SKILL.md` (D10), a rules snippet, and config nodes; verify `go test ./internal/kit/...` and that an assembled Dockerfile contains the command and skill
- [x] 4.4 Serialize uploads with a size-scaled hold (D11); verify handler tests prove uploads never overlap, a stuck upload releases after its hold, and a busy queue answers 503, and that the overlap test fails without the lock
- [x] 4.3 Write `SKILL.md` covering the confirm-before-publishing rule, the secrets rule, rename-on-disk, the scratchpad copy step, the 330 000 ms Bash timeout, exit codes, and GitHub embedding; verify the skill loads in a container session

## 5. End-to-end on macOS

- [x] 5.1 In a container with the kit enabled, upload a project file and verify the command prints a URL that serves the file's bytes
- [x] 5.2 Upload a file named `a & b#1+c=ü%20.txt` and verify the printed URL serves that file
- [x] 5.3 Run simultaneous uploads from two projects and verify each prints its own file's URL; the two-project races exposed the dropped callbacks (D11), and the serialized rerun from one broker uses the same cross-process `flock`, since each request opens the lock file separately
- [x] 5.4 Verify the terminal keeps focus during an upload; if the applet steals it, record the finding and adjust the applet before continuing
- [x] 5.5 Quit Dropshare, upload, and verify the kit launches Dropshare and the upload succeeds
- [x] 5.7 Rerun the simultaneous-pair and mid-transfer tests on the Mac and verify every upload prints its URL
- [x] 5.6 On the Mac, upload a scratchpad file and a file from another mount, and verify a symlink to a host-only canary file outside every mount is rejected instead of uploading the canary

## 6. Docs

- [x] 6.1 Add `docs/kits/dropshare.md` with enabling the kit, macOS-only scope, host state under `~/.asylum/dropshare/`, uninstall steps, and the name clash with a hand-written `~/.claude/skills/dropshare`; add it to `mkdocs.yml` and verify `mkdocs build` succeeds
- [x] 6.2 Add an **Added** entry to `CHANGELOG.md` under Unreleased
