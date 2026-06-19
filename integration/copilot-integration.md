# Copilot integration test plan

Purpose: manual verification steps to validate Copilot auth and session behavior inside Asylum containers.

Prereqs:
- Docker installed and running
- Copilot CLI install available inside image (image built with copilot install snippet)
- A personal GH_TOKEN with Copilot Requests permission for non-interactive flows

Steps:
1. Build test image: `make build` or `make build-all` and ensure Dockerfile includes Copilot install snippet.
2. Start Asylum for the project: `asylum shell` (or `asylum` to start agent) with GH_TOKEN mounted via config or env.
3. Verify copilot binary: `copilot --version`
4. Test interactive login: run `copilot` and perform `/login` if interactive environment available. Verify `~/.copilot` populated.
5. Test token auth: set GH_TOKEN in container and run a simple `copilot` command that lists sessions or uses `/model` to ensure auth works.
6. Create a session and verify files under `~/.copilot` (note paths). Document discovered paths and update spec `openspec/changes/copilot-agent/specs/copilot-session/spec.md`.
7. Attempt resume: if copilot supports a resume flag, test command and record behavior; otherwise document how to re-open prior session interactively.
8. Test LSP: place `.github/lsp.json` in repo and start container; verify `copilot` picks up LSP servers per config.

Notes:
- If interactive login fails inside container, prefer token-based mounting for CI runs.
- Record all observed paths/flags and add to the spec files.
