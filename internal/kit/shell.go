package kit

func init() {
	Register(&Kit{
		Name:           "shell",
		Description:    "Shell configuration (oh-my-zsh, tmux, direnv hooks)",
		DockerPriority: 32,
		Tier:           TierAlwaysOn,
		Tools:       []string{"tmux"},
		ConfigSnippet: `  # shell:              # Custom Dockerfile/entrypoint steps (on by default)
  #   build:             # Commands run at image build time
  #     - "curl -fsSL https://example.com/install.sh | sh"
  #   entrypoint:        # Commands run at container start
  #     - "source ~/.nvm/nvm.sh"
`,
		ConfigComment: "shell:                # Custom Dockerfile/entrypoint steps (on by default)\n  build:              # Commands run at image build time\n    - \"curl -fsSL https://example.com/install.sh | sh\"\n  entrypoint:         # Commands run at container start\n    - \"source ~/.nvm/nvm.sh\"",
		DockerSnippet: `# Install oh-my-zsh and setup PATH/fnm/mise for zsh
# oh-my-zsh replaces .zshrc, so PATH must be re-added after install
RUN sh -c "$(wget -O- https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" "" --unattended && \
    sed -i 's/ZSH_THEME=".*"/ZSH_THEME="robbyrussell"/' ~/.zshrc && \
    echo 'export PATH="$HOME/.local/share/fnm:$HOME/.local/bin:$PATH"' >> ~/.zshrc && \
    echo 'eval "$(fnm env)"' >> ~/.zshrc && \
    echo 'eval "$(mise activate zsh)"' >> ~/.zshrc

# Setup direnv hooks
RUN echo 'eval "$(direnv hook bash)"' >> ~/.bashrc && \
    echo 'eval "$(direnv hook zsh)"' >> ~/.zshrc

# Terminal size handling
RUN cat >> ~/.zshrc <<'STTYEOF'

if [[ -n "$PS1" ]] && command -v stty >/dev/null; then
  function _update_size {
    local rows cols
    { stty size } 2>/dev/null | read rows cols
    ((rows)) && export LINES=$rows COLUMNS=$cols
  }
  TRAPWINCH() { _update_size }
  _update_size
fi
STTYEOF

# Setup tmux configuration
RUN cat > ~/.tmux.conf <<'TMUXEOF'
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
TMUXEOF
`,
	})
}
