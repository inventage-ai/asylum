## 1. Kit rules snippet infrastructure

- [x] 1.1 Add `RulesSnippet string` field to `Kit` struct in `internal/kit/kit.go`
- [x] 1.2 Add `AssembleRulesSnippets` function in `internal/kit/kit.go` following the `AssembleDockerSnippets` pattern
- [x] 1.3 Add test for `AssembleRulesSnippets` in `internal/kit/kit_test.go`

## 2. Kit rules snippets

- [x] 2.1 Add `RulesSnippet` to the docker kit (`internal/kit/docker.go`)
- [x] 2.2 Add `RulesSnippet` to the java kit and relevant sub-kits (`internal/kit/java.go`)
- [x] 2.3 Add `RulesSnippet` to the node kit and relevant sub-kits (`internal/kit/node.go`)
- [x] 2.4 Add `RulesSnippet` to the python kit and relevant sub-kits (`internal/kit/python.go`)

## 3. Rules file generation and mounting

- [x] 3.1 Add `Kits []*kit.Kit` field to `RunOpts` in `internal/container/container.go`
- [x] 3.2 Implement `generateSandboxRules` function that writes the core template + assembled kit snippets to `~/.asylum/projects/<container>/sandbox-rules.md`
- [x] 3.3 Add the read-only file mount of the generated rules file at `<project>/.claude/rules/asylum-sandbox.md` in `RunArgs`
- [x] 3.4 Add test for rules file generation (core content + kit snippets)
- [x] 3.5 Add test that the mount arg appears in `RunArgs` output

## 4. Wire up in main

- [x] 4.1 Pass `allKits` to `RunOpts.Kits` in `cmd/asylum/main.go`

## 5. Changelog

- [x] 5.1 Add entry under Unreleased/Added in `CHANGELOG.md`
