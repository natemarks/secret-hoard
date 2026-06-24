# Architecture Documentation

## Overview

secret-hoard is organized into clear layers with well-defined responsibilities. This document explains the architecture, design patterns, and how the pieces fit together.

## Package Structure

```
secret-hoard/
├── cmd/                    # Command-line executables (entry points)
│   ├── sh-upload/         # Batch CSV upload
│   ├── sh-pull/           # Interactive/flag download
│   ├── sh-push/           # Upload with diff
│   └── sh-generate/       # Scaffolding generator
├── secrets/               # NEW: Core interfaces and operations
│   ├── interfaces.go      # Secret, SecretsManager, FileSystem interfaces
│   ├── aws_impl.go        # AWS Secrets Manager implementation
│   └── operations.go      # Generic operations on Secret interface
├── secretlogic/           # Pure business logic (no I/O, fully testable)
│   ├── secretid.go        # Secret ID generation
│   ├── filepath.go        # File path logic
│   └── metadata.go        # Metadata structure generation
├── jsondoc/               # JSON document secret type
├── textfile/              # Text file secret type
├── sslcert/               # SSL certificate secret type
├── rdspostgres/           # RDS PostgreSQL secret type
├── snowflake/             # Snowflake secret type
├── pull/                  # Pull (download) logic
├── push/                  # Push (upload) logic
├── generate/              # Generation logic
├── uploader/              # CSV upload logic
└── tools/                 # Shared utilities
    ├── constants.go       # Named constants
    ├── output.go          # Consistent output formatting
    ├── file.go            # File operations
    ├── helper.go          # AWS helpers
    └── config.go          # Configuration
```

## Layer Responsibilities

### Layer 1: Commands (cmd/)

**Responsibility**: CLI parsing, user interaction, orchestration

**What they do:**
- Parse command-line flags
- Handle interactive prompts
- Validate user input
- Delegate to business logic layers
- Format output for users

**What they DON'T do:**
- Business logic
- Direct AWS calls
- File I/O (except via tools)

**Example flow (sh-push):**
```go
main()
  → GetConfig() (parse flags)
  → push.PushSecret() (delegate to business logic)
  → Format and display results
```

### Layer 2: Business Logic (pull/, push/, generate/, uploader/)

**Responsibility**: Workflow orchestration, type-specific logic

**What they do:**
- Coordinate multi-step workflows
- Route to type-specific handlers
- Handle secret type differences
- Orchestrate I/O operations

**What they DON'T do:**
- Parse CLI flags (that's cmd layer)
- Direct AWS SDK calls (use secrets/ interfaces)

**Example flow (push):**
```go
PushSecret()
  → Read metadata file
  → Determine secret type
  → Route to pushJSONDoc/pushTextFile/etc
  → Check existence
  → Create or update
```

### Layer 3: Core Interfaces (secrets/)

**Responsibility**: Abstract AWS operations, define contracts

**Key interfaces:**

```go
type Secret interface {
    SecretID() string
    Metadata() map[string]string
    Data() interface{}
    Exists(log *zerolog.Logger) bool
    Create(log *zerolog.Logger) error
    Update(overwrite bool, log *zerolog.Logger) error
}

type SecretsManager interface {
    DescribeSecret(ctx context.Context, secretID string) (bool, error)
    CreateSecret(ctx context.Context, secretID string, value interface{}, tags map[string]string) error
    UpdateSecret(ctx context.Context, secretID string, value interface{}) error
    GetSecretValue(ctx context.Context, secretID string) (string, error)
}
```

**Benefits:**
- Enables testing with mocks
- Abstracts AWS SDK complexity
- Single source of truth for operations
- Consistent behavior across types

**Generic operations:**
```go
GenericExists(s Secret, log, sm SecretsManager) bool
GenericCreate(s Secret, log, sm SecretsManager) error
GenericUpdate(s Secret, log, sm SecretsManager, overwrite bool) error
```

### Layer 4: Secret Types (jsondoc/, textfile/, etc.)

**Responsibility**: Type-specific data structures and AWS operations

**What they implement:**
```go
type Secret struct {
    Metadata Metadata
    Data     Data
}

// Implement secrets.Secret interface
func (s Secret) SecretID() string
func (s Secret) Metadata() map[string]string
func (s Secret) Data() interface{}
func (s Secret) Exists(log *zerolog.Logger) bool
func (s Secret) Create(log *zerolog.Logger) error
func (s Secret) Update(overwrite bool, log *zerolog.Logger) error
```

**Current state:**
Each type package has its own `Exists()`, `Create()`, `Update()` implementations.

**Future optimization:**
These could delegate to `secrets.GenericExists()`, `secrets.GenericCreate()`, etc., reducing duplication. However, the current implementation works and all secret types follow the same pattern.

### Layer 5: Pure Business Logic (secretlogic/)

**Responsibility**: Pure functions with no I/O - fully unit testable

**Characteristics:**
- No AWS calls
- No filesystem operations
- No external dependencies
- Deterministic (same input → same output)
- Fast tests (milliseconds)

**Examples:**
```go
// Pure function - easy to test
BuildSecretID(secretType string, metadata map[string]string) (string, error)

// Pure function - no I/O
BuildBaseFilename(secretType string, metadata map[string]string) (string, error)

// Pure function - generates structure
BuildMetadataMap(secretType string, metadata map[string]string) (map[string]interface{}, error)
```

**Test coverage:**
- 55 unit tests
- 86.9% coverage
- ~3ms execution
- No AWS credentials needed

### Layer 6: Utilities (tools/)

**Responsibility**: Shared utilities and helpers

**Key modules:**

**constants.go** - Named constants
```go
const (
    DefaultFilePermissions = 0644
    DefaultDirPermissions  = 0755
    PlaceholderValue      = "REPLACE-ME"
)
```

**output.go** - Consistent formatting
```go
out := tools.DefaultOutput()
out.Success("Secret created successfully")
out.Progress("Fetching from AWS...")
out.FilesList("Created files", files)
```

**file.go** - File operations
```go
GetWorkingDir() (string, error)
WriteStringToFile(content, filename string) error
FileExists(path string) bool
GetSHA256Sum(filePath string) (string, error)
```

## Design Patterns

### Interface-Based Design

All AWS operations go through the `SecretsManager` interface:

**Benefits:**
- Testable without AWS
- Mockable for unit tests
- Single implementation point
- Consistent error handling

**Usage:**
```go
// Production
sm := secrets.DefaultSecretsManager()
exists := secrets.GenericExists(secret, log, sm)

// Testing (future)
mockSM := &MockSecretsManager{...}
exists := secrets.GenericExists(secret, log, mockSM)
```

### Consistent Output Formatting

All user-facing output uses `tools.Output`:

```go
out := tools.DefaultOutput()
out.Progress("Checking if secret exists: jsondoc/dev/app")
out.Success("Secret created successfully")
out.Error("Failed to connect to AWS")
```

**Benefits:**
- Consistent look and feel
- Easy to change formatting globally
- Testable (inject mock writer)
- Professional appearance

### Pure Functions for Business Logic

Core business logic extracted into pure functions:

**Before:**
```go
func pullJSONDoc(...) error {
    // AWS call
    // Parse JSON
    // Build paths
    // Write files
    // All mixed together - can't test without AWS
}
```

**After:**
```go
// Pure function - fully testable
func BuildSecretID(type, metadata) (string, error) {
    // No I/O - just logic
}

// Uses pure function
func pullJSONDoc(...) error {
    secretID := secretlogic.BuildSecretID(...)
    // Now fetch from AWS
}
```

### Type-Specific Handlers with Common Interface

All secret types implement the same `Secret` interface, allowing generic operations while preserving type-specific behavior:

```go
// Works with ANY secret type
func ProcessSecret(s secrets.Secret) {
    if s.Exists(log) {
        s.Update(true, log)
    } else {
        s.Create(log)
    }
}
```

## Data Flow Examples

### Pull (Download) Flow

```
User: sh-pull -type=jsondoc -env=dev -access=app
  ↓
cmd/sh-pull/main.go
  → GetConfig() (parse flags)
  → pull.PullSecret(metadata)
    ↓
pull/pull.go
  → pullJSONDoc(metadata)
    → secretlogic.BuildSecretID() (pure function)
    → tools.GetSecretValue() (AWS call)
    → Parse JSON
    → tools.GetWorkingDir()
    → tools.WriteStringToFile() × 2
    → Display success with tools.Output
```

### Push (Upload) Flow

```
User: sh-push -metadata=jsondoc.dev.app.metadata.json
  ↓
cmd/sh-push/main.go
  → GetConfig() (parse flags, resolve path)
  → push.PushSecret(metadataFile)
    ↓
push/push.go
  → pushJSONDoc(metadataFile)
    → Read local files
    → Build jsondoc.Secret struct
    → secret.Exists() (AWS check via secrets/ layer)
    → If not exists:
        → Display contents
        → Confirm with user
        → secret.Create() (AWS call via secrets/ layer)
    → If exists:
        → Fetch remote
        → Generate diff
        → Confirm with user  
        → secret.Update() (AWS call via secrets/ layer)
```

### Generate (Scaffolding) Flow

```
User: sh-generate
  ↓
cmd/sh-generate/main.go
  → GetConfig() (just debug flag)
  → generate.GenerateSecretFiles()
    ↓
generate/generate.go
  → Interactive prompts for type + metadata
  → Route to generateJSONDoc/generateTextFile/etc
    → secretlogic.BuildMetadataMap() (pure function)
    → tools.GetWorkingDir()
    → tools.WriteStringToFile() × N
    → Display next steps with tools.Output
```

## Secret Types

### Common Pattern

All secret types follow the same structure:

```go
type Metadata struct {
    ResourceType string
    Environment  string
    // Type-specific fields
}

type Data struct {
    // Type-specific data fields
}

type Secret struct {
    Metadata Metadata
    Data     Data
}

// Implement secrets.Secret interface
func (s Secret) SecretID() string { return s.Metadata.SecretID() }
func (s Secret) Metadata() map[string]string { return s.Metadata.Map() }
func (s Secret) Data() interface{} { return s.Data }
func (s Secret) Exists(log) bool { /* AWS check */ }
func (s Secret) Create(log) error { /* AWS create */ }
func (s Secret) Update(overwrite, log) error { /* AWS update */ }
```

### Type-Specific Differences

| Type | Metadata Fields | Data Fields | File Count |
|------|----------------|-------------|------------|
| jsondoc | env, access | JSONContents, SHA256 | 2 (metadata + contents) |
| text_file | env, access | Contents, SHA256 | 2 (metadata + contents) |
| ssl_certificate | env, commonname | Cert, Key, Expiration, etc | 3 (metadata + crt + key) |
| rdspostgres | env, inst, db, access | Password, Host, Port, etc | 1 (combined JSON) |
| snowflake | env, warehouse, access | Password, Account, etc | 1 (combined JSON) |

## Testing Strategy

### Layer-by-Layer Testing

**secretlogic/ (Pure Business Logic)**
- Unit tests with table-driven tests
- Golden file tests for complex outputs
- No AWS, no filesystem needed
- Coverage: 86.9%
- Execution: ~3ms

**secrets/ (Interfaces & Operations)**
- Currently: tested indirectly through integration
- Future: mock-based unit tests
- Test generic operations with mock SecretsManager

**cmd/ (Commands)**
- Manual testing (see MANUAL-TEST.md)
- Integration tests with real AWS (optional)

**Type packages (jsondoc/, etc.)**
- CSV parsing tested
- Exists/Create/Update tested with real AWS (integration tests)
- Future: use mock SecretsManager

### Test Types

**Unit Tests** (Current: secretlogic/)
- Fast (<1ms per test)
- No external dependencies
- Run on every change
- High coverage target (>80%)

**Integration Tests** (Current: some in get/)
- Require AWS credentials
- Test real AWS operations
- Slower (seconds)
- Run before release

**Manual Tests** (MANUAL-TEST.md)
- End-to-end workflows
- All 5 secret types
- Human verification
- Run before major releases

## Configuration

### Working Directory

All file operations use `~/.secret-hoard/` as the working directory:

```go
workingDir, err := tools.GetWorkingDir()
// Returns: /home/user/.secret-hoard/
// Creates directory if it doesn't exist
```

### File Naming Conventions

**Pattern:** `{type}.{env}.{metadata-fields}.{extension}`

**Examples:**
```
jsondoc.dev.app.metadata.json
jsondoc.dev.app.contents.json
rdspostgres.prod.db1.appdb.readonly.json
sslcert.staging.example.com.metadata.json
sslcert.staging.example.com.crt
sslcert.staging.example.com.key
```

### AWS Credentials

Uses standard AWS SDK credential chain:
1. Environment variables (AWS_ACCESS_KEY_ID, etc.)
2. Shared credentials file (~/.aws/credentials)
3. IAM role (if running on EC2/ECS/Lambda)

## Error Handling

### Graceful Degradation

```go
// Commands exit with proper codes
if err != nil {
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    os.Exit(1)
}
```

### Helpful Error Messages

All errors provide context and examples:

```
Error: metadata file is required

Usage:
  sh-push -metadata=<file>

Examples:
  sh-push -metadata=jsondoc.dev.app.metadata.json

Tip: Files are typically in ~/.secret-hoard/
```

### Error Context

Errors wrap with context as they bubble up:

```go
if err != nil {
    return fmt.Errorf("error fetching secret %s: %w", secretID, err)
}
```

## Future Improvements

### Phase 2 (Optional): Full Interface Migration

Currently, secret type packages still call AWS directly. They could be refactored to use `secrets.GenericExists()`, `secrets.GenericCreate()`, etc.

**Benefits:**
- Further code reduction
- Easier to test with mocks
- Single source of truth

**Trade-off:**
- More refactoring work
- Current implementation works fine

### Phase 3 (Optional): Generic Pull/Push

Currently, we have 5 pull functions and 5 push functions. These could be consolidated:

```go
// One pull function for all types
func GenericPull(secretType string, metadata map[string]string) error

// One push function for all types
func GenericPush(metadataFile string) error
```

**Benefits:**
- Significant code reduction
- Bug fixes in one place

**Trade-off:**
- Loss of type-specific customization
- More complex generic code

## Conclusion

The architecture balances:
- **Simplicity**: Clear layers, obvious flow
- **Testability**: Pure functions, interfaces for mocking
- **Maintainability**: Consistent patterns, low duplication
- **User Experience**: Helpful errors, progress indicators, professional output

The foundation is in place for further improvements while remaining pragmatic about what's "good enough" for production use.
