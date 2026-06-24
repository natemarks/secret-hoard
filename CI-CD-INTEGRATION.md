# CI/CD Integration Documentation

## Overview

The project now has comprehensive static checks integrated into the development workflow:
- **Unit tests** run as part of `make static`
- **Pre-commit hook** available for local enforcement
- **GitHub Actions** workflow for CI/CD

---

## Make Target: `static`

The `make static` target now includes unit tests as the first check.

### What It Runs

```makefile
static: unittest goimports fmt vet lint gocyclo godeadcode govulncheck test
```

**Execution order:**
1. `unittest` - Unit tests (no external dependencies, ~3ms)
2. `goimports` - Go imports formatting
3. `fmt` - Go code formatting
4. `vet` - Go vet static analysis
5. `lint` - Golint checks
6. `gocyclo` - Cyclomatic complexity (max 25)
7. `godeadcode` - Dead code detection
8. `govulncheck` - Vulnerability scanning
9. `test` - All Go tests

### Usage

```bash
# Run all static checks (includes unit tests)
make static

# Just unit tests (fast)
make unittest

# Update golden files
make unittest-update

# Check coverage
make unittest-coverage
```

### Benefits

- **Fast feedback** - Unit tests run first (3ms)
- **Comprehensive** - All checks in one command
- **Consistent** - Same checks locally and in CI

---

## Pre-Commit Hook

A git pre-commit hook script is available to run `make static` automatically before every commit.

### Installation

```bash
# Install the hook
./scripts/git-commithook.sh install

# Check if installed
./scripts/git-commithook.sh status

# Uninstall
./scripts/git-commithook.sh uninstall

# Show help
./scripts/git-commithook.sh help
```

### What Happens

When you run `git commit`, the hook:
1. Runs `make static` from repository root
2. Shows colored output with progress
3. **Blocks the commit** if any check fails
4. **Allows the commit** if all checks pass

### Example Output

```bash
$ git commit -m "Add feature"

Running pre-commit checks (make static)...

=== RUN   TestBuildSecretID
--- PASS: TestBuildSecretID (0.00s)
...

✓ All static checks passed

[main abc1234] Add feature
 1 file changed, 10 insertions(+)
```

### Bypassing the Hook

**Not recommended**, but you can bypass with:
```bash
git commit --no-verify
```

### Hook Features

- ✅ Backs up existing hooks before installing
- ✅ Shows clear error messages
- ✅ Runs from repository root (works from any subdirectory)
- ✅ Colored output for visibility
- ✅ Easy to uninstall

---

## GitHub Actions Workflow

A GitHub Actions workflow runs `make static` on every push and pull request.

### Configuration

**File:** `.github/workflows/static-checks.yml`

**Triggers:**
- Push to `main`, `master`, or `develop` branches
- Pull requests to these branches

### What It Does

1. **Checkout code**
2. **Set up Go** (version 1.21)
3. **Cache dependencies** for faster builds
4. **Install tools** (goimports, golint, gocyclo, deadcode, govulncheck)
5. **Run `make static`**
6. **Upload artifacts** (test results, logs)
7. **Generate summary** in GitHub UI

### Viewing Results

**In GitHub UI:**
1. Go to **Actions** tab
2. Click on the workflow run
3. See check results and logs

**In Pull Requests:**
- Status check appears automatically
- Red X = checks failed
- Green checkmark = checks passed

### Example

```
✓ Static Checks / Run Static Checks
  All checks passed
```

### Customization

To change Go version, edit `.github/workflows/static-checks.yml`:
```yaml
- name: Set up Go
  uses: actions/setup-go@v5
  with:
    go-version: '1.22'  # Change version here
```

---

## Development Workflows

### Quick Development Loop

```bash
# Make changes
vim secretlogic/secretid.go

# Run just unit tests (fast)
make unittest

# If tests fail, fix and retry
make unittest

# Once working, run all checks
make static
```

### Before Committing

**Option 1: Manual**
```bash
make static
git add .
git commit -m "Your message"
```

**Option 2: With Pre-Commit Hook**
```bash
# Install once
./scripts/git-commithook.sh install

# Then just commit normally
git add .
git commit -m "Your message"
# Hook runs automatically
```

### Updating Golden Files

```bash
# After intentional logic changes
make unittest-update

# Review changes
git diff secretlogic/testdata/

# Commit both code and golden files
git add secretlogic/
git commit -m "Update metadata structure"
```

---

## CI/CD Pipeline

### Local → CI Flow

```
┌─────────────────┐
│ Developer       │
│ makes changes   │
└────────┬────────┘
         │
         ├─ make unittest (fast feedback)
         │
         ├─ make static (full checks)
         │
         ├─ git commit (pre-commit hook runs make static)
         │
         ├─ git push
         │
         └─────────────────┐
                           │
                  ┌────────▼────────┐
                  │ GitHub Actions  │
                  │ runs make static│
                  └────────┬────────┘
                           │
                  ┌────────▼────────┐
                  │ Status check    │
                  │ on PR/commit    │
                  └─────────────────┘
```

### Stages

1. **Local development** - `make unittest` (3ms)
2. **Before commit** - `make static` (via hook or manual)
3. **On push** - GitHub Actions runs `make static`
4. **PR review** - Status check must pass

---

## Troubleshooting

### Pre-Commit Hook Fails

**Symptom:** Commit is blocked

**Solution:**
```bash
# See what failed
make static

# Fix the issue
vim <file>

# Retry
git commit -m "Your message"
```

### GitHub Actions Fails

**Symptom:** Red X on commit/PR

**Solution:**
1. Click on the failing check
2. View logs to see which step failed
3. Run same check locally: `make static`
4. Fix and push again

### Hook Not Running

**Check status:**
```bash
./scripts/git-commithook.sh status
```

**Reinstall:**
```bash
./scripts/git-commithook.sh install
```

### Golden File Mismatches

**Symptom:** Unit tests fail with JSON differences

**If change is intentional:**
```bash
make unittest-update
git add secretlogic/testdata/
git commit -m "Update expected outputs"
```

**If change is unintentional:**
- Fix the code to match expected output
- Run `make unittest` to verify

---

## Best Practices

### ✅ Do

- Run `make unittest` frequently during development (fast)
- Run `make static` before pushing
- Install the pre-commit hook for automatic checks
- Review golden file diffs carefully in PRs
- Update golden files only when changes are intentional

### ❌ Don't

- Don't use `--no-verify` to bypass hooks (unless emergency)
- Don't commit without running static checks
- Don't update golden files without understanding why they changed
- Don't ignore failing GitHub Actions checks

---

## Metrics & Monitoring

### Local Checks

```bash
# Time various checks
time make unittest          # ~3ms
time make unittest-coverage # ~5ms
time make static            # ~30s (full suite)
```

### Coverage

```bash
make unittest-coverage
```

**Current:** 86.9% coverage of secretlogic package

### CI Performance

Check GitHub Actions run time in the Actions tab. Typical:
- Unit tests: <1s
- Full static checks: ~30-60s (depending on caching)

---

## Configuration Files

### Makefile Targets

```makefile
unittest              # Unit tests only
unittest-update       # Update golden files
unittest-coverage     # With coverage report
static                # All checks (includes unittest)
```

### Git Hook Script

**Location:** `scripts/git-commithook.sh`

**Commands:** install, uninstall, status, help

### GitHub Actions

**Location:** `.github/workflows/static-checks.yml`

**Triggers:** Push/PR to main/master/develop

---

## Integration Summary

| Check | Local | Hook | CI |
|-------|-------|------|-----|
| Unit tests | ✓ | ✓ | ✓ |
| Formatting | ✓ | ✓ | ✓ |
| Linting | ✓ | ✓ | ✓ |
| Vet | ✓ | ✓ | ✓ |
| Cyclomatic complexity | ✓ | ✓ | ✓ |
| Dead code | ✓ | ✓ | ✓ |
| Vulnerabilities | ✓ | ✓ | ✓ |
| All tests | ✓ | ✓ | ✓ |

**Consistency:** Same checks everywhere ensures no surprises in CI.

---

## Quick Reference

```bash
# Development
make unittest                              # Fast unit tests
make static                                # All checks

# Pre-commit hook
./scripts/git-commithook.sh install        # Set up hook
./scripts/git-commithook.sh status         # Check if installed

# Commit workflow
make static                                # Run checks
git add .
git commit -m "message"                    # Hook runs automatically
git push                                   # Triggers CI

# Golden files
make unittest-update                       # Update after changes
git diff secretlogic/testdata/             # Review changes
```

---

## Future Enhancements

### Potential Additions

1. **Pre-push hook** - Run on `git push` (slower checks)
2. **Code coverage enforcement** - Fail if coverage drops below threshold
3. **Benchmark tests** - Track performance over time
4. **Integration tests in CI** - With AWS mocks
5. **Release automation** - Build and publish on tag

### Adding More Tests

See `TESTING.md` for guidance on:
- Adding table-driven test cases
- Creating new golden file tests
- Testing additional business logic

---

## Support

### Documentation

- `TESTING.md` - Comprehensive testing guide
- `TEST-IMPLEMENTATION-SUMMARY.md` - Test overview
- `MANUAL-TEST.md` - Manual testing procedures
- `CI-CD-INTEGRATION.md` - This document

### Getting Help

```bash
# Show all make targets
make help

# Show hook commands
./scripts/git-commithook.sh help

# Run specific test
go test ./secretlogic -run TestBuildSecretID -v
```

---

## Summary

✅ **Unit tests integrated** into `make static`  
✅ **Pre-commit hook script** available for local enforcement  
✅ **GitHub Actions workflow** for CI/CD  
✅ **Consistent checks** across local, hook, and CI  
✅ **Fast feedback** with 3ms unit tests  
✅ **Easy setup** with one-command hook installation  

The development workflow now has comprehensive quality gates at every stage!
