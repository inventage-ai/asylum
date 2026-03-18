## Why

Projects using mise or asdf have a `.tool-versions` file specifying language versions. When asylum's container has a different Java patch version than what `.tool-versions` asks for, mise warns about missing versions on every container start. Users shouldn't need to duplicate version config in both `.tool-versions` and `.asylum`.

## What Changes

- Read `.tool-versions` from the project directory during config loading
- Use the Java version from `.tool-versions` as `versions.java` if not already set by asylum config or CLI flags
- This flows through the existing machinery: pre-installed versions are selected at runtime, others are installed via the project Dockerfile

## Capabilities

### New Capabilities

### Modified Capabilities

- `config-loading`: Config loading reads `.tool-versions` as an additional source for the `versions` map

## Impact

- `internal/config/config.go`: Parse `.tool-versions` and merge Java version at lowest priority (below all asylum config files and CLI flags)
