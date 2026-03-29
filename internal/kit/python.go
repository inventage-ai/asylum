package kit

func init() {
	Register(&Kit{
		Name:        "python",
		Description: "Python tools via uv",
		ConfigSnippet: `  python:
    # versions:
    #   - 3.14
    # packages:          # Python tools installed via uv
    #   - ansible
`,
		ConfigNodes:   configNodes("python", "", nil),
		ConfigComment: "versions:\n#   - 3.14\n# packages:          # Python tools installed via uv\n#   - ansible",
		DockerSnippet: `# Install Python build dependencies
USER root
RUN apt-get update && apt-get install -y --no-install-recommends \
    python3-dev python3-pip python3-venv libssl-dev libffi-dev && \
    rm -rf /var/lib/apt/lists/*
USER ${USERNAME}

# Setup Python tools
RUN $HOME/.local/bin/uv tool install black && \
    $HOME/.local/bin/uv tool install ruff && \
    $HOME/.local/bin/uv tool install mypy && \
    $HOME/.local/bin/uv tool install pytest && \
    $HOME/.local/bin/uv tool install ipython && \
    $HOME/.local/bin/uv tool install poetry && \
    $HOME/.local/bin/uv tool install pipenv
`,
		RulesSnippet: `### Python (python kit)
Python 3 with uv package manager. Pre-installed tools: black, ruff, mypy, pytest, ipython, poetry, pipenv. Use ` + "`uv`" + ` for fast package installation and virtual environment management.
`,
		BannerLines: `    echo "Python:    $(python3 --version 2>&1 | cut -d' ' -f2) (uv available)"
`,
		SubKits: map[string]*Kit{
			"uv": {
				Name:        "python/uv",
				Description: "Python venv auto-creation and pip caching",
				EntrypointSnippet: `# Create Python virtual environment if project has Python markers
has_python_marker() {
    for f in requirements.txt pyproject.toml setup.py; do
        [ -f "$HOST_PROJECT_DIR/$f" ] && return 0
    done
    return 1
}
if [ -n "$HOST_PROJECT_DIR" ] && [ ! -d "$HOST_PROJECT_DIR/.venv" ] && has_python_marker; then
    echo "Python project detected, creating virtual environment..."
    cd "$HOST_PROJECT_DIR"
    if uv venv .venv; then
        echo "Virtual environment created at .venv/"
        echo "  Activate with: source .venv/bin/activate"
    else
        echo "Warning: failed to create virtual environment (continuing)"
    fi
fi
`,
				CacheDirs: map[string]string{"pip": "~/.cache/pip"},
			},
		},
	})
}
