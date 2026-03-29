# Node.js Kit

Node.js LTS via fnm with global development packages.

## What's Included

- **Node.js LTS** managed by [fnm](https://github.com/Schniz/fnm)
- **Global packages**: typescript, @types/node, ts-node, eslint, prettier, nodemon
- **Package managers** (via sub-kits): npm (built-in), pnpm, yarn

## Configuration

```yaml
kits:
  node:
    packages:                    # additional global npm packages
      - tsx
      - vitest
    shadow-node-modules: true    # default: true
    onboarding: true             # auto-detect and install deps
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `packages` | list | `[]` | Additional npm packages to install globally in the project image |
| `shadow-node-modules` | bool | `true` | Shadow `node_modules` with Docker volumes (see below) |
| `onboarding` | bool | `false` | Auto-detect lockfiles and run `npm install` on first container start |

## Sub-Kits

### node/npm

Included by default. Provides npm cache persistence at `/home/claude/.npm` and the npm onboarding task.

### node/pnpm

Installs pnpm globally.

### node/yarn

Installs yarn globally.

## Shadow node_modules

On macOS, Node.js native binaries built on the host won't work inside the Linux container. Asylum automatically shadows each `node_modules` directory with a named Docker volume so each platform has its own binaries. Your source files are shared — only `node_modules` is isolated.

Disable with:

```yaml
kits:
  node:
    shadow-node-modules: false
```

## Onboarding

When `onboarding: true` is set, Asylum scans for lockfiles on first container start and runs the appropriate install command:

| Lockfile | Command |
|----------|---------|
| `package-lock.json` | `npm ci` |
| `pnpm-lock.yaml` | `pnpm install --frozen-lockfile` |
| `yarn.lock` | `yarn install --frozen-lockfile` |
| `bun.lock` / `bun.lockb` | `bun install --frozen-lockfile` |

Onboarding state is tracked — it won't re-prompt unless a lockfile changes. Skip for a single run with `--skip-onboarding`.

## Version Switching

fnm is available inside the container. To switch Node versions:

```sh
fnm install 20
fnm use 20
```
