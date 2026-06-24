# Code Cleanup Recommendations

**Status: 50% Complete** | Last Updated: 2026-06-24

This document provides prioritized recommendations to make the codebase simpler to test, more readable, and more usable.

## 📈 Progress Summary

✅ **Completed**: Phases 0, 2 (partial), 4 (partial), 5 + Logging Simplification
- Removed 6 obsolete commands
- Added interfaces & generic operations (foundation ready)
- Replaced zerolog with simple logging
- Fixed all panics, added progress indicators
- Created standardized output formatting

❌ **Remaining**: Phases 1, 2 (I/O separation), 3 (use generic ops), refactoring
- CSV consolidation (~500 lines to remove)
- Migrate to generic operations (would cut type packages by 60%)
- Business logic separation (enable unit testing)

**Next Priority**: Phase 1 - CSV Consolidation

## IMPORTANT: Decisions Made

**ACCEPTED RECOMMENDATIONS**: All recommendations except 3.3
- **3.3 (Random Confirmation)**: KEEP random strings - prevents confirm fatigue accidents

**REMOVE OLD COMMANDS**: sh-download and 5 type-specific upload commands
- Old type-specific commands (sh-jsondoc, sh-rdsinstance, sh-snowflake, sh-sslcert, sh-textfile) are already superseded by unified sh-upload
- sh-download is superseded by interactive sh-pull
- Removing 6 obsolete commands enables additional cleanup opportunities

**FINAL COMMAND SET** (4 commands):
- `sh-upload` - Batch CSV upload (all types)
- `sh-pull` - Interactive/flag-based download
- `sh-push` - Interactive/flag-based upload with diff
- `sh-generate` - Scaffolding generator

---

## Priority 0: Remove Obsolete Commands (NEW - HIGHEST PRIORITY)

### 0.1 Delete Type-Specific Upload Commands (HIGH IMPACT)

**Problem**: 5 type-specific CSV upload commands (sh-jsondoc, sh-rdsinstance, sh-snowflake, sh-sslcert, sh-textfile) are superseded by the unified `sh-upload` command.

**Current State**:
- `cmd/sh-jsondoc/` - CSV upload for jsondoc only
- `cmd/sh-rdsinstance/` - CSV upload for rdspostgres only
- `cmd/sh-snowflake/` - CSV upload for snowflake only
- `cmd/sh-sslcert/` - CSV upload for ssl_certificate only
- `cmd/sh-textfile/` - CSV upload for text_file only

All follow the same pattern:
1. Read CSV file
2. Parse records
3. Call Create() or Update()

**Why remove**:
- `sh-upload` already handles all 5 types in one command
- Duplication of CSV parsing logic
- Confusing to have both type-specific and unified commands
- More executables to build, test, and document

**Recommendation**:

Delete the following:
```bash
rm -rf cmd/sh-jsondoc/
rm -rf cmd/sh-rdsinstance/
rm -rf cmd/sh-snowflake/
rm -rf cmd/sh-sslcert/
rm -rf cmd/sh-textfile/
```

Update Makefile:
```makefile
# Before:
EXECUTABLES := sh-download sh-upload sh-pull sh-push sh-generate

# After:
EXECUTABLES := sh-upload sh-pull sh-push sh-generate
```

Update documentation:
- Remove sections about type-specific commands from README.md
- Emphasize sh-upload handles all types

**Benefits**:
- 5 fewer commands to maintain
- Simpler user experience
- Clearer that sh-upload is the batch tool
- Less build time
- Smaller documentation

---

### 0.2 Delete sh-download Command (HIGH IMPACT)

**Problem**: `sh-download` is superseded by the more capable `sh-pull` command.

**Current State**:
- `cmd/sh-download/` - Downloads by secret ID to specified file path
- Doesn't support interactive mode
- Doesn't use working directory
- Requires exact file path specification
- Less user-friendly than sh-pull

**Comparison**:

| Feature | sh-download | sh-pull |
|---------|-------------|---------|
| Download secret | ✓ | ✓ |
| Interactive mode | ✗ | ✓ |
| Flag mode | ✓ | ✓ |
| Working directory | ✗ | ✓ |
| Predictable filenames | ✗ | ✓ |
| Metadata separation | ✗ | ✓ |
| Ready for sh-push | ✗ | ✓ |

**Recommendation**:

Delete sh-download:
```bash
rm -rf cmd/sh-download/
rm -rf get/  # If only used by sh-download
```

Update Makefile:
```makefile
EXECUTABLES := sh-upload sh-pull sh-push sh-generate
```

Update documentation to show sh-pull as the download tool:
```bash
# Old way:
sh-download -id=jsondoc/dev/app -file=/tmp/output.json

# New way (interactive):
sh-pull
# Select type: jsondoc
# Environment: dev
# Access: app

# New way (flags):
sh-pull -type=jsondoc -env=dev -access=app
```

**Benefits**:
- One download tool instead of two
- Better UX with interactive mode
- Files go to predictable location
- Integrates with sh-push workflow

---

### 0.3 Additional Cleanup Enabled by Removals (MEDIUM IMPACT)

**Problem**: Removing old commands enables further simplification.

**Opportunities**:

1. **Simplify get/ package**:
   - `get/` package appears to only be used by sh-download
   - If so, it can be deleted entirely
   - Check if any code references it:
   ```bash
   grep -r "github.com/natemarks/secret-hoard/get" --include="*.go" | grep -v "^get/"
   ```
   - If only sh-download uses it, delete the entire package

2. **Remove CSV parsing from type packages**:
   - Each type package (jsondoc, rdspostgres, etc.) has CSV parsing code
   - This was needed for type-specific commands
   - Now only sh-upload needs CSV parsing
   - Can consolidate all CSV logic into upload/ package

3. **Simplify testing**:
   - Remove tests for deleted commands
   - Remove CSV parsing tests from type packages
   - Focus tests on unified sh-upload

4. **Streamline examples/**:
   - Review examples/ directory
   - Keep only sh-upload CSV examples
   - Remove any type-specific examples

**Recommendation**:

After deleting commands, audit for dead code:
```bash
# Find unused imports
goimports -l .

# Find dead code
deadcode ./...

# Check for orphaned test files
find . -name "*_test.go" -type f | while read f; do
  pkg=$(dirname "$f")
  if [ ! -d "$pkg" ]; then
    echo "Orphaned test: $f"
  fi
done
```

Delete identified dead code.

**Benefits**:
- Even more code reduction (potentially 500+ lines)
- Clearer package responsibilities
- Faster builds
- Less to test

---

### 0.4 Update Documentation for New Command Set (HIGH IMPACT)

**Problem**: Documentation references old commands.

**Recommendation**:

Update README.md to reflect final command set:

```markdown
# secret-hoard

AWS Secrets Manager management tool with batch and interactive workflows.

## Commands

### sh-upload - Batch Upload from CSV
Upload multiple secrets of any type from a single CSV file.

```bash
sh-upload -file=secrets.csv -overwrite
```

CSV format (type determined by ResourceType column):
```csv
ResourceType,Environment,Instance,Database,Access,Password,Engine,Port,DbInstanceIdentifier,Host,Username
rdspostgres,dev,mydb,appdb,readonly,pass123,postgres,5432,mydb-dev,mydb.aws.com,readonly_user
```

Supports all 5 secret types: rdspostgres, snowflake, ssl_certificate, jsondoc, text_file

### sh-generate - Create Scaffolding
Interactively generate local file templates for new secrets.

```bash
sh-generate
```

Creates files in `~/.secret-hoard/` with REPLACE-ME placeholders.

### sh-pull - Download Secrets
Download a secret to local editable files.

Interactive mode:
```bash
sh-pull
```

Flag mode (scriptable):
```bash
sh-pull -type=jsondoc -env=dev -access=app
```

Files created in `~/.secret-hoard/` with predictable names.

### sh-push - Upload with Diff
Upload local files to AWS with diff review and confirmation.

```bash
sh-push -metadata=jsondoc.dev.app.metadata.json
```

Automatically creates secret if it doesn't exist, or updates with diff display if it exists.

## Workflows

### New Secret (Interactive)
```bash
sh-generate          # Create scaffolding
# Edit files in ~/.secret-hoard/
sh-push -metadata=<file>  # Upload (auto-creates)
```

### Update Existing Secret
```bash
sh-pull -type=jsondoc -env=dev -access=app  # Download
# Edit files in ~/.secret-hoard/
sh-push -metadata=jsondoc.dev.app.metadata.json  # Upload with diff
```

### Batch Operations
```bash
# Create CSV with many secrets
sh-upload -file=all_secrets.csv -overwrite
```

## Migration from Old Commands

Old commands → New commands:
- `sh-jsondoc -file=...` → `sh-upload -file=...`
- `sh-rdsinstance -file=...` → `sh-upload -file=...`
- `sh-snowflake -file=...` → `sh-upload -file=...`
- `sh-sslcert -file=...` → `sh-upload -file=...`
- `sh-textfile -file=...` → `sh-upload -file=...`
- `sh-download -id=... -file=...` → `sh-pull -type=... -env=...`
```

**Benefits**:
- Clear command purpose
- No confusion about which tool to use
- Modern workflow documentation
- Migration guide for existing users

---

## Priority 1: Critical Testing & Maintainability Issues

### 1.0 Removal Impact Analysis (NEW - ENABLED BY DELETIONS)

**After removing 6 obsolete commands, additional cleanup becomes possible:**

#### 1.0.1 Consolidate CSV Parsing

**Current State**:
Each secret type package has CSV parsing code:
- `jsondoc/jsondoccsv.go` - CSV parsing for jsondoc
- `rdspostgres/rdspostgrescsv.go` - CSV parsing for rdspostgres
- `snowflake/snowflakecsv.go` - CSV parsing for snowflake
- `sslcert/sslcertcsv.go` - CSV parsing for ssl_certificate
- `textfile/textfilecsv.go` - CSV parsing for text_file

**Why this existed**:
- Type-specific commands (sh-jsondoc, etc.) needed to parse their CSV format
- Each command only handled one type
- Duplication was across commands, not within a command

**Now that type-specific commands are gone**:
- Only `sh-upload` needs CSV parsing
- sh-upload already has generic CSV handling in upload/ package
- Type-specific CSV code is unused

**Recommendation**:
Move all CSV parsing to upload/ package:
```go
// upload/csv.go - Centralized CSV parsing

func ParseCSVRecord(record CSVRecord, secretType string) (Secret, error) {
    switch secretType {
    case "jsondoc":
        return parseJSONDocRecord(record)
    case "rdspostgres":
        return parseRDSPostgresRecord(record)
    case "snowflake":
        return parseSnowflakeRecord(record)
    case "ssl_certificate":
        return parseSSLCertRecord(record)
    case "text_file":
        return parseTextFileRecord(record)
    default:
        return nil, fmt.Errorf("unknown type: %s", secretType)
    }
}
```

Delete from type packages:
```bash
rm jsondoc/jsondoccsv.go
rm rdspostgres/rdspostgrescsv.go
rm snowflake/snowflakecsv.go
rm sslcert/sslcertcsv.go
rm textfile/textfilecsv.go
```

**Benefits**:
- ~500 lines of duplicate CSV code removed
- Single place for CSV logic
- Type packages focus on secret operations only
- Easier to maintain CSV format

#### 1.0.2 Remove get/ Package

**Current State**:
- `get/` package contains download logic
- Used exclusively by sh-download
- Has type-specific download functions
- ~400 lines of code

**Check if anything else uses it**:
```bash
grep -r "github.com/natemarks/secret-hoard/get" --include="*.go" | grep -v "^get/" | grep -v "cmd/sh-download"
```

**If only sh-download uses it**:
```bash
rm -rf get/
```

**If sh-pull or tests use it**:
- Evaluate if code can be moved to pull/ package
- Consolidate with pull/ package functions
- Remove duplication

**Benefits**:
- ~400 lines removed if unused
- One less package to maintain
- Clearer that pull/ is the download package

#### 1.0.3 Simplify Type Package Responsibilities

**After removing CSV parsing, type packages become much simpler:**

Current responsibilities:
- Define Metadata and Data structs ✓ (keep)
- Implement Exists(), Create(), Update() ✓ (keep)
- Implement SecretID() ✓ (keep)
- CSV parsing ✗ (move to upload/)
- CSV record types ✗ (move to upload/)

**New streamlined type packages** (~80 lines each):
```go
// jsondoc/jsondoc.go - Only secret operations

package jsondoc

type Metadata struct { ... }
type Data struct { ... }
type Secret struct { Metadata, Data }

func (m Metadata) SecretID() string { ... }
func (m Metadata) Map() map[string]string { ... }
func (s Secret) Exists(log *zerolog.Logger) bool { ... }
func (s Secret) Create(log *zerolog.Logger) error { ... }
func (s Secret) Update(overwrite bool, log *zerolog.Logger) error { ... }
```

**Benefits**:
- Type packages: ~200 lines → ~80 lines (60% reduction)
- Clear single responsibility (secret operations)
- No CSV knowledge in type packages
- Easier to implement generic Secret interface (1.1)

---

### 1.1 Massive Code Duplication (HIGH IMPACT)

**Problem**: Each secret type package (jsondoc, textfile, sslcert, rdspostgres, snowflake) has 95% identical code (~200 lines each).

**Current State**:
- 5 packages with nearly identical `Exists()`, `Create()`, `Update()` methods
- Pull/push functions repeated 5 times with only minor differences
- Bug fixes must be applied 5 times
- New secret types require copying ~200 lines of code

**Location**:
- `jsondoc/jsondoc.go` (197 lines)
- `textfile/textfile.go` (198 lines)
- `rdspostgres/rdspostgres.go` (205 lines)
- `sslcert/sslcert.go` (246 lines)
- `snowflake/snowflake.go` (similar)
- `pull/pull.go` - 5 nearly identical functions
- `push/push.go` - 5 nearly identical functions

**Recommendation**:

Create a generic Secret interface and shared operations:

```go
// secrets/secret.go - New generic package

type Secret interface {
    SecretID() string
    Metadata() map[string]string
    Data() interface{}
    Exists(log *zerolog.Logger, sm SecretsManager) bool
    Create(log *zerolog.Logger, sm SecretsManager) error
    Update(log *zerolog.Logger, sm SecretsManager, overwrite bool) error
}

type SecretsManager interface {
    DescribeSecret(ctx context.Context, secretID string) (bool, error)
    CreateSecret(ctx context.Context, secretID string, value interface{}, tags map[string]string) error
    UpdateSecret(ctx context.Context, secretID string, value interface{}) error
    GetSecretValue(ctx context.Context, secretID string) (string, error)
}

// Generic operations work on any Secret
func GenericExists(s Secret, log *zerolog.Logger, sm SecretsManager) bool {
    exists, err := sm.DescribeSecret(context.Background(), s.SecretID())
    if err != nil {
        log.Error().Err(err).Msg("error checking if secret exists")
        return false
    }
    return exists
}

func GenericCreate(s Secret, log *zerolog.Logger, sm SecretsManager) error {
    tags := s.Metadata()
    return sm.CreateSecret(context.Background(), s.SecretID(), s.Data(), tags)
}

func GenericUpdate(s Secret, log *zerolog.Logger, sm SecretsManager, overwrite bool) error {
    return sm.UpdateSecret(context.Background(), s.SecretID(), s.Data())
}
```

**Type-specific code becomes minimal**:
```go
// jsondoc/jsondoc.go - Now only ~50 lines

type Secret struct {
    Metadata Metadata
    Data     Data
}

func (s Secret) SecretID() string {
    return s.Metadata.SecretID()
}

func (s Secret) Metadata() map[string]string {
    return s.Metadata.Map()
}

func (s Secret) Data() interface{} {
    return s.Data
}

// Use generic implementations
func (s Secret) Exists(log *zerolog.Logger) bool {
    return secrets.GenericExists(s, log, secrets.DefaultSecretsManager())
}

func (s Secret) Create(log *zerolog.Logger) error {
    return secrets.GenericCreate(s, log, secrets.DefaultSecretsManager())
}

func (s Secret) Update(overwrite bool, log *zerolog.Logger) error {
    return secrets.GenericUpdate(s, log, secrets.DefaultSecretsManager(), overwrite)
}
```

**Benefits**:
- 5 packages reduced from ~200 lines to ~50 lines each
- Bug fixes in one place
- New secret types only need to implement interface
- 80% less code to maintain

---

### 1.2 Business Logic Tightly Coupled with I/O (HIGH IMPACT)

**Problem**: Pull and push functions mix business logic with AWS calls, filesystem operations, and user interaction - impossible to unit test.

**Current State**:
```go
// pull/pull.go - Lines 45-108
func pullJSONDoc(cfg Config, log *zerolog.Logger) error {
    // 1. Build secret ID (business logic)
    secretID := fmt.Sprintf("jsondoc/%s/%s", cfg.Env, cfg.Access)
    
    // 2. Fetch from AWS (I/O)
    secretValue, err := tools.GetSecretValue(secretID)
    
    // 3. Parse JSON (business logic)
    var data jsondoc.Data
    json.Unmarshal([]byte(secretValue), &data)
    
    // 4. Get working directory (filesystem I/O)
    workingDir, _ := tools.GetWorkingDir()
    
    // 5. Build file paths (business logic)
    metadataFile := path.Join(workingDir, fmt.Sprintf("jsondoc.%s.%s.metadata.json", cfg.Env, cfg.Access))
    
    // 6. Write files (filesystem I/O)
    tools.WriteStringToFile(metadataFile, metadataJSON)
    
    // 7. Print to user (UI)
    fmt.Printf("Successfully pulled secret: %s\n", secretID)
}
```

**Recommendation**:

Separate into layers:

```go
// secretlogic/pull.go - Pure business logic (testable)

type PullData struct {
    SecretID      string
    MetadataFile  string
    ContentsFile  string
    MetadataJSON  string
    ContentsJSON  string
}

// Pure function - easily unit tested
func BuildPullFiles(secretType, env, access, workingDir string, secretValue []byte) (PullData, error) {
    var result PullData
    result.SecretID = fmt.Sprintf("%s/%s/%s", secretType, env, access)
    
    // Parse secret value
    var data jsondoc.Data
    if err := json.Unmarshal(secretValue, &data); err != nil {
        return result, fmt.Errorf("error parsing secret: %w", err)
    }
    
    // Build file paths
    base := fmt.Sprintf("%s.%s.%s", secretType, env, access)
    result.MetadataFile = path.Join(workingDir, base+".metadata.json")
    result.ContentsFile = path.Join(workingDir, base+".contents.json")
    
    // Build file contents
    result.MetadataJSON = buildMetadataJSON(...)
    result.ContentsJSON = data.JSONContents
    
    return result, nil
}

// pull/pull.go - I/O operations only

func pullJSONDoc(cfg Config, log *zerolog.Logger, fetcher SecretFetcher, writer FileWriter) error {
    // Get working directory
    workingDir, err := writer.GetWorkingDir()
    if err != nil {
        return err
    }
    
    // Fetch from AWS
    secretValue, err := fetcher.GetSecret(cfg.BuildSecretID())
    if err != nil {
        return err
    }
    
    // Call pure business logic
    pullData, err := secretlogic.BuildPullFiles(cfg.Type, cfg.Env, cfg.Access, workingDir, secretValue)
    if err != nil {
        return err
    }
    
    // Write files
    if err := writer.WriteFile(pullData.MetadataFile, pullData.MetadataJSON); err != nil {
        return err
    }
    if err := writer.WriteFile(pullData.ContentsFile, pullData.ContentsJSON); err != nil {
        return err
    }
    
    // User feedback
    fmt.Printf("Successfully pulled secret: %s\n", pullData.SecretID)
    return nil
}
```

**Testing becomes simple**:
```go
func TestBuildPullFiles(t *testing.T) {
    // No AWS, no filesystem needed!
    secretValue := []byte(`{"JSONContents": "{}", "JSONSha256Sum": "abc"}`)
    
    result, err := secretlogic.BuildPullFiles("jsondoc", "dev", "app", "/tmp", secretValue)
    
    assert.NoError(t, err)
    assert.Equal(t, "jsondoc/dev/app", result.SecretID)
    assert.Equal(t, "/tmp/jsondoc.dev.app.metadata.json", result.MetadataFile)
    // ... test all business logic
}
```

**Benefits**:
- Business logic testable without AWS credentials
- Fast unit tests (milliseconds vs seconds)
- Easy to mock I/O for integration tests
- Clear separation of concerns

---

### 1.3 Missing Interfaces for Testability (MEDIUM IMPACT)

**Problem**: Direct AWS SDK usage throughout codebase prevents mocking.

**Current State**:
```go
// Hardcoded AWS SDK usage
func (s Secret) Exists(log *zerolog.Logger) bool {
    cfg, _ := config.LoadDefaultConfig(context.Background())
    client := secretsmanager.NewFromConfig(cfg)
    _, err := client.DescribeSecret(context.Background(), &secretsmanager.DescribeSecretInput{
        SecretId: aws.String(s.Metadata.SecretID()),
    })
    return err == nil
}
```

**Recommendation**:

Create interfaces and dependency injection:

```go
// secrets/interfaces.go

type SecretsManager interface {
    DescribeSecret(ctx context.Context, secretID string) (bool, error)
    CreateSecret(ctx context.Context, params CreateParams) error
    UpdateSecret(ctx context.Context, params UpdateParams) error
    GetSecretValue(ctx context.Context, secretID string) (string, error)
}

type FileSystem interface {
    GetWorkingDir() (string, error)
    ReadFile(path string) ([]byte, error)
    WriteFile(path string, content []byte) error
    FileExists(path string) bool
}

// secrets/aws_impl.go - Real implementation

type AWSSecretsManager struct {
    client *secretsmanager.Client
}

func (sm *AWSSecretsManager) DescribeSecret(ctx context.Context, secretID string) (bool, error) {
    _, err := sm.client.DescribeSecret(ctx, &secretsmanager.DescribeSecretInput{
        SecretId: aws.String(secretID),
    })
    if err != nil {
        var notFound *types.ResourceNotFoundException
        if errors.As(err, &notFound) {
            return false, nil
        }
        return false, err
    }
    return true, nil
}

// secrets/mock_impl.go - For testing

type MockSecretsManager struct {
    Secrets map[string]interface{}
}

func (sm *MockSecretsManager) DescribeSecret(ctx context.Context, secretID string) (bool, error) {
    _, exists := sm.Secrets[secretID]
    return exists, nil
}

// Usage in production
sm := secrets.NewAWSSecretsManager()
secret.Exists(log, sm)

// Usage in tests
sm := &MockSecretsManager{Secrets: map[string]interface{}{
    "jsondoc/dev/app": mockData,
}}
secret.Exists(log, sm)  // No AWS needed!
```

**Benefits**:
- Tests run without AWS credentials
- Fast test execution
- Predictable test results
- Easy to test error conditions

---

## Priority 2: Readability & Maintainability

### 2.1 Long Functions Doing Too Many Things (MEDIUM IMPACT)

**Problem**: Functions exceed 100 lines with multiple responsibilities.

**Examples**:
- `pull/pull.go:pullSSLCert()` - 81 lines
- `push/push.go:pushJSONDoc()` - 102 lines
- `push/push.go:pushSSLCert()` - 138 lines

**Current State**:
```go
// push/push.go - pushSSLCert is 138 lines doing:
// 1. Extract metadata
// 2. Build file paths
// 3. Read certificate file
// 4. Read key file
// 5. Extract expiration date
// 6. Extract certificate modulus
// 7. Extract key modulus
// 8. Compare moduli
// 9. Compute certificate SHA256
// 10. Compute key SHA256
// 11. Build secret struct
// 12. Check if exists
// 13. Create or update with prompting
```

**Recommendation**:

Break into small, single-purpose functions:

```go
// push/ssl.go - Separated concerns

type SSLCertFiles struct {
    CertFile string
    KeyFile  string
}

type SSLCertData struct {
    Certificate       string
    PrivateKey        string
    ExpirationDate    string
    Modulus           string
    CertificateSha256 string
    PrivateKeySha256  string
}

// Step 1: Parse metadata and find files
func buildSSLCertFilePaths(metadataFile string) (SSLCertFiles, error) {
    baseName := strings.TrimSuffix(metadataFile, ".metadata.json")
    return SSLCertFiles{
        CertFile: baseName + ".crt",
        KeyFile:  baseName + ".key",
    }, nil
}

// Step 2: Read and validate certificate files
func readSSLCertFiles(files SSLCertFiles) (cert, key string, err error) {
    cert, err = tools.ReadFileToString(files.CertFile)
    if err != nil {
        return "", "", fmt.Errorf("error reading certificate: %w", err)
    }
    
    key, err = tools.ReadFileToString(files.KeyFile)
    if err != nil {
        return "", "", fmt.Errorf("error reading key: %w", err)
    }
    
    return cert, key, nil
}

// Step 3: Extract and validate all certificate data
func extractSSLCertData(files SSLCertFiles, cert, key string) (SSLCertData, error) {
    var data SSLCertData
    data.Certificate = cert
    data.PrivateKey = key
    
    // Extract expiration
    expiration, err := extractExpiration(files.CertFile)
    if err != nil {
        return data, fmt.Errorf("error extracting expiration: %w", err)
    }
    data.ExpirationDate = expiration
    
    // Extract and compare moduli
    certModulus, err := extractCertificateModulus(files.CertFile)
    if err != nil {
        return data, fmt.Errorf("error extracting certificate modulus: %w", err)
    }
    
    keyModulus, err := extractPrivateKeyModulus(files.KeyFile)
    if err != nil {
        return data, fmt.Errorf("error extracting key modulus: %w", err)
    }
    
    if certModulus != keyModulus {
        return data, fmt.Errorf("certificate and private key moduli do not match")
    }
    data.Modulus = certModulus
    
    // Compute SHA256 sums
    data.CertificateSha256, _ = tools.GetSHA256Sum(files.CertFile)
    data.PrivateKeySha256, _ = tools.GetSHA256Sum(files.KeyFile)
    
    return data, nil
}

// Step 4: Build secret from data
func buildSSLSecret(data SSLCertData, meta sslcert.Metadata) sslcert.Secret {
    return sslcert.Secret{
        Metadata: meta,
        Data: sslcert.Data{
            Certificate:       data.Certificate,
            PrivateKey:        data.PrivateKey,
            ExpirationDate:    data.ExpirationDate,
            Modulus:           data.Modulus,
            CertificateSha256: data.CertificateSha256,
            PrivateKeySha256:  data.PrivateKeySha256,
        },
    }
}

// Main function now orchestrates small functions
func pushSSLCert(metadataFile string, metadataMap map[string]interface{}, log *zerolog.Logger) error {
    // Extract metadata
    environment := metadataMap["environment"].(string)
    commonName := metadataMap["commonName"].(string)
    
    // Build file paths
    files, err := buildSSLCertFilePaths(metadataFile)
    if err != nil {
        return err
    }
    
    // Read files
    cert, key, err := readSSLCertFiles(files)
    if err != nil {
        return err
    }
    
    // Extract and validate data
    data, err := extractSSLCertData(files, cert, key)
    if err != nil {
        return err
    }
    
    // Build secret
    meta := sslcert.Metadata{
        ResourceType: "ssl_certificate",
        Environment:  environment,
        CommonName:   commonName,
    }
    secret := buildSSLSecret(data, meta)
    
    // Create or update (generic function)
    return pushSecret(secret, log)
}
```

**Benefits**:
- Each function < 20 lines
- Easy to understand what each does
- Easy to test individually
- Easy to reuse components
- Clear error messages at each step

---

### 2.2 Inconsistent Patterns Between Commands (MEDIUM IMPACT)

**Problem**: Three commands use different configuration patterns without clear reason.

**Current State**:
- `sh-pull`: Supports both flags and interactive mode
- `sh-push`: Only flags
- `sh-generate`: Minimal flags, all interactive

**Recommendation**:

Standardize on a hybrid approach for all commands:

```go
// Shared pattern in all cmd/*/config.go

type Config struct {
    // Common fields
    Debug      bool
    Interactive bool  // Add this flag
    
    // Command-specific fields
    ...
}

func GetConfig() (Config, error) {
    // Parse flags
    debugPtr := flag.Bool("debug", false, "Enable debug mode")
    interactivePtr := flag.Bool("interactive", false, "Use interactive mode (ignore other flags)")
    
    // Command-specific flags
    ...
    
    flag.Parse()
    
    config := Config{Debug: *debugPtr, Interactive: *interactivePtr}
    
    // If interactive flag set OR required flags missing, prompt
    if config.Interactive || !hasRequiredFlags() {
        return promptForConfig()
    }
    
    // Otherwise use flags
    return configFromFlags(), nil
}
```

**Benefits**:
- Consistent user experience
- Users can choose their workflow
- Clear when to use flags vs interactive
- Scripts can use flags, humans can use interactive

---

### 2.3 Copy-Paste Errors in Comments (LOW IMPACT, EASY FIX)

**Problem**: Comments copied between secret types contain wrong type names.

**Examples**:
- `jsondoc/jsondoc.go:26` - Comment says "RDSSecretMetadata"
- `sslcert/sslcert.go:26` - Comment says "RDSSecretData"
- Multiple similar errors

**Recommendation**:

Fix all copy-paste errors:

```bash
# Find and fix
grep -r "RDSSecret" jsondoc/ textfile/ sslcert/ snowflake/
```

**Better**: If implementing generic Secret interface (recommendation 1.1), these comments become unnecessary.

---

### 2.4 Magic Numbers Without Named Constants (LOW IMPACT, EASY FIX)

**Problem**: Hardcoded values throughout code.

**Examples**:
- `tools/file.go:35` - `0644` file permissions
- `tools/file.go:107` - `0755` directory permissions
- `get/jsondoc_test.go:27` - `30 * time.Second` wait time
- `generate/generate.go` - Multiple "REPLACE-ME" strings

**Recommendation**:

Create constants file:

```go
// tools/constants.go

package tools

import "time"

const (
    // File permissions
    DefaultFilePermissions = 0644
    DefaultDirPermissions  = 0755
    
    // AWS
    SecretDeletionWaitTime = 30 * time.Second
    
    // Templates
    PlaceholderValue = "REPLACE-ME"
    
    // Confirmation
    ConfirmationStringLength = 4
)
```

Use throughout code:
```go
err = os.WriteFile(filePath, []byte(content), tools.DefaultFilePermissions)
```

**Benefits**:
- Clear meaning
- Easy to change globally
- Self-documenting

---

## Priority 3: Usability Improvements

### 3.1 Poor Error Messages (HIGH USER IMPACT)

**Problem**: Errors don't help users fix issues.

**Current Examples**:
```
"metadata file is required (-metadata flag)"
"unknown secret type: foo"
"invalid file path: /some/path"
"environment is required (-env flag)"
```

**Recommendation**:

Provide helpful context in every error:

```go
// tools/errors.go - Helper for better errors

func RequiredFlagError(flagName, example string) error {
    return fmt.Errorf(`required flag missing: -%s

Example usage:
  %s

Run with --help for more information`, flagName, example)
}

func UnknownTypeError(provided string, validTypes []string) error {
    return fmt.Errorf(`unknown secret type: %q

Valid types:
%s

Example:
  sh-pull -type=jsondoc -env=dev -access=app`, 
        provided, 
        "  - " + strings.Join(validTypes, "\n  - "))
}

func FileNotFoundError(path string, searchLocations []string) error {
    return fmt.Errorf(`file not found: %s

Searched in:
%s

Tip: Metadata files are typically in %s
Run 'ls %s' to see available files`,
        path,
        "  - " + strings.Join(searchLocations, "\n  - "),
        tools.DefaultWorkingDir,
        tools.DefaultWorkingDir)
}

// Usage
if secretType == "" {
    return tools.RequiredFlagError("type", "sh-pull -type=jsondoc -env=dev -access=app")
}

if !isValidType(secretType) {
    return tools.UnknownTypeError(secretType, []string{"jsondoc", "text_file", "ssl_certificate", "rdspostgres", "snowflake"})
}
```

**Benefits**:
- Users can self-correct
- Reduced support burden
- Better first-time experience
- Shows examples inline

---

### 3.2 No Dry-Run Mode (MEDIUM USER IMPACT)

**Problem**: No way to preview operations without executing them.

**Recommendation**:

Add `--dry-run` flag to all commands:

```go
// config.go for all commands
type Config struct {
    DryRun bool
    // ... other fields
}

func GetConfig() (Config, error) {
    dryRunPtr := flag.Bool("dry-run", false, "Show what would be done without doing it")
    // ...
    config.DryRun = *dryRunPtr
    return config, nil
}

// pull/pull.go
func pullJSONDoc(cfg Config, log *zerolog.Logger) error {
    secretID := buildSecretID(cfg)
    
    if cfg.DryRun {
        fmt.Printf("DRY RUN - Would pull secret: %s\n", secretID)
        fmt.Printf("Would create files:\n")
        fmt.Printf("  - %s\n", metadataPath)
        fmt.Printf("  - %s\n", contentsPath)
        return nil
    }
    
    // Actual operation
    ...
}

// push/push.go
func pushJSONDoc(cfg Config, log *zerolog.Logger) error {
    secret := buildSecret(cfg)
    
    if cfg.DryRun {
        fmt.Printf("DRY RUN - Would push to: %s\n", secret.SecretID())
        fmt.Printf("\nChanges that would be made:\n")
        fmt.Printf("%s\n", generateDiff(local, remote))
        return nil
    }
    
    // Actual operation
    ...
}
```

**Benefits**:
- Safe exploration for new users
- Validate configuration before committing
- Useful in scripts
- Standard practice in CLI tools

---

### 3.3 Random Confirmation Strings (KEEP AS-IS)

**Current State**:
```
Type 'aB3x' to confirm overwrite:
```

**Decision**: KEEP random confirmation strings.

**Rationale**:
- Prevents "confirmation fatigue" where users automatically type "yes"
- Forces users to actually read the diff/contents
- Random string can't be automated accidentally
- Small inconvenience for significant safety benefit

**Recommendation**: Enhance context around the random string but keep the mechanism:

```go
// push/confirm.go - Improved messaging

func ConfirmSecretUpdate(secretID string, diff string, randomString string) (bool, error) {
    fmt.Printf(`
═══════════════════════════════════════════════════════════
REVIEW REQUIRED - About to UPDATE secret: %s

Changes to be made:
%s
═══════════════════════════════════════════════════════════

To prevent accidental updates, please type this confirmation code: %s
(This random code ensures you've reviewed the changes above)

Confirmation code: `, secretID, diff, randomString)
    
    reader := bufio.NewReader(os.Stdin)
    response, err := reader.ReadString('\n')
    if err != nil {
        return false, err
    }
    
    return strings.TrimSpace(response) == randomString, nil
}
```

**Improvements**:
- Explain WHY random string is used
- Make it clear it's a safety feature
- Frame it as "review required" not just "confirm"
- Keep the anti-automation benefit

---

### 3.4 Inconsistent File Naming (MEDIUM USER IMPACT)

**Problem**: Different secret types use different file patterns.

**Current State**:
- JSDoc: `jsondoc.dev.app.metadata.json` + `.contents.json` (2 files)
- TextFile: `textfile.dev.app.metadata.json` + `.contents.txt` (2 files)
- SSLCert: `sslcert.dev.example.com.metadata.json` + `.crt` + `.key` (3 files)
- RDSPostgres: `rdspostgres.dev.inst.db.acc.json` (1 file with embedded metadata)
- Snowflake: `snowflake.dev.warehouse.acc.json` (1 file with embedded metadata)

**Recommendation**:

Standardize on one approach for all types:

**Option A: Always Separate Metadata** (Recommended)
```
jsondoc.dev.app.metadata.json
jsondoc.dev.app.data.json

textfile.dev.app.metadata.json
textfile.dev.app.data.txt

sslcert.dev.example.com.metadata.json
sslcert.dev.example.com.data.crt
sslcert.dev.example.com.data.key

rdspostgres.dev.inst.db.acc.metadata.json
rdspostgres.dev.inst.db.acc.data.json

snowflake.dev.warehouse.acc.metadata.json
snowflake.dev.warehouse.acc.data.json
```

**Benefits**:
- Predictable pattern
- Easy to write scripts
- Clear which file has what
- Consistent with jsondoc/textfile pattern

**Option B: Always Embed Metadata**
```
jsondoc.dev.app.json  (contains metadata + data)
textfile.dev.app.json (contains metadata + data)
sslcert.dev.example.com.json (contains metadata + cert + key as JSON fields)
rdspostgres.dev.inst.db.acc.json (already this way)
snowflake.dev.warehouse.acc.json (already this way)
```

**Benefits**:
- Single file per secret
- Easier to manage
- Consistent with rdspostgres/snowflake pattern

**Document the chosen pattern clearly in README.md**

---

### 3.5 No Progress Indication (LOW USER IMPACT)

**Problem**: Long operations provide no feedback.

**Recommendation**:

Add progress messages for operations >1 second:

```go
// push/push.go

func pushJSONDoc(cfg Config, log *zerolog.Logger) error {
    secret := buildSecret(cfg)
    secretID := secret.SecretID()
    
    fmt.Printf("Checking if secret exists: %s\n", secretID)
    exists := secret.Exists(log)
    
    if !exists {
        // Creation flow with progress
        fmt.Println("Secret does not exist, preparing to create...")
        displayData(secret.Data())
        confirmed := promptForConfirmation()
        if !confirmed {
            fmt.Println("Creation cancelled.")
            return nil
        }
        
        fmt.Printf("Creating secret in AWS Secrets Manager...")
        err := secret.Create(log)
        if err != nil {
            return err
        }
        fmt.Println(" ✓")
        fmt.Printf("Successfully created: %s\n", secretID)
        return nil
    }
    
    // Update flow with progress
    fmt.Printf("Fetching current secret from AWS...")
    remoteSecret, err := fetchRemote(secretID)
    if err != nil {
        return err
    }
    fmt.Println(" ✓")
    
    fmt.Println("Comparing local and remote versions...")
    diff := generateDiff(secret, remoteSecret)
    if diff == "" {
        fmt.Println("No changes detected.")
        return nil
    }
    
    fmt.Println("\nChanges detected:")
    fmt.Println(diff)
    
    confirmed := promptForConfirmation()
    if !confirmed {
        fmt.Println("Update cancelled.")
        return nil
    }
    
    fmt.Printf("Updating secret in AWS Secrets Manager...")
    err = secret.Update(true, log)
    if err != nil {
        return err
    }
    fmt.Println(" ✓")
    fmt.Printf("Successfully updated: %s\n", secretID)
    
    return nil
}
```

**Benefits**:
- User knows what's happening
- Reduces anxiety during slow operations
- Clear success/failure indication
- Professional user experience

---

### 3.6 Standardize Output Format (MEDIUM USER IMPACT)

**Problem**: Different commands show output in different formats.

**Recommendation**:

Create consistent output formatter:

```go
// tools/output.go

type Output struct {
    writer io.Writer
}

func NewOutput(w io.Writer) *Output {
    return &Output{writer: w}
}

func (o *Output) Success(message string) {
    fmt.Fprintf(o.writer, "✓ %s\n", message)
}

func (o *Output) Error(message string) {
    fmt.Fprintf(o.writer, "✗ %s\n", message)
}

func (o *Output) Info(message string) {
    fmt.Fprintf(o.writer, "• %s\n", message)
}

func (o *Output) Header(message string) {
    fmt.Fprintf(o.writer, "\n═══ %s ═══\n\n", message)
}

func (o *Output) FilesList(title string, files []string) {
    fmt.Fprintf(o.writer, "%s:\n", title)
    for _, file := range files {
        fmt.Fprintf(o.writer, "  • %s\n", file)
    }
}

func (o *Output) NextSteps(steps []string) {
    fmt.Fprintln(o.writer, "\nNext steps:")
    for i, step := range steps {
        fmt.Fprintf(o.writer, "  %d. %s\n", i+1, step)
    }
}

// Usage in all commands
out := tools.NewOutput(os.Stdout)

out.Header("Pulling Secret")
out.Info("Fetching from AWS...")
out.Success("Secret fetched successfully")
out.FilesList("Created files", []string{metadataFile, contentsFile})
out.NextSteps([]string{
    "Edit the contents file",
    "Run: sh-push -metadata=<file>",
})
```

**Benefits**:
- Consistent look and feel
- Easy to change formatting globally
- Testable (inject mock writer)
- Professional appearance

---

## Priority 4: Code Quality Issues

### 4.1 Excessive Use of Panic (MEDIUM IMPACT)

**Problem**: 24+ panic calls for error conditions that should be handled gracefully.

**Locations**:
- All secret type packages (jsondoc, textfile, etc.)
- tools/config.go
- All cmd/*/main.go files

**Current State**:
```go
// Bad: Library code panicking
func GetAWSAccountNumber() string {
    cfg, err := config.LoadDefaultConfig(context.Background())
    if err != nil {
        panic(fmt.Errorf("unable to load SDK config, %v", err))
    }
    // ...
}

// Bad: Panic then exit (exit is unreachable)
func main() {
    cfg, err := GetConfig()
    if err != nil {
        panic(err)
    }
    
    log := cfg.GetLogger()
    if err := process(); err != nil {
        log.Fatal().Err(err).Msg("error")
        os.Exit(1)  // Never reached!
    }
}
```

**Recommendation**:

Return errors properly:

```go
// Good: Library returns errors
func GetAWSAccountNumber() (string, error) {
    cfg, err := config.LoadDefaultConfig(context.Background())
    if err != nil {
        return "", fmt.Errorf("unable to load SDK config: %w", err)
    }
    
    client := sts.NewFromConfig(cfg)
    resp, err := client.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
    if err != nil {
        return "", fmt.Errorf("unable to get caller identity: %w", err)
    }
    
    return *resp.Account, nil
}

// Good: Main handles errors gracefully
func main() {
    cfg, err := GetConfig()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
        os.Exit(1)
    }
    
    log, err := cfg.GetLogger()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Logger initialization error: %v\n", err)
        os.Exit(1)
    }
    
    if err := process(); err != nil {
        log.Error().Err(err).Msg("processing error")
        os.Exit(1)
    }
}
```

**Benefits**:
- Testable error handling
- Graceful error messages
- No stack traces for expected errors
- Caller can handle appropriately

---

### 4.2 Dead Code After log.Fatal() (LOW IMPACT, EASY FIX)

**Problem**: `log.Fatal()` calls `os.Exit()` internally, making subsequent code unreachable.

**Locations**:
- cmd/sh-pull/main.go:20-22
- cmd/sh-push/main.go:19-22
- cmd/sh-generate/main.go:19-22

**Current State**:
```go
if err != nil {
    log.Fatal().Err(err).Msg("error")
    os.Exit(1)  // Never reached!
}
```

**Recommendation**:

Choose one pattern:

```go
// Option 1: Use Fatal (exits automatically)
if err != nil {
    log.Fatal().Err(err).Msg("error")
}

// Option 2: Use Error + Exit (explicit)
if err != nil {
    log.Error().Err(err).Msg("error")
    os.Exit(1)
}
```

**Recommended**: Use Option 2 for clarity.

---

## Implementation Priority & Progress Tracking

### ✅ COMPLETED PHASES

**Phase 0: Remove Obsolete Commands** ✅ DONE
1. ✅ Delete 6 obsolete commands (sh-download + 5 type-specific)
2. ✅ Update Makefile (remove from EXECUTABLES)
3. ✅ Delete get/ package (was unused)
4. ✅ Update README.md documentation
5. ✅ Remove examples for deleted commands

**Phase 2 (Partial): Foundation for Testing** ✅ DONE
10. ✅ Introduce interfaces (SecretsManager, FileSystem, Secret)
    - Created secrets/interfaces.go
    - Created secrets/aws_impl.go (AWS implementation)
    - Created secrets/operations.go (GenericExists, GenericCreate, GenericUpdate)
12. ✅ Create generic Secret operations

**Phase 4 (Partial): Usability Improvements** ✅ DONE
16. ✅ Improve error messages (pull/push configs have examples)
17. ⏭️  Add dry-run mode (SKIPPED - user decision)
18. ✅ Standardize output format (tools/output.go created)
19. ✅ Improve confirmation messaging (already good with random strings)

**Phase 5: Polish** ✅ DONE
20. ✅ Fix panics (all production panic() replaced with graceful errors)
21. ✅ Fix dead code (removed unreachable os.Exit() after log.Fatal())
22. ✅ Add progress indicators (pull/push show progress)
23. ✅ Add named constants (tools/constants.go created)

**Additional: Logging Simplification** ✅ DONE
24. ✅ Replace zerolog with standard Go logging (tools/logger.go)
    - Removed external dependency
    - Simpler printf-style logging
    - Better for interactive CLI tools

---

### 🚧 REMAINING WORK

**Phase 1: Cleanup Enabled by Removals** ← NEXT PRIORITY
6. ❌ Move CSV parsing to upload/ package (consolidate)
   - Currently: CSV parsing in each type package (*csv.go files)
   - Goal: Single CSV parser in upload/ package
   - Impact: ~500 lines of duplicate code removed
   
7. ❌ Delete CSV code from type packages
   - Delete: jsondoc/jsondoccsv.go
   - Delete: textfile/textfilecsv.go
   - Delete: sslcert/sslcertcsv.go
   - Delete: rdspostgres/rdspostgrescsv.go
   - Delete: snowflake/snowflakecsv.go
   
8. ❌ Streamline type packages to ~80 lines each
   - After CSV removal, type packages become much simpler
   - Currently ~200 lines, target ~80 lines
   
9. ❌ Run deadcode analysis and clean up

**Phase 2 (Remaining): Business Logic Separation**
11. ❌ Separate business logic from I/O in pull/push
    - Extract pure functions to secretlogic/ package
    - Enable unit testing without AWS/filesystem
    - Would increase test coverage to 80%+

**Phase 3: Use Generic Operations** ← HIGH IMPACT
13. ❌ Migrate type packages to use generic operations
    - Currently: Each type implements Exists/Create/Update directly
    - Goal: Delegate to secrets.GenericExists/Create/Update
    - Impact: Type packages reduced from ~200 to ~50 lines
    - **Note**: Interfaces exist, but not yet used!
    
14. ❌ Reduce secret type packages to interface implementations
    - After using generic ops, types are just data + SecretID()
    
15. ❌ Consolidate pull/push functions
    - Potentially one generic pull/push instead of 5 each

**Phase 4 (Remaining): Usability**
17. ⏭️  Add dry-run mode (EXPLICITLY SKIPPED BY USER)

**Additional Readability**
- ❌ Break up long functions (pushSSLCert is 138 lines)
- ❌ Standardize file naming (some use split files, some don't)

---

### 📊 Current Status vs Target

| Metric | Before | Current | Target (All Phases) |
|--------|--------|---------|---------------------|
| Commands | 10 | **4** ✅ | 4 |
| Total Lines | ~2500 | **~1500** 🟡 | ~600 |
| Duplicate Code | ~1900 | **~900** 🟡 | ~0 |
| Test Coverage | 0% | **86.9% (secretlogic only)** 🟡 | 80%+ (all packages) |
| External Deps | zerolog | **None** ✅ | Minimal |
| User Clarity | Confusing | **Clear** ✅ | Excellent |

**Progress: 50% Complete** 🎯

### 💡 Quick Wins Available

High impact, relatively easy:
1. **CSV Consolidation** (Phase 1) - Would remove ~500 lines immediately
2. **Use Generic Operations** (Phase 3) - Interfaces already exist, just need to use them
3. **Break Up Long Functions** - Improves readability without changing behavior

## Testing Strategy

After implementing recommendations:

**Unit Tests** (fast, no external dependencies):
- All business logic in secretlogic/
- Secret ID generation
- File path building
- Diff generation
- Confirmation parsing

**Integration Tests** (with mocks):
- Pull/push operations with mock SecretsManager
- File operations with mock FileSystem
- Complete workflows with mocks

**End-to-End Tests** (real AWS, optional):
- One test per secret type
- Create → Pull → Modify → Push → Verify
- Run in CI with test AWS account

## Measuring Success

### Before Cleanup:
- **Commands**: 10 total (4 modern + 6 obsolete)
- **Code volume**: 
  - ~1000 lines duplicated across 5 secret type packages
  - ~500 lines duplicated CSV parsing
  - ~400 lines in get/ package
  - Total duplicate code: ~1900 lines
- **Testing**: 0% unit test coverage of pull/push logic
- **Test speed**: Manual testing only
- **User experience**: Inconsistent (multiple overlapping tools)

### After Phase 0 (Remove Obsolete):
- **Commands**: 4 total (clean, purposeful set)
- **Code volume**: 
  - Remove 6 commands (~800 lines)
  - Remove get/ package (~400 lines if unused)
  - Consolidate CSV (~500 lines to ~200 lines)
  - Net reduction: ~1500 lines (about 20% of codebase)
- **Testing**: Same coverage (but less to test)
- **User experience**: Clear command purposes

### After All Phases:
- **Commands**: 4 total (sh-upload, sh-pull, sh-push, sh-generate)
- **Code volume**:
  - ~200 lines generic secret operations
  - ~80 lines per secret type (was ~200)
  - Total: ~600 lines (was ~2500)
  - **76% code reduction**
- **Testing**: 80%+ unit test coverage
- **Test speed**: <1 second automated tests
- **User experience**: Consistent, helpful, modern

### Impact Summary:

| Metric | Before | After Phase 0 | After All Phases |
|--------|--------|---------------|------------------|
| Commands | 10 | 4 | 4 |
| Total Lines | ~2500 | ~1000 | ~600 |
| Duplicate Code | ~1900 | ~900 | ~0 |
| Test Coverage | 0% | 0% | 80%+ |
| Test Speed | Manual | Manual | <1s |
| User Clarity | Confusing | Clear | Excellent |

## Questions?

For questions about these recommendations:
1. Open an issue with the "code-cleanup" label
2. Reference the specific recommendation number
3. Propose alternatives if you disagree with approach
