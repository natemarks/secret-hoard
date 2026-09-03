# secret-hoard — project conventions

## Release process

Releases are created with `make semver-release`. The target:
1. Builds statically-linked binaries for linux/amd64
2. Packages binaries into a versioned tarball (no install script inside the tarball)
3. Generates `checksums.txt` (aggregate sha256 of all tarballs)
4. Optionally publishes tarballs + checksums.txt to GitHub Releases via `gh`

## Installer scripts

Installer scripts that users run remotely (e.g. via `curl | bash`) live in the repo at a stable path — **not** inside release tarballs. The canonical location is `scripts/install.sh`.

This keeps the curl URL constant across versions:
```
https://raw.githubusercontent.com/natemarks/secret-hoard/main/scripts/install.sh
```

The script itself fetches the version-specific tarball and checksums from GitHub Releases, verifies the checksum, then installs.

## Binaries

The four released binaries are `sh-pull`, `sh-push`, `sh-generate`, `sh-contents`. All live under `cmd/`.
