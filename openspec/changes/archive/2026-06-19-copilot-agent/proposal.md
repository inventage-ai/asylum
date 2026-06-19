# Proposal: Copilot agent

## Why

Add first-class support for GitHub Copilot CLI as an asylum agent to provide feature parity with the existing Claude integration and enable users who prefer Copilot to run the CLI inside Asylum containers.

## What Changes

- Add a new agent implementation: internal/agent/copilot.go implementing the Agent interface.
- Register a Docker install snippet for the copilot CLI and add a banner line to the image build output.
- Implement session detection (HasSession) for copilot's persisted chat data and a Command() wrapper supporting relevant flags (experimental, model selection, LSP integration).
- Document Copilot-specific onboarding: auth (GH_TOKEN), LSP server config (.github/lsp.json), and MCP server wiring.
- Add specs for Copilot session handling and MCP server support under openspec/specs/copilot-session and openspec/specs/copilot-mcp.

## Capabilities

### New Capabilities
- `copilot-agent`: Provide a first-class Asylum agent adapter for the GitHub Copilot CLI. Covers install in images, session detection, command invocation, env var wiring, and onboarding docs.
- `copilot-session-detection`: Requirements for detecting and resuming copilot sessions from configuration or storage.
- `copilot-mcp-integration`: Define how Asylum should expose or configure MCP servers for Copilot to consume (mounts, env, flags).

### Modified Capabilities
- (none)

## Impact

- Affected code: `internal/agent/*` (new file copilot.go), `internal/agent/install.go` (install registry), docs/ and CLAUDE.md (new copilot notes), openspec/ (new specs), and Makefile/tests (unit & integration tests).
- Dependencies: copilot CLI install script (external), potential Node/npm if chosen install path uses npm.
- Systems: Onboarding and credential mounting systems (for GH_TOKEN) and MCP wiring.
