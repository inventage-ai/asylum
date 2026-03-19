## Why

Node.js projects built on the host (e.g., macOS) produce `node_modules` with OS-specific native binaries (esbuild, fsevents, sharp, etc.). When the project directory is mounted into the Linux container, these binaries fail to execute. Users must manually delete and reinstall `node_modules` inside the container every time.

## What Changes

- `node_modules` directories are automatically shadowed with named Docker volumes during container assembly. The host's `node_modules` becomes invisible inside the container, and dependencies installed in-container persist across restarts.
- The feature is on by default but can be disabled via `features: { shadow-node-modules: false }` in config.
- A `FeatureOff()` config helper is added for features that default to on (complement to the existing `Feature()` for opt-in features).

## Capabilities

### New Capabilities

- `shadow-node-modules`: Detect and shadow `node_modules` directories with named Docker volumes during container assembly.

### Modified Capabilities

- `container-assembly`: Volume assembly gains node_modules shadow logic.
- `config-loading`: New `FeatureOff()` method for default-on features.

## Impact

- **Container assembly**: Additional `--mount` flags appended for each `node_modules` found.
- **Docker volumes**: Named volumes created per project per `node_modules` path (e.g., `asylum-a1b2c3d4e5f6-npm-16b61a18f68`).
- **Config**: No new fields — uses existing `features` map.
- **Performance**: `findNodeModules` walks the project tree but short-circuits on missing `package.json` and skips irrelevant directories.
