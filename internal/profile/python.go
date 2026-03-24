package profile

func init() {
	Register(&Profile{
		Name:        "python",
		Description: "Python tools via uv",
		DockerSnippet: `# Setup Python tools
RUN $HOME/.local/bin/uv tool install black && \
    $HOME/.local/bin/uv tool install ruff && \
    $HOME/.local/bin/uv tool install mypy && \
    $HOME/.local/bin/uv tool install pytest && \
    $HOME/.local/bin/uv tool install ipython && \
    $HOME/.local/bin/uv tool install poetry && \
    $HOME/.local/bin/uv tool install pipenv
`,
		BannerLines: `    echo "Python:    $(python3 --version 2>&1 | cut -d' ' -f2) (uv available)"
`,
		SubProfiles: map[string]*Profile{
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
				CacheDirs: map[string]string{"pip": "/home/claude/.cache/pip"},
			},
		},
	})
}
