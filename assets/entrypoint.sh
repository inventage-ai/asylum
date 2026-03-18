#!/bin/bash
# Asylum entrypoint script

set -e

# Start Docker daemon if enabled and running privileged
if [ "${ASYLUM_DOCKER:-}" = "1" ]; then
    if [ -S /var/run/docker.sock ] && docker info >/dev/null 2>&1; then
        echo "Docker socket already available"
    elif command -v dockerd >/dev/null 2>&1; then
        echo "Starting Docker daemon..."
        sudo dockerd --storage-driver=vfs --log-level=warn >/tmp/dockerd.log 2>&1 &
        DOCKERD_PID=$!
        disown $DOCKERD_PID
        for i in $(seq 1 30); do
            if docker info >/dev/null 2>&1; then
                echo "Docker daemon ready"
                break
            fi
            sleep 1
        done
        if ! docker info >/dev/null 2>&1; then
            echo "Warning: Docker daemon failed to start (check /tmp/dockerd.log)"
        fi
    fi
fi

# Ensure proper PATH
export PATH="$HOME/.local/share/fnm:$HOME/.local/bin:$PATH"

# Setup fnm if available
if command -v fnm >/dev/null 2>&1; then
    eval "$(fnm env --shell bash)"
fi

# Source SDKMAN if available, and select Java version if configured
if [ -f "$HOME/.sdkman/bin/sdkman-init.sh" ]; then
    source "$HOME/.sdkman/bin/sdkman-init.sh"
    if [ -n "${ASYLUM_JAVA_VERSION:-}" ]; then
        case "${ASYLUM_JAVA_VERSION}" in
            17|21|25)
                match=$(ls -d "$HOME/.sdkman/candidates/java/${ASYLUM_JAVA_VERSION}"*-tem 2>/dev/null | head -1)
                if [ -n "$match" ]; then
                    export JAVA_HOME="$match"
                    export PATH="$JAVA_HOME/bin:$PATH"
                else
                    echo "Warning: Java version matching '${ASYLUM_JAVA_VERSION}' not found. Installed:"
                    ls "$HOME/.sdkman/candidates/java/" 2>/dev/null || true
                fi
                ;;
            *)
                echo "Warning: ASYLUM_JAVA_VERSION '${ASYLUM_JAVA_VERSION}' is not a supported version (17, 21, 25). Ignoring."
                ;;
        esac
    fi
fi

# Create Python virtual environment if project has Python markers
if [ -n "$HOST_PROJECT_DIR" ] && [ ! -d "$HOST_PROJECT_DIR/.venv" ] && [ -f "$HOST_PROJECT_DIR/requirements.txt" -o -f "$HOST_PROJECT_DIR/pyproject.toml" -o -f "$HOST_PROJECT_DIR/setup.py" ]; then
    echo "Python project detected, creating virtual environment..."
    cd "$HOST_PROJECT_DIR"
    if uv venv .venv; then
        echo "Virtual environment created at .venv/"
        echo "  Activate with: source .venv/bin/activate"
    else
        echo "Warning: failed to create virtual environment (continuing)"
    fi
fi

# Set proper permissions on mounted SSH directory
if [ -d "/home/claude/.ssh" ]; then
    chmod 700 /home/claude/.ssh 2>/dev/null || true
    chmod 600 /home/claude/.ssh/* 2>/dev/null || true
    chmod 644 /home/claude/.ssh/*.pub 2>/dev/null || true
    chmod 644 /home/claude/.ssh/authorized_keys 2>/dev/null || true
    chmod 644 /home/claude/.ssh/known_hosts 2>/dev/null || true
fi

# Restore Claude config from backup if missing or incomplete (no auth)
if [ -n "${CLAUDE_CONFIG_DIR:-}" ]; then
    cfg="$CLAUDE_CONFIG_DIR/.claude.json"
    if [ ! -f "$cfg" ] || ! grep -q '"oauthAccount"' "$cfg" 2>/dev/null; then
        latest=$(find "$CLAUDE_CONFIG_DIR/backups" -maxdepth 1 -name '.claude.json.backup.*' -printf '%T@\t%p\n' 2>/dev/null \
            | sort -rn | head -1 | cut -f2-)
        if [ -n "$latest" ] && grep -q '"oauthAccount"' "$latest" 2>/dev/null; then
            cp "$latest" "$cfg"
            echo "Restored Claude config from backup"
        fi
    fi
fi

# Translate host direnv approvals to container paths
if [ -d "/tmp/host_direnv_allow" ] && [ -n "$HOST_PROJECT_DIR" ] && [ -f "$HOST_PROJECT_DIR/.envrc" ]; then
    mkdir -p /home/claude/.local/share/direnv/allow

    host_envrc_path="$HOST_PROJECT_DIR/.envrc"
    expected_host_hash=$(printf "%s\n" "$host_envrc_path" | cat - "$HOST_PROJECT_DIR/.envrc" | sha256sum | cut -d' ' -f1)

    if [ -f "/tmp/host_direnv_allow/$expected_host_hash" ]; then
        approved_path=$(cat "/tmp/host_direnv_allow/$expected_host_hash")
        if [ "$approved_path" = "$host_envrc_path" ]; then
            container_envrc="$HOST_PROJECT_DIR/.envrc"
            container_hash=$(printf "%s\n" "$container_envrc" | cat - "$container_envrc" | sha256sum | cut -d' ' -f1)
            echo "$container_envrc" > /home/claude/.local/share/direnv/allow/"$container_hash"
        fi
    fi
fi

# Set up git config
if [ -f "/tmp/host_gitconfig" ]; then
    cp /tmp/host_gitconfig /home/claude/.gitconfig
else
    cat > /home/claude/.gitconfig << 'EOF'
[user]
    email = claude@asylum
    name = Claude (Asylum)
[init]
    defaultBranch = main
EOF
fi

# Trust all mounted repositories (container is ephemeral, all mounts are user-chosen)
git config --global --add safe.directory '*'

# Ignore file mode changes (Docker Desktop mounts lose execute bits on file rewrites)
if grep -q linuxkit /proc/version 2>/dev/null; then
    git config --global core.fileMode false
fi

# Check for MCP configuration
if [ -n "$HOST_PROJECT_DIR" ] && { [ -f "$HOST_PROJECT_DIR/.mcp.json" ] || [ -f "$HOST_PROJECT_DIR/mcp.json" ]; }; then
    echo "MCP configuration detected."
fi

# Set terminal
export TERM=xterm-256color

# Handle terminal size
if [ -t 0 ]; then
    eval $(resize 2>/dev/null || true)
fi

# Welcome message for interactive sessions
if [ -t 0 ] && [ -t 1 ]; then
    echo "Asylum Development Environment"
    echo "================================"
    echo "Workspace: $(pwd)"
    echo "Python:    $(python3 --version 2>&1 | cut -d' ' -f2) (uv available)"
    echo "Node.js:   $(node --version 2>/dev/null || echo 'not found')"
    echo "Java:      $(java -version 2>&1 | head -1 | cut -d'"' -f2 || echo 'not found')"
    echo "Claude:    $(claude --version 2>/dev/null || echo 'not found')"
    echo "Gemini:    $(gemini --version 2>/dev/null || echo 'not found')"
    echo "Codex:     $(codex --version 2>/dev/null || echo 'not found')"
    echo "================================"
    echo ""
fi

# If dockerd is running, skip exec so the trap can kill it on exit.
# Otherwise, exec replaces this shell (saves a process).
if [ -n "${DOCKERD_PID:-}" ]; then
    trap 'sudo kill -9 $DOCKERD_PID 2>/dev/null' EXIT
    "$@"
else
    exec "$@"
fi
