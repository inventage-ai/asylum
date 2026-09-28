# Dropshare Kit

Upload files through [Dropshare](https://dropshare.app) on your Mac and get the share URL back inside the container.

**Activation: Opt-in** — macOS with Dropshare 5 only. Asylum only offers it on a macOS host. On other hosts it is left out of the first-run wizard, the new-kit prompt, `asylum config`, and generated configs. Enabling it by hand still works as configured.

```yaml
kits:
  dropshare:
```

## What It Does

- Adds `asylum-dropshare <file>`. It uploads the file with Dropshare's default connection and prints the share URL.
- Adds a `dropshare` skill, so Claude knows when to upload, which files qualify, and how to embed links in GitHub issues and PRs.

```bash
url=$(asylum-dropshare ./shots/login-error.png)
gh issue create --title "Login error" --body "![Login error]($url)"
```

| Exit | Meaning |
|---|---|
| `0` | Uploaded. Stdout holds the URL. |
| `1` | Rejected or failed. Stderr holds the reason. |
| `2` | No link within 5 minutes. If the upload finishes, the link is on your clipboard. |

## Requirements

- Dropshare 5 (direct or Setapp) with a default connection configured.
- Integrations enabled in Dropshare: **Preferences → General → Integrations**.
- The file must be a regular file the container can see: in the project, another mounted directory, or container-local paths like `/tmp`.

## How It Works

1. The command sends the file's path to the host broker.
2. The broker copies the file out of the container with `docker cp` into a staging directory on the host. Docker resolves the path and its symlinks inside the container, so only what the container can see is reachable.
3. The broker hands the staged copy to Dropshare in the background, with a callback URL that carries a random request ID.
4. Dropshare opens the callback when the upload finishes. A small applet registered for `asylum-dropshare:` stores it in an inbox.
5. The broker picks up the callback for its request ID and returns the share URL.

Uploads from all projects take turns. Dropshare loses the callback of an upload that starts while another is still transferring. An upload keeps Dropshare to itself until its callback arrives, or for at most 10 s plus 1 s per MiB, so a failed upload cannot block the others for long.

Asylum never reads your clipboard. If Dropshare is not running, the broker starts it in the background first.

## Host State

Asylum creates `~/.asylum/dropshare/` on the first upload. One copy serves all projects.

| Path | Purpose |
|---|---|
| `AsylumDropshare.app` | Callback applet, registered with LaunchServices for `asylum-dropshare:` |
| `callback.sh` | Script the applet runs for each callback |
| `inbox/` | Callback records, removed once read or after one hour |
| `staging/` | Copies of files being uploaded, removed once the upload reports back or after one hour |

Asylum rebuilds the applet when a new release changes it.

To remove it:

```bash
/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister -u ~/.asylum/dropshare/AsylumDropshare.app
rm -rf ~/.asylum/dropshare
```

## Trust Boundary

- The container chooses only which file to upload, and only among files it can see. A symlink to a host file that isn't mounted into the container, such as one in `~/Documents`, resolves inside the container and never reaches the host file.
- Dropshare reads a staged copy the container cannot modify, so replacing the file after the request changes nothing.
- The container cannot set the connection, the display name, or the callback.
- Uploads publish files. The skill tells Claude to confirm before uploading anything you did not ask for, and never to upload secrets unconfirmed. The broker cannot tell a screenshot from a key file. Anything the container can read, it could already send elsewhere over its own network.

## Troubleshooting

- **Name clash with a personal skill** — if you have your own `~/.claude/skills/dropshare`, delete it, so Claude uses the kit's skill.
- **Every upload times out** — check that Integrations are enabled in Dropshare and that no other app claims `asylum-dropshare:`. Deleting `~/.asylum/dropshare/` forces a clean rebuild on the next upload.
