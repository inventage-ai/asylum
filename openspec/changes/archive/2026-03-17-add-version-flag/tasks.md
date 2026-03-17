## 1. Flag Parsing

- [x] 1.1 Add `Version` field to the `Flags` struct in `cmd/asylum/main.go`
- [x] 1.2 Handle `--version` in `parseArgs` to set `flags.Version = true`

## 2. Dispatch

- [x] 2.1 Add version dispatch in `main()` after the help check: print `asylum <version>` and exit

## 3. Usage & Tests

- [x] 3.1 Add `--version` to the usage/help text in `printUsage`
- [x] 3.2 Add test case for `--version` flag parsing in `parseArgs`
