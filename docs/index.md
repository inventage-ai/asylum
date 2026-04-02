# Asylum

Docker sandbox for AI coding agents. Single binary, runs on Linux and macOS (ARM and x86).

Asylum wraps Docker to give [Claude Code](https://claude.ai) a full development environment with Python, Node.js, Java, and Docker-in-Docker — while keeping your host clean. Containers are ephemeral, but caches, auth, and history persist. Experimental support is available for [Gemini CLI](https://github.com/google-gemini/gemini-cli) and [Codex](https://github.com/openai/codex).

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/inventage-ai/asylum/main/install.sh | sh
```

Or download a binary from the [releases page](https://github.com/inventage-ai/asylum/releases) and put it in your PATH (you'll need to `chmod +x` it before use).

**Requires**: [Docker](https://docs.docker.com/get-docker/) installed and running.

## Quick Start

```sh
cd your-project/

# Start Claude Code (default)
asylum

# Start Gemini CLI
asylum -a gemini

# Start Codex
asylum -a codex

# Interactive shell (no agent)
asylum shell
```

On first run, Asylum builds a Docker image (~5 min) and seeds agent config from your host. Subsequent runs start in seconds.

See [Getting Started](getting-started.md) for a full walkthrough of your first run.
