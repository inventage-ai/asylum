---
name: install
description: Build and install asylum for the host platform from the current source tree. Use when the user wants to test local changes.
disable-model-invocation: true
---

Build asylum from the current source and install it to the host's `~/.asylum/bin/`.

**Steps**

1. **Detect the host platform**

   Detect the host OS and architecture:
   - Check `/proc/version` for "linuxkit" — if present, host is `darwin`, otherwise `linux`
   - Run `uname -m` — map `aarch64`/`arm64` to `arm64`, `x86_64`/`amd64` to `amd64`

2. **Get the current commit**

   Run `git rev-parse --short HEAD` to get the commit hash.

3. **Build the binary**

   Cross-compile for the host platform:
   ```bash
   GOOS=<os> GOARCH=<arch> go build -ldflags "-X main.version=local -X main.commit=<commit>" -o build/asylum ./cmd/asylum
   ```

4. **Install**

   Copy the binary to `~/.asylum/bin/asylum`:
   ```bash
   cp build/asylum ~/.asylum/bin/asylum
   ```

   The host's install script already set up the symlink from `/usr/local/bin/asylum` to `~/.asylum/bin/asylum`, so no symlink management needed.

5. **Verify**

   Run `~/.asylum/bin/asylum --version` to confirm the build.

6. **Report**

   Show the version string and confirm installation.
