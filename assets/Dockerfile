# Asylum - Multi-language development environment for AI coding agents
FROM debian:trixie

# Prevent interactive prompts during installation (build-time only)
ARG DEBIAN_FRONTEND=noninteractive
ENV LANG=en_US.UTF-8
ENV LANGUAGE=en_US:en
ENV LC_ALL=en_US.UTF-8

# Install system dependencies and essential tools
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    rm -f /etc/apt/apt.conf.d/docker-clean && \
    echo 'Binary::apt::APT::Keep-Downloaded-Packages "true";' > /etc/apt/apt.conf.d/keep-cache && \
    apt-get update && \
    apt-get install -y --no-install-recommends \
        # Essential tools
        ca-certificates curl wget gnupg lsb-release sudo \
        # Development tools
        git vim nano tmux htop tree direnv \
        # Build tools
        build-essential gcc g++ make cmake pkg-config \
        # Shell and utilities
        zsh bash-completion locales \
        # Network tools
        openssh-client netcat-openbsd socat dnsutils iputils-ping \
        # Archive tools
        zip unzip tar gzip bzip2 xz-utils \
        # JSON/YAML tools
        jq yq \
        # Process management
        procps psmisc \
        # Python build dependencies
        python3-dev python3-pip python3-venv \
        libssl-dev libffi-dev \
        # Java dependencies
        maven \
        # Search tools
        ripgrep fd-find && \
    # Setup locale
    echo "en_US.UTF-8 UTF-8" > /etc/locale.gen && \
    locale-gen && \
    # Cleanup
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Install Docker (full engine for Docker-in-Docker support)
RUN curl -fsSL https://download.docker.com/linux/debian/gpg | \
    gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg && \
    chmod 644 /usr/share/keyrings/docker-archive-keyring.gpg && \
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/debian $(lsb_release -cs) stable" \
    > /etc/apt/sources.list.d/docker.list && \
    apt-get update && \
    apt-get install -y docker-ce docker-buildx-plugin docker-compose-plugin && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Install GitHub CLI
RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | \
    gpg --dearmor -o /usr/share/keyrings/githubcli-archive-keyring.gpg && \
    chmod 644 /usr/share/keyrings/githubcli-archive-keyring.gpg && \
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" \
    > /etc/apt/sources.list.d/github-cli.list && \
    apt-get update && \
    apt-get install -y gh && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Install GitLab CLI
RUN ARCH=$(dpkg --print-architecture) && \
    GLAB_VERSION=$(curl -sL "https://gitlab.com/api/v4/projects/34675721/releases/permalink/latest" | sed -n 's/.*"tag_name":"v\?\([^"]*\)".*/\1/p') && \
    [ -n "${GLAB_VERSION}" ] || { echo "ERROR: could not determine glab version"; exit 1; } && \
    echo "Installing glab version ${GLAB_VERSION} for ${ARCH}" && \
    curl -fsSL -o /tmp/glab.deb \
        "https://gitlab.com/gitlab-org/cli/-/releases/v${GLAB_VERSION}/downloads/glab_${GLAB_VERSION}_linux_${ARCH}.deb" && \
    dpkg -i /tmp/glab.deb || apt-get install -f -y && \
    rm /tmp/glab.deb && \
    glab --version

# Create non-root user
ARG USER_ID=1000
ARG GROUP_ID=1000
ARG USERNAME=claude

RUN groupadd -g ${GROUP_ID} ${USERNAME} || true && \
    useradd -m -u ${USER_ID} -g ${GROUP_ID} -s /bin/zsh ${USERNAME} && \
    echo "${USERNAME} ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/${USERNAME} && \
    chmod 0440 /etc/sudoers.d/${USERNAME} && \
    usermod -aG docker ${USERNAME}

# Switch to user for language installations
USER ${USERNAME}
WORKDIR /home/${USERNAME}

# Install uv for Python package management
RUN curl -LsSf https://astral.sh/uv/install.sh | sh && \
    echo 'export PATH="$HOME/.local/share/fnm:$HOME/.local/bin:$PATH"' >> ~/.bashrc && \
    echo 'export PATH="$HOME/.local/share/fnm:$HOME/.local/bin:$PATH"' >> ~/.zshrc

# Install Node.js via fnm
RUN curl -fsSL https://fnm.vercel.app/install | bash -s -- --skip-shell && \
    export PATH="$HOME/.local/share/fnm:$PATH" && \
    eval "$(fnm env)" && \
    fnm install --lts && \
    fnm default lts-latest

# Setup fnm and mise in bash only (zsh will be set up after oh-my-zsh)
RUN echo 'eval "$(fnm env)"' >> ~/.bashrc && \
    echo 'eval "$(mise activate bash)"' >> ~/.bashrc

# Install Node.js global packages
RUN bash -c 'export PATH="$HOME/.local/share/fnm:$PATH" && eval "$(fnm env)" && \
    npm install -g \
        typescript \
        @types/node \
        ts-node \
        eslint \
        prettier \
        nodemon \
        yarn \
        pnpm \
        @fission-ai/openspec@latest'

# Install Claude Code
RUN curl -fsSL https://claude.ai/install.sh | bash && \
    ~/.local/bin/claude --version

# Install Gemini CLI
RUN bash -c 'export PATH="$HOME/.local/share/fnm:$PATH" && eval "$(fnm env)" && npm install -g @google/gemini-cli'

# Install Codex
RUN bash -c 'export PATH="$HOME/.local/share/fnm:$PATH" && eval "$(fnm env)" && npm install -g @openai/codex'

# Install mise for Java/Gradle version management
RUN curl https://mise.run | sh && \
    ~/.local/bin/mise install java@temurin-17 java@temurin-21 java@temurin-25 && \
    ~/.local/bin/mise use --global java@temurin-21 && \
    ~/.local/bin/mise install gradle@latest && \
    ~/.local/bin/mise use --global gradle@latest

# Setup Python tools
RUN /home/${USERNAME}/.local/bin/uv tool install black && \
    /home/${USERNAME}/.local/bin/uv tool install ruff && \
    /home/${USERNAME}/.local/bin/uv tool install mypy && \
    /home/${USERNAME}/.local/bin/uv tool install pytest && \
    /home/${USERNAME}/.local/bin/uv tool install ipython && \
    /home/${USERNAME}/.local/bin/uv tool install poetry && \
    /home/${USERNAME}/.local/bin/uv tool install pipenv

# Install oh-my-zsh and setup fnm/mise for zsh
RUN sh -c "$(wget -O- https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" "" --unattended && \
    sed -i 's/ZSH_THEME=".*"/ZSH_THEME="robbyrussell"/' ~/.zshrc && \
    echo 'eval "$(fnm env)"' >> ~/.zshrc && \
    echo 'eval "$(mise activate zsh)"' >> ~/.zshrc

# Setup direnv hooks
RUN echo 'eval "$(direnv hook bash)"' >> ~/.bashrc && \
    echo 'eval "$(direnv hook zsh)"' >> ~/.zshrc

# Terminal size handling
RUN cat >> ~/.zshrc <<'EOF'

if [[ -n "$PS1" ]] && command -v stty >/dev/null; then
  function _update_size {
    local rows cols
    { stty size } 2>/dev/null | read rows cols
    ((rows)) && export LINES=$rows COLUMNS=$cols
  }
  TRAPWINCH() { _update_size }
  _update_size
fi
EOF

# Configure git
RUN git config --global init.defaultBranch main && \
    git config --global pull.rebase false

# Setup tmux configuration
RUN cat > ~/.tmux.conf <<'EOF'
set -g mouse on
set -g default-terminal "screen-256color"
set -ga terminal-overrides ",xterm-256color:Tc"
set -g history-limit 50000
bind | split-window -h
bind - split-window -v
unbind '"'
unbind %
bind r source-file ~/.tmux.conf \; display-message "Config reloaded!"
set -g status-bg black
set -g status-fg white
set -g status-left '#[fg=green]#H '
set -g status-right '#[fg=yellow]%Y-%m-%d %H:%M'
EOF

# Create workspace directory
RUN mkdir -p /home/${USERNAME}/workspace

# Switch back to root for entrypoint setup
USER root

# Copy entrypoint script
COPY --chmod=755 entrypoint.sh /usr/local/bin/entrypoint.sh

# Set working directory
WORKDIR /workspace

# Set the user for runtime
USER ${USERNAME}

# Entrypoint
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["/bin/zsh"]

ENV ASYLUM_DISABLE_FEEDBACK_SURVEY=1
ENV CLAUDE_CODE_DISABLE_FEEDBACK_SURVEY=1
