## 1. Docker package: container detection

- [x] 1.1 Add `IsRunning(name string) bool` to `internal/docker/docker.go` using `docker inspect --format '{{.State.Running}}'`

## 2. Container package: exec args

- [x] 2.1 Export `ContainerName(projectDir string) string` (currently unexported `containerName`)
- [x] 2.2 Add `ExecArgs(containerName string, mode Mode, extraArgs []string) []string` that builds `docker exec -it` args, with `-u root` for admin shell

## 3. Main: dispatch logic

- [x] 3.1 In `main.go`, after resolving the container mode, check if mode is shell/run and container is running — if so, skip image build and use `ExecArgs` instead of `RunArgs`

## 4. Test and verify

- [x] 4.1 Add unit test for `ExecArgs` covering shell, admin shell, and run modes
- [x] 4.2 Manually verify: start agent, then `asylum shell` in second terminal execs into running container
