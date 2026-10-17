# Tasks

## 1. Command line

- [x] 1.1 Move resolved `-e NAME=value` args that docker's env file parser accepts unchanged (`fitsEnvFile`) into an env file passed as `--env-file /dev/stdin` after all other options; `RunArgs` returns its content and `docker.RunDetached` feeds it to stdin. Remove `HostEnv` and the `DOCKER_*` special case from `env-from-host`. Verify with `TestRunArgsEnvViaStdin`, `TestSplitEnvArgs` and a manual `docker run -d --env-file /dev/stdin` check.

## 2. Debug output

- [x] 2.1 Mask env values in `RunArg.String()`, use it in `FormatDebug` and in the conflict error. Verify with `TestFormatDebugBasic` and a conflict test.

## 3. Documentation

- [x] 3.1 Update `docs/configuration/index.md` and CHANGELOG. Verify by reading the sections.

## 4. Verification

- [x] 4.1 Run `go vet ./...`, `go test ./...` and `openspec validate env-values-off-command-line --strict`; verify all pass.
