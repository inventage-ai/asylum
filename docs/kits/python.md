# Python Kit

Python 3 with uv for fast package management, plus pre-installed development tools.

**Activation: Default** — added to config on first detection; active when present.

## What's Included

- **Python 3** (system package)
- **[uv](https://github.com/astral-sh/uv)** — fast Python package installer and resolver
- **Pre-installed tools**: black, ruff, mypy, pytest, ipython, poetry, pipenv
- **Build dependencies**: python3-dev, python3-pip, python3-venv, libssl-dev, libffi-dev

## Configuration

```yaml
kits:
  python:
    packages:          # additional tools to install via uv
      - pandas
      - numpy
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `packages` | list | `[]` | Additional Python packages to install via `uv tool install` (base image if declared in the global config, else the project image) |

## Sub-Kits

### python/uv

Included by default. Provides:

- **Pip cache** persisted at `~/.cache/pip`
- **Auto-venv**: automatically creates a `.venv/` virtual environment on container start if the project has `requirements.txt`, `pyproject.toml`, or `setup.py`

## Package Installation

Packages listed in the config are installed with `uv tool install` during the image build. For runtime dependencies in your project, use `uv pip install` or `pip install` inside the container.
