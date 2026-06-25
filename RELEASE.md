# Release Process

This document explains how to create and publish releases for the secret-hoard project.

## Prerequisites

### Required Tools

1. **Go** - Build the binaries
   ```bash
   go version  # Should be 1.19 or later
   ```

2. **GitHub CLI (gh)** - Automate GitHub release creation
   ```bash
   gh --version
   # If not installed: https://cli.github.com/
   ```

3. **Git** - Version control
   ```bash
   git --version
   ```

### Required Permissions

- Write access to the GitHub repository
- Authenticated with GitHub CLI: `gh auth status`

## Release Workflow

### 1. Prepare for Release

Before creating a release, ensure the codebase is ready:

```bash
# Ensure working directory is clean
git status

# Run all static checks
make static

# Verify tests pass
go test ./...

# Commit any pending changes
git add .
git commit -m "Prepare for release vX.Y.Z"
git push origin main
```

### 2. Create Release

Use the `make semver-release` target to create a semantic version release:

```bash
make semver-release
```

This command will:
1. Prompt for a semantic version (e.g., `v1.0.0`, `v2.1.3`)
2. Verify git status is clean
3. Build binaries for all platforms
4. Create release tarballs with install scripts
5. Optionally create a GitHub release

**Interactive Prompts:**
```
Enter semver version (e.g., v1.0.0): v1.2.0
Creating release v1.2.0...
...
Create GitHub release? (y/N): y
Enter release title (default: v1.2.0): Release v1.2.0 - Enhanced sh-contents
Enter release notes (Ctrl-D when done):
## What's New
- Added post-write verification with SHA256 checksums
- Improved error messages with examples
- Minimal output mode for script-friendly usage
^D
```

## Semantic Versioning

Follow [Semantic Versioning 2.0.0](https://semver.org/):

- **Major version (vX.0.0)** - Incompatible API changes
  - Breaking changes to command-line interfaces
  - Removal of features
  - Changes to output format that break scripts

- **Minor version (v0.X.0)** - New features (backwards compatible)
  - New commands or flags
  - New secret types
  - Enhanced functionality
  
- **Patch version (v0.0.X)** - Bug fixes (backwards compatible)
  - Bug fixes
  - Documentation updates
  - Performance improvements

### Examples

```bash
# Major version - Breaking changes
v2.0.0 - Redesigned CLI interface

# Minor version - New features
v1.3.0 - Added sh-contents command

# Patch version - Bug fixes
v1.2.1 - Fixed SSL certificate permissions bug
```

## Release Artifacts

Each release creates the following artifacts:

### Directory Structure
```
release/
└── v1.2.0/
    └── secret-hoard_v1.2.0_linux_amd64.tar.gz
```

### Tarball Contents
```
secret-hoard_v1.2.0_linux_amd64.tar.gz
├── sh-pull            # Binary
├── sh-push            # Binary
├── sh-generate        # Binary
├── sh-contents        # Binary
└── install.sh         # Installation script
```

### Install Script

The included `install.sh` script:
- Installs binaries to `$HOME/bin`
- Makes them executable
- Checks if `$HOME/bin` is in PATH
- Provides instructions if not

**Usage:**
```bash
tar -xzf secret-hoard_v1.2.0_linux_amd64.tar.gz
cd secret-hoard_v1.2.0_linux_amd64
./install.sh
```

## Release Checklist

Use this checklist when creating releases:

- [ ] All tests passing (`make static`)
- [ ] Working directory clean (`git status`)
- [ ] Documentation updated
- [ ] CHANGELOG.md updated (if exists)
- [ ] Version follows semantic versioning
- [ ] Choose appropriate version number (major/minor/patch)
- [ ] Run `make semver-release`
- [ ] Enter version (e.g., `v1.2.0`)
- [ ] Verify tarballs created successfully
- [ ] Create GitHub release (if prompted)
- [ ] Write clear release notes
- [ ] Verify GitHub release published
- [ ] Test installation from release tarball
- [ ] Announce release (if applicable)

## Writing Release Notes

Good release notes include:

### Structure
```markdown
## What's New

- New feature 1
- New feature 2

## Improvements

- Enhancement 1
- Enhancement 2

## Bug Fixes

- Fix 1
- Fix 2

## Breaking Changes (if any)

- Breaking change 1 with migration instructions
```

### Best Practices

1. **User-focused** - Explain what changed for users, not implementation details
2. **Actionable** - Include migration steps for breaking changes
3. **Grouped** - Organize by category (features, fixes, improvements)
4. **Links** - Reference issues/PRs when relevant
5. **Examples** - Show usage examples for new features

### Example Release Notes

```markdown
## What's New in v1.3.0

### New Command: sh-contents
Download secret contents directly to files for use in scripts.
```bash
CONFIG=$(sh-contents jsondoc/dev/app-config /tmp)
cat "$CONFIG"
```

### Post-Write Verification
All file writes now include automatic SHA256 checksum verification
to detect corruption, hardware failures, and filesystem issues.

### Enhanced Error Messages
Error messages now include:
- Expected format with examples
- Possible causes
- Suggested remediation steps

## Improvements

- Reduced code duplication by 60%
- Improved test coverage to 89.4%
- Script-friendly minimal output mode

## Documentation

- Added RELEASE.md with release process
- Enhanced README with output modes
- Comprehensive CONTENTS.md tracking
```

## Checking Latest Release

To view the current published release:

```bash
make check-release
```

This shows:
- Version tag
- Release title
- Author
- Creation date
- Release notes
- GitHub URL

## Manual Release Process (Without gh CLI)

If you don't have GitHub CLI installed:

1. Run `make semver-release` (will skip GitHub release creation)
2. Manually create release on GitHub:
   - Go to: https://github.com/natemarks/secret-hoard/releases/new
   - Tag version: `v1.2.0`
   - Release title: `Release v1.2.0`
   - Upload tarballs from `release/v1.2.0/`
   - Write release notes
   - Publish release

## Troubleshooting

### Error: Working directory is dirty

```bash
Error - working directory is dirty. Commit those changes!
```

**Solution:** Commit or stash your changes:
```bash
git status
git add .
git commit -m "Your commit message"
# OR
git stash
```

### Error: gh CLI not installed

```bash
Note: Install gh CLI to create GitHub releases automatically
```

**Solution:** Install GitHub CLI:
```bash
# macOS
brew install gh

# Linux
# See: https://cli.github.com/
```

### Error: Not authenticated with GitHub

```bash
gh auth status
# Not logged into any GitHub hosts
```

**Solution:** Authenticate:
```bash
gh auth login
```

### Release tarball not created

**Check:**
- Build succeeded: `make build`
- Release directory exists: `ls release/`
- Disk space available: `df -h`

## Advanced Usage

### Building Without Release

To build binaries without creating a release:

```bash
make build
```

Binaries will be in: `build/$(git rev-parse HEAD)/linux/amd64/`

### Testing Release Locally

To test the release tarball locally:

```bash
# After running make semver-release
cd /tmp
tar -xzf ~/projects/secret-hoard/release/v1.2.0/secret-hoard_v1.2.0_linux_amd64.tar.gz
./install.sh
# Test installed binaries
sh-contents --help
```

## Support

For issues with the release process:
- Check the Makefile: `cat Makefile`
- Run with verbose output: `make semver-release -d`
- File an issue: https://github.com/natemarks/secret-hoard/issues

---

**Last Updated:** 2026-06-25  
**Process Owner:** Project Maintainers
