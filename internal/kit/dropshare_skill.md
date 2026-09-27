---
name: dropshare
description: Upload a file from the container to Dropshare on the user's Mac and get its share URL back. Use when the user wants to share, upload, or publish a local file — a screenshot, build artifact, log, PDF, image, or archive — or needs a link to embed in a GitHub issue, PR, or comment, or says "dropshare this", "upload this file", "give me a link for this file". Not for git pushes, deploys, or uploads to a named third-party service (S3, Netlify, Drive).
---

# Dropshare Upload

`asylum-dropshare` hands a file to the Dropshare app on the user's Mac, waits for the upload, and prints the share URL.

## Usage

```bash
asylum-dropshare ./shots/login-error.png
```

- Run it with a Bash timeout of at least `330000` ms. An upload can take up to 5 minutes, and the default 2-minute timeout would cut it off.
- On success it prints only the URL on stdout.
- It uploads one regular file at a time. Uploads from all projects take turns, so an upload may wait for another one to finish first.

## Exit codes

| Exit | Meaning | What to do |
|---|---|---|
| `0` | Uploaded. Stdout holds the URL. | Use the URL. |
| `1` | Rejected or failed. Stderr holds the reason. | Fix the cause (see below) or report it. If Dropshare stayed busy with another upload, try again later. |
| `2` | No link within 5 minutes. | The upload may still finish, and then the link will be on the user's clipboard. Ask the user for it. Do not retry right away, because a retry uploads the file a second time. |

## Which files work

Any regular file you can see works: the project, other mounted directories, `/tmp`, and the scratchpad. Symlinks resolve the way they do for you, inside the container. Dropshare uploads a copy taken at the moment you run the command.

- **Directories**: archive them first (`zip -r`, `tar czf`), then upload the archive.
- **Unhelpful names** (`out.png`, `tmp123.pdf`): rename the file on disk first. The share URL stays random whatever the name, and there is no display-name option.

## GitHub

`gh` cannot attach files to issues, so upload the file and link it instead:

```bash
url=$(asylum-dropshare ./shots/login-error.png)
gh issue create --title "Login error on submit" --body "![Login error]($url)"
```

## Constraints

- An upload publishes the file, and anyone with the link can open it. Treat it as an outward-facing action. If the user has not clearly asked to upload *this* file, confirm first.
- Private repositories: a Dropshare link in an issue is public even though the issue is not. Say so when you embed one.
- Never upload files containing credentials, tokens, `.env` contents, or private keys without the user confirming that specific file.
- Upload only what was asked for.
