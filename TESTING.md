# Testing Documentation

## Overview

The `secretlogic` package contains pure, testable business logic extracted from the main application. Tests use table-driven patterns and golden files for comprehensive coverage **without external dependencies** (no AWS, no filesystem I/O for tests).

## Test Organization

```
secretlogic/
├── secretid.go           # Secret ID generation
├── secretid_test.go      # 25+ table-driven test cases
├── filepath.go           # File path generation  
├── filepath_test.go      # File split strategy tests
├── metadata.go           # Metadata structure generation
├── metadata_test.go      # Golden file tests
└── testdata/
    └── metadata/         # Golden files for expected outputs
        ├── jsondoc_dev_app.golden.json
        ├── rdspostgres_dev_db1.golden.json
        ├── sslcert_prod_example.golden.json
        └── ...
```

## Running Tests

### Quick Test Run
```bash
make unittest
```

### With Coverage Report
```bash
make unittest-coverage
```
**Current Coverage: 86.9%**

### Update Golden Files
When logic changes require updating expected outputs:
```bash
make unittest-update
```

### Individual Test Functions
```bash
# Run specific test
go test ./secretlogic -run TestBuildSecretID

# Verbose output
go test -v ./secretlogic -run TestBuildMetadataMap

# Update golden files for one test
go test ./secretlogic -run TestBuildMetadataMap -update
```

---

## What Is Tested

### 1. Secret ID Generation (`TestBuildSecretID`)

**Tests**: 25+ combinations covering all 5 secret types

**Coverage:**
- ✅ All 5 secret types (jsondoc, text_file, ssl_certificate, snowflake, rdspostgres)
- ✅ Multiple environments (dev, staging, prod, integration)
- ✅ Various metadata combinations
- ✅ Special cases (wildcards in commonname)
- ✅ Error cases (missing fields, empty values, unknown types)

**Examples:**
```go
// Valid cases
"jsondoc/dev/app"
"text_file/prod/config"
"ssl_certificate/staging/*.example.com"
"snowflake/prod/analytics/readonly"
"rdspostgres/dev/db1/myapp/readwrite"

// Error cases
missing environment
missing required fields (access, instance, database, etc.)
empty field values
unknown secret types
```

**Why critical:** Wrong Secret ID = fetching/updating wrong secret in AWS

---

### 2. File Path Generation (`TestBuildBaseFilename`, `TestFilesToGenerate`)

**Tests**: 15+ combinations

**Coverage:**
- ✅ Base filename construction for all types
- ✅ File split strategy (metadata + contents vs single file)
- ✅ Correct extensions (.json, .txt, .crt, .key)
- ✅ Error cases

**Examples:**
```go
// Base filenames
"jsondoc.dev.app"
"rdspostgres.prod.primary.orders.readonly"
"ssl_certificate.staging.example.com"

// File lists
jsondoc → ["jsondoc.dev.app.metadata.json", "jsondoc.dev.app.contents.json"]
text_file → ["text_file.prod.config.metadata.json", "text_file.prod.config.contents.txt"]
sslcert → ["sslcert.staging.example.com.metadata.json", ".crt", ".key"]
rdspostgres → ["rdspostgres.dev.db1.app.ro.json"]  // single file
```

**Why critical:** Wrong filenames = can't find files to push

---

### 3. Metadata Structure Generation (`TestBuildMetadataMap`)

**Tests**: 7 golden file comparisons

**Coverage:**
- ✅ Correct field names for each type
- ✅ Correct field order (alphabetical in JSON)
- ✅ Required fields included
- ✅ Error cases (missing required fields)

**Golden Files:**
```json
// jsondoc_dev_app.golden.json
{
  "access": "app",
  "environment": "dev",
  "resourceType": "jsondoc"
}

// rdspostgres_dev_db1.golden.json
{
  "access": "readonly",
  "database": "myapp",
  "environment": "dev",
  "instance": "db1",
  "resourceType": "rdspostgres"
}
```

**Why critical:** Wrong field names/structure = can't parse in push operations

---

### 4. Required Fields Validation (`TestRequiredMetadataFields`)

**Tests**: 6 cases (one per type + error case)

**Coverage:**
- ✅ Returns correct field list for each type
- ✅ Correct field count (2-4 fields depending on type)
- ✅ Error case for unknown types

**Examples:**
```go
jsondoc      → ["environment", "access"]           // 2 fields
text_file    → ["environment", "access"]           // 2 fields  
ssl_cert     → ["environment", "commonname"]       // 2 fields
snowflake    → ["environment", "warehouse", "access"]  // 3 fields
rdspostgres  → ["environment", "instance", "database", "access"]  // 4 fields
```

**Why critical:** Used by interactive prompts to ask correct questions

---

## Test Patterns Used

### Pattern 1: Table-Driven Tests

```go
func TestBuildSecretID(t *testing.T) {
    tests := []struct {
        name       string
        secretType string
        metadata   map[string]string
        want       string
        wantErr    bool
    }{
        {
            name: "jsondoc_basic",
            secretType: "jsondoc",
            metadata: map[string]string{"environment": "dev", "access": "app"},
            want: "jsondoc/dev/app",
        },
        // ... 24 more cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := BuildSecretID(tt.secretType, tt.metadata)
            // ... assertions
        })
    }
}
```

**Benefits:**
- Easy to add new test cases
- Self-documenting (shows all supported patterns)
- Fast execution (no I/O)
- Clear failure messages

### Pattern 2: Golden File Tests

```go
func TestBuildMetadataMap(t *testing.T) {
    tests := []struct {
        name       string
        secretType string
        metadata   map[string]string
        goldenFile string
    }{
        {
            name: "jsondoc_dev_app",
            secretType: "jsondoc",
            metadata: map[string]string{"environment": "dev", "access": "app"},
            goldenFile: "testdata/metadata/jsondoc_dev_app.golden.json",
        },
        // ...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, _ := BuildMetadataMap(tt.secretType, tt.metadata)
            gotJSON, _ := json.MarshalIndent(got, "", "  ")
            
            wantJSON, _ := os.ReadFile(tt.goldenFile)
            
            if string(gotJSON) != string(wantJSON) {
                t.Errorf("mismatch\nGot:\n%s\n\nWant:\n%s", gotJSON, wantJSON)
            }
        })
    }
}
```

**Benefits:**
- Complex JSON structures validated
- Easy to review changes in PRs (git diff shows actual output changes)
- `-update` flag regenerates expected outputs
- Prevents unintended structure changes

---

## Test Statistics

### By Function
```
TestBuildSecretID           25 cases   ✓ All pass
TestBuildBaseFilename        9 cases   ✓ All pass
TestFilesToGenerate          6 cases   ✓ All pass
TestBuildMetadataMap         9 cases   ✓ All pass (7 with golden files)
TestRequiredMetadataFields   6 cases   ✓ All pass
──────────────────────────────────────
Total:                      55 cases   ✓ All pass
```

### By Secret Type
```
jsondoc         11 test cases
text_file        7 test cases
ssl_certificate  9 test cases
snowflake        8 test cases
rdspostgres     11 test cases
Error cases      9 test cases
```

### Coverage Metrics
```
Total Coverage:     86.9%
Functions Tested:   6/6 (100%)
Secret Types:       5/5 (100%)
Error Cases:        9 different scenarios
Golden Files:       7 files
```

---

## What Is NOT Tested (By Design)

These require integration or manual testing:

### ❌ AWS Operations
- Fetching secrets from Secrets Manager
- Updating secrets in AWS
- Tag management
- **Why:** External dependency, covered by manual testing

### ❌ File I/O
- Reading from filesystem
- Writing to filesystem
- Directory creation
- **Why:** External dependency, tested via `make build` and manual testing

### ❌ User Interaction
- Interactive prompts
- Confirmation strings
- User input validation
- **Why:** UI testing, covered by manual testing (see MANUAL-TEST.md)

### ❌ Certificate Parsing
- X.509 certificate parsing
- RSA key extraction
- Modulus comparison
- **Why:** Could be added but relies on crypto library correctness

---

## Benefits of This Test Structure

### For Development
1. **Fast iteration** - Tests run in ~3ms
2. **No setup required** - No AWS credentials, no test data
3. **Comprehensive** - 55 test cases covering all types
4. **Clear failures** - Table tests show exactly what failed

### For Maintenance
1. **Catch regressions** - Any logic change caught immediately
2. **Safe refactoring** - Tests validate behavior unchanged
3. **Documentation** - Tests show all supported patterns
4. **Easy updates** - `make unittest-update` regenerates golden files

### For Code Review
1. **Golden file diffs** - Reviewers see exact output changes in git diff
2. **Test coverage visible** - Can spot missing cases
3. **Intent clear** - Test names document purpose
4. **No flaky tests** - Pure functions = deterministic results

---

## Adding New Tests

### Adding a Table-Driven Test Case

```go
// In secretid_test.go, add to the tests slice:
{
    name:       "jsondoc_new_env",
    secretType: "jsondoc",
    metadata:   map[string]string{"environment": "qa", "access": "test"},
    want:       "jsondoc/qa/test",
},
```

### Adding a Golden File Test

1. Add test case:
```go
{
    name:       "snowflake_qa",
    secretType: "snowflake",
    metadata: map[string]string{
        "environment": "qa",
        "warehouse": "test",
        "access": "dev",
    },
    goldenFile: "testdata/metadata/snowflake_qa.golden.json",
},
```

2. Generate golden file:
```bash
make unittest-update
```

3. Review the generated file:
```bash
cat secretlogic/testdata/metadata/snowflake_qa.golden.json
```

4. Commit both test code and golden file

---

## Testing Workflow

### During Development
```bash
# Make changes to secretlogic/*.go
vim secretlogic/secretid.go

# Run tests
make unittest

# If output changed intentionally, update golden files
make unittest-update

# Verify changes
git diff secretlogic/testdata/
```

### Before Committing
```bash
# Run all unit tests
make unittest

# Check coverage
make unittest-coverage

# Run static checks
make static
```

### In CI/CD
```bash
# Unit tests (fast, no dependencies)
make unittest

# Integration tests (requires AWS)
# (see MANUAL-TEST.md)
```

---

## Future Test Additions

### Easy Wins (Recommend Adding)
1. **Diff generation tests** - Test JSON diff output format
2. **Template generation tests** - Test sh-generate output with golden files
3. **Confirmation string tests** - Validate random string generation (alphanumeric, length)

### Medium Effort
1. **Certificate parsing tests** - Test expiration/modulus extraction with test certs
2. **Error message tests** - Validate helpful error messages

### Lower Priority
1. **Path resolution tests** - Test working directory path logic
2. **File existence checks** - Would require mocking filesystem

---

## Troubleshooting

### Test Failures After Code Changes

**Symptom:** Golden file mismatch
```
BuildMetadataMap() mismatch
Got:
{
  "access": "app",
  "environment": "dev",
  "newField": "value",    ← Added field
  "resourceType": "jsondoc"
}

Want:
{
  "access": "app",
  "environment": "dev",
  "resourceType": "jsondoc"
}
```

**Solution:** If change is intentional:
```bash
make unittest-update
git diff secretlogic/testdata/  # Review changes
git add secretlogic/testdata/
```

### Golden File Missing

**Symptom:**
```
Golden file does not exist: testdata/metadata/new_test.golden.json 
(run with -update to create)
```

**Solution:**
```bash
make unittest-update
```

### Test Coverage Drop

**Symptom:** Coverage drops below 80%

**Solution:**
1. Identify uncovered code: `go test -coverprofile=coverage.out ./secretlogic`
2. View coverage: `go tool cover -html=coverage.out`
3. Add test cases for uncovered branches

---

## Comparison: Unit vs Integration vs Manual

| Aspect | Unit Tests | Integration | Manual |
|--------|-----------|-------------|---------|
| **Speed** | ~3ms | ~5-10s | Minutes |
| **Dependencies** | None | AWS | AWS + Human |
| **Coverage** | Business logic | AWS integration | Full workflows |
| **Frequency** | Every commit | Before merge | Before release |
| **Automation** | CI/CD | CI/CD | Checklist |
| **Examples** | `make unittest` | (future) | MANUAL-TEST.md |

**All three are important** - unit tests give fast feedback, integration tests validate AWS operations, manual tests verify user experience.

---

## Summary

✅ **55 test cases** covering critical business logic  
✅ **86.9% code coverage** of secretlogic package  
✅ **0 external dependencies** (no AWS, no filesystem)  
✅ **~3ms execution time** - instant feedback  
✅ **7 golden files** for complex output validation  
✅ **Make targets** for easy execution (`unittest`, `unittest-update`, `unittest-coverage`)  
✅ **Table-driven patterns** for easy maintenance  
✅ **All 5 secret types** comprehensively tested  

The test suite provides **high confidence** in the critical business logic without requiring AWS credentials or external setup.
