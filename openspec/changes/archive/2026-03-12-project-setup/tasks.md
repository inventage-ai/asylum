## 1. Module and Dependencies

- [x] 1.1 Run `go mod init github.com/binaryben/asylum` and add `gopkg.in/yaml.v3` dependency
- [x] 1.2 Create directory structure: `cmd/asylum/`, `internal/{agent,config,container,docker,image,log,ssh}/`, `assets/`

## 2. Build System

- [x] 2.1 Create Makefile with `build`, `build-all`, `clean`, and `test` targets using `-ldflags -X` for version injection
- [x] 2.2 Update `.gitignore` to include `build/` directory

## 3. Entry Point

- [x] 3.1 Create `cmd/asylum/main.go` with version variable and print-version behavior
- [x] 3.2 Verify `make build` and `make build-all` produce working binaries
