## Why

`mergeKitConfig` and `ConfigHash` both manually enumerate `KitConfig` fields. When a new field is added, both must be updated — and forgetting either produces a silent bug (credentials merge and hash were both missed by the `kit-credentials` change). This is a structural problem: an open set of fields with a closed enumeration.

## What Changes

- `mergeKitConfig` uses reflection to iterate `KitConfig` fields, reading a `merge:"concat"` struct tag to determine strategy. Default (no tag) is last-wins. Adding a field no longer requires merge code changes.
- `ConfigHash` serializes the full `Config` to YAML (deterministic map key ordering) after zeroing non-runtime fields, instead of manually serializing each field. New fields are included automatically.

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `deep-merge-kit-config`: Merge strategy is now declared via struct tags rather than manual field enumeration. Behavior is unchanged for all existing fields.
- `stale-container-detection`: Config hash now covers the full config (minus non-runtime fields) via YAML serialization instead of a hardcoded field list.

## Impact

- `internal/config/config.go` — `KitConfig` struct tags, `mergeKitConfig`, `ConfigHash`
- `internal/config/config_test.go` — updated hash test expectations
- New dependency on `reflect` (stdlib)
