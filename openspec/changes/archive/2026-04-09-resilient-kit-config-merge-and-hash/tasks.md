## 1. Tag-driven mergeKitConfig

- [x] 1.1 Add `merge:"concat"` struct tags to `Packages` and `Build` fields on `KitConfig`
- [x] 1.2 Replace manual field enumeration in `mergeKitConfig` with reflection loop
- [x] 1.3 Verify all existing `TestMergeKitConfig` cases still pass

## 2. YAML-based ConfigHash

- [x] 2.1 Replace hand-rolled `ConfigHash` serialization with `yaml.Marshal` after zeroing non-runtime fields
- [x] 2.2 Sort `Volumes` and `Ports` before marshaling for order independence
- [x] 2.3 Update "nil kit config" test expectation (declared kit differs from absent kit)
- [x] 2.4 Verify all existing `TestConfigHash` cases still pass
