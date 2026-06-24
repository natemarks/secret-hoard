# Changes Summary

## Interactive Mode for sh-pull

`sh-pull` now supports interactive mode, matching the UX of `sh-generate`. Users no longer need to remember which flags are required for each secret type.

### Key Features

1. **Automatic Interactive Mode**: Run `sh-pull` without flags to be prompted
2. **Type-Specific Prompts**: Only asks for metadata required by the selected type
3. **Flag Mode Still Supported**: Use flags for scripting/automation

### Usage Examples

**Interactive Mode (New):**
```bash
sh-pull

# Prompts based on type:
# - jsondoc: Select type → Environment → Access
# - rdspostgres: Select type → Environment → Instance → Database → Access
# - ssl_certificate: Select type → Environment → Common Name
# - etc.
```

**Flag Mode (Original):**
```bash
sh-pull -type=jsondoc -env=dev -access=app
```

### Implementation

- Added `promptForConfig()` function to `cmd/sh-pull/config.go`
- Reuses `generate.PromptForChoice()` and `generate.PromptForString()` functions
- Automatically enters interactive mode if no `-type` flag provided
- Validates all input before proceeding

### Benefits

1. **Easier to Use**: No need to memorize flags for each type
2. **Fewer Errors**: Can't forget required flags
3. **Discoverable**: Shows available secret types and required fields
4. **Flexible**: Both interactive and scripted modes supported

---

## Working Directory Update

All three new executables now use `$HOME/.secret-hoard/` as the default working directory for file operations. This provides a predictable, centralized location for secret files.

### Key Changes

1. **Added `tools.GetWorkingDir()` function** (`tools/file.go`)
   - Returns `$HOME/.secret-hoard/` path
   - Creates directory automatically if it doesn't exist
   - Uses permissions 0755

2. **Updated `pull` package** (`pull/pull.go`)
   - All pull functions now write files to working directory
   - Files created: `$HOME/.secret-hoard/<filename>`
   - Debug logging shows working directory path

3. **Updated `generate` package** (`generate/generate.go`)
   - All generate functions create files in working directory
   - Scaffolding created: `$HOME/.secret-hoard/<filename>`

4. **Updated `push` config** (`cmd/sh-push/config.go`)
   - Resolves relative paths relative to working directory
   - Absolute paths work as-is
   - Example: `-metadata=jsondoc.dev.app.metadata.json` → `$HOME/.secret-hoard/jsondoc.dev.app.metadata.json`

### File Locations

#### Before (current directory)
```
./jsondoc.dev.app.metadata.json
./jsondoc.dev.app.contents.json
```

#### After (working directory)
```
$HOME/.secret-hoard/jsondoc.dev.app.metadata.json
$HOME/.secret-hoard/jsondoc.dev.app.contents.json
```

### Benefits

1. **Predictable Location**: Users always know where their secret files are
2. **Centralized Management**: All secrets in one directory
3. **Easier Backups**: Single directory to backup
4. **No File Pollution**: Doesn't clutter current working directory
5. **Easy Cleanup**: Delete entire `.secret-hoard` directory when done

### Usage Examples

#### Generate
```bash
sh-generate
# Files created in ~/.secret-hoard/
```

#### Pull (Interactive)
```bash
sh-pull
# Select type and provide metadata
# Downloads to ~/.secret-hoard/
```

#### Pull (Flag Mode)
```bash
sh-pull -type=jsondoc -env=dev -access=app
# Downloads to ~/.secret-hoard/jsondoc.dev.app.*
```

#### Push (relative path)
```bash
sh-push -metadata=jsondoc.dev.app.metadata.json
# Resolves to ~/.secret-hoard/jsondoc.dev.app.metadata.json
```

#### Push (absolute path)
```bash
sh-push -metadata=/home/user/.secret-hoard/jsondoc.dev.app.metadata.json
# Uses absolute path as-is
```

### Backward Compatibility

- Absolute paths in `-metadata` flag still work
- Push command resolves relative paths to working directory
- No breaking changes to existing CSV-based workflows (sh-upload, sh-download)

### Directory Structure

```
$HOME/
└── .secret-hoard/
    ├── jsondoc.dev.app1.metadata.json
    ├── jsondoc.dev.app1.contents.json
    ├── textfile.prod.config.metadata.json
    ├── textfile.prod.config.contents.txt
    ├── sslcert.staging.example.com.metadata.json
    ├── sslcert.staging.example.com.crt
    ├── sslcert.staging.example.com.key
    ├── rdspostgres.dev.db1.app.ro.json
    └── snowflake.prod.analytics.dev.json
```

---

## Summary of All Changes

### New Features
1. ✅ Interactive mode for `sh-pull` (auto-detects when no flags provided)
2. ✅ Working directory `~/.secret-hoard/` for all file operations
3. ✅ Type-specific prompts in `sh-pull` matching secret requirements

### Files Modified
1. `tools/file.go` - Added GetWorkingDir() function
2. `pull/pull.go` - Updated all pull functions to use working directory
3. `generate/generate.go` - Updated all generate functions to use working directory
4. `cmd/sh-push/config.go` - Updated path resolution logic
5. `cmd/sh-pull/config.go` - Added interactive mode with type-specific prompts
6. `IMPLEMENTATION.md` - Added working directory note
7. `QUICK_START.md` - Updated examples with interactive mode and working directory paths
8. `MANUAL-TEST.md` - Added interactive mode tests and comprehensive testing guide

### Build Verification

✅ All executables compile successfully:
- `sh-pull` (12MB) - with interactive mode
- `sh-push` (12MB)
- `sh-generate` (12MB)

✅ Static checks pass:
- `go fmt`
- `go vet`
- `go build`

### Testing

See `MANUAL-TEST.md` for comprehensive testing guide covering:
- Interactive mode for sh-pull (Test 6)
- All 5 secret types with complete workflows
- Working directory verification
- Error conditions
- Integration tests
