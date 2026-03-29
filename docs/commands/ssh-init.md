# ssh-init

Set up SSH keys for use inside containers.

## Usage

```
asylum ssh-init
```

## Description

Creates `~/.asylum/ssh/` with an Ed25519 key pair for use inside containers. The SSH directory is automatically mounted into every container at `/home/claude/.ssh/`.

Specifically, `ssh-init`:

1. Creates `~/.asylum/ssh/` (mode 0700)
2. Copies `~/.ssh/known_hosts` into `~/.asylum/ssh/known_hosts` (merges and deduplicates if the file already exists)
3. Generates an Ed25519 key pair at `~/.asylum/ssh/id_ed25519` (if no key exists)
4. Prints the public key with instructions to add it to GitHub/GitLab

## After Setup

Add the printed public key to your Git hosting provider:

- **GitHub**: Settings > SSH and GPG keys > New SSH key
- **GitLab**: Preferences > SSH Keys > Add new key

You can also replace the generated key with your own — just put your key files in `~/.asylum/ssh/`.

## Notes

- Running `ssh-init` again is safe: it merges new `known_hosts` entries and skips key generation if a key already exists.
- The key comment is `asylum@<hostname>` to help identify it on your Git provider.
