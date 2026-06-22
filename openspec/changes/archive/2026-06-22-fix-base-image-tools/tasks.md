## 1. Dockerfile changes

- [x] 1.1 Add `file` to the apt install list in `assets/Dockerfile.core` (under "Search tools" or a fitting group).
- [x] 1.2 Add a symlink so `fd` resolves to the `fdfind` binary (e.g. `ln -s "$(command -v fdfind)" /usr/local/bin/fd`), placed so it survives the apt cleanup in the same/late build stage.

## 2. Verification

- [x] 2.1 Verified the changed apt+symlink step in an isolated `debian:trixie` build: `rg --version` (14.1.1), `fd --version` (10.2.0), and `file --version` (5.46) all succeed. (Used an isolated build of the changed instruction rather than a full 5.6 GB base rebuild.)
- [x] 2.2 Confirmed `command -v fd` resolves to `/usr/local/bin/fd` (the symlink), which runs the `fdfind` binary.

## 3. Docs

- [x] 3.1 Add a `CHANGELOG.md` Unreleased entry under **Fixed** (fd now available under canonical name; file installed).
