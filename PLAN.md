# Implementation Plan: Three Secret Synchronization Executables

## Context

The secret-hoard project currently supports batch CSV-based secret uploads via `sh-upload` and single-secret downloads via `sh-download`. This works well for bulk operations but lacks support for interactive, single-secret workflows that developers need for day-to-day secret management.

This plan adds three new executables to support interactive secret synchronization workflows:
1. **sh-pull**: Download a secret by metadata and create local editable files
2. **sh-push**: Upload local files after showing a diff and requiring confirmation
3. **sh-generate**: Interactively create local file scaffolding for a new secret

These tools enable a more iterative workflow: generate → edit → pull/compare → push with safety checks.

## Design Decisions

### File Naming Convention
All files follow the pattern: `{resourceType}.{metadata-fields}.{extension}`

Examples:
- `jsondoc.dev.idemia.metadata.json` + `jsondoc.dev.idemia.contents.json`
- `textfile.prod.myfile.metadata.json` + `textfile.prod.myfile.contents.txt`
- `sslcert.staging.example.com.metadata.json` + `sslcert.staging.example.com.crt` + `sslcert.staging.example.com.key`
- `rdspostgres.dev.mydb.master.json` (single file, no split)
- `snowflake.prod.warehouse.admin.json` (single file, no split)

### File Splitting Strategy

**Split into metadata + contents** (types with `Contents`/`JSONContents` + `Sha256Sum` fields):
- **jsondoc**: metadata.json (Metadata + JSONSha256Sum) + contents.json (the actual JSON document)
- **textfile**: metadata.json (Metadata + Sha256Sum) + contents.txt (the actual text file)
- **sslcert**: metadata.json (Metadata + all computed Data fields: expirationDate, modulus, certificateSha256, privateKeySha256) + .crt + .key files

**Single JSON file** (all other types):
- **rdspostgres**: Single JSON with both Metadata and Data (password, host, port, etc.)
- **snowflake**: Single JSON with both Metadata and Data (accountName, warehouse, username, password)

### Diff Format
JSON diff with +/- indicators showing before/after values for changed fields.

### Generate Output
Minimal files with REPLACE-ME placeholders for values that must be filled in.

---

## Implementation Plan

### 1. Create `cmd/sh-pull/` executable

**Purpose**: Download a secret and create local editable files.

**Files to create**:
- `cmd/sh-pull/main.go` - Entry point
- `cmd/sh-pull/config.go` - CLI configuration
- `pull/pull.go` - Core pull logic (new package)

**CLI flags**:
```
-type     string  Secret type (required)
-env      string  Environment (required)
-access   string  Access type (for jsondoc, textfile, snowflake)
-instance string  Instance (for rdspostgres)
-database string  Database (for rdspostgres)
-warehouse string Warehouse (for snowflake)
-commonname string Common name (for sslcert)
-debug    bool    Enable debug logging
```

**Flow**:
1. Parse flags and validate based on secret type
2. Build secret ID from metadata flags using existing patterns (e.g., `{type}/{env}/{instance}/{database}/{access}`)
3. Call `tools.GetSecretValue(secretID)` to fetch secret from AWS
4. Route to type-specific pull handler based on resource type
5. Create local files with naming pattern: `{type}.{metadata}.{extension}`

**Type-specific handlers**:

*For jsondoc*:
- Unmarshal secret value to `jsondoc.Data`
- Create `jsondoc.{env}.{access}.metadata.json` containing:
  ```json
  {
    "resourceType": "jsondoc",
    "environment": "dev",
    "access": "idemia",
    "JSONSha256Sum": "abc123..."
  }
  ```
- Create `jsondoc.{env}.{access}.contents.json` containing the actual JSON document from `JSONContents` field
- Verify SHA256 matches

*For textfile*:
- Unmarshal secret value to `textfile.Data`
- Create `textfile.{env}.{access}.metadata.json` containing:
  ```json
  {
    "resourceType": "text_file",
    "environment": "prod",
    "access": "myfile",
    "sha256Sum": "def456..."
  }
  ```
- Create `textfile.{env}.{access}.contents.txt` containing the text from `Contents` field
- Verify SHA256 matches

*For sslcert*:
- Unmarshal secret value to `sslcert.Data`
- Create `sslcert.{env}.{commonname}.metadata.json` containing:
  ```json
  {
    "resourceType": "ssl_certificate",
    "environment": "staging",
    "commonName": "example.com",
    "expirationDate": "2026-12-31T23:59:59Z",
    "modulus": "ABC123...",
    "certificateSha256": "xyz789...",
    "privateKeySha256": "uvw456..."
  }
  ```
- Create `sslcert.{env}.{commonname}.crt` containing certificate
- Create `sslcert.{env}.{commonname}.key` containing private key
- Verify SHA256 sums match

*For rdspostgres*:
- Create single `rdspostgres.{env}.{instance}.{database}.{access}.json` containing both Metadata and Data:
  ```json
  {
    "metadata": {
      "resourceType": "rdspostgres",
      "environment": "dev",
      "instance": "myinstance",
      "database": "mydb",
      "access": "master"
    },
    "data": {
      "password": "secret",
      "engine": "postgres",
      "port": 5432,
      "dbInstanceIdentifier": "...",
      "host": "...",
      "username": "admin"
    }
  }
  ```

*For snowflake*:
- Create single `snowflake.{env}.{warehouse}.{access}.json` containing both Metadata and Data

**Reuse existing code**:
- `tools.GetSecretValue()` - Fetch secret from AWS
- `tools.WriteStringToFile()` - Write file contents
- `tools.CheckSha256Sum()` - Verify integrity
- Unmarshal logic similar to `get/get.go` handlers

---

### 2. Create `cmd/sh-push/` executable

**Purpose**: Upload local files after showing diff and requiring user confirmation.

**Files to create**:
- `cmd/sh-push/main.go` - Entry point
- `cmd/sh-push/config.go` - CLI configuration
- `push/push.go` - Core push logic (new package)
- `push/diff.go` - Diff generation utilities (new package)
- `push/confirm.go` - Confirmation prompt with random string (new package)

**CLI flags**:
```
-metadata string  Path to metadata.json file (required)
-debug    bool    Enable debug logging
```

**Flow**:
1. Parse `-metadata` flag to get metadata file path
2. Read and parse metadata JSON to determine secret type and metadata fields
3. Build secret ID from metadata
4. Determine if contents file(s) are needed based on type:
   - jsondoc: look for `{basename}.contents.json`
   - textfile: look for `{basename}.contents.txt`
   - sslcert: look for `{basename}.crt` and `{basename}.key`
5. Read all local files and reconstruct the secret Data struct:
   - For jsondoc: compute SHA256 of contents.json
   - For textfile: compute SHA256 of contents.txt
   - For sslcert: extract expiration date, modulus from cert/key files, compute SHA256 sums
   - For rdspostgres/snowflake: read data directly from single JSON file
6. Fetch current secret from AWS using `tools.GetSecretValue(secretID)`
7. Generate JSON diff showing changes (use +/- indicators)
8. Display diff to user
9. Generate random 4-character confirmation string (alphanumeric)
10. Prompt user: "Type '{random-string}' to confirm overwrite:"
11. Read user input and compare (case-sensitive)
12. If match: proceed with update; if not: abort with message

**Diff format example**:
```
Comparing local files to remote secret: jsondoc/dev/idemia

{
  "metadata": {
    "resourceType": "jsondoc",
    "environment": "dev",
    "access": "idemia"
  },
  "data": {
-   "JSONContents": "{\"old\":\"value\"}",
+   "JSONContents": "{\"new\":\"value\"}",
-   "JSONSha256Sum": "abc123..."
+   "JSONSha256Sum": "def456..."
  }
}

Type 'A7x2' to confirm overwrite:
```

**Upload logic**:
- Construct appropriate Secret struct for the type (e.g., `jsondoc.Secret`, `rdspostgres.Secret`)
- Call existing `secret.Update(true, &log)` method from type packages
- This reuses all existing AWS SDK calls and tag management

**Reuse existing code**:
- `tools.GetSecretValue()` - Fetch current secret
- `tools.GetSHA256Sum()` - Compute file hashes
- `tools.ReadFileToString()` - Read file contents
- Type-specific `Secret.Update()` methods from rdspostgres, snowflake, jsondoc, textfile, sslcert packages
- SSL certificate utilities from `sslcert/sslcertcsv.go` (modulus extraction, expiration parsing)

**New utilities needed**:
- `push/diff.go`: `GenerateJSONDiff(local, remote interface{}) string` - Format diff with +/- indicators
- `push/confirm.go`: `GenerateRandomString(length int) string` and `PromptForConfirmation(expected string) bool`

---

### 3. Create `cmd/sh-generate/` executable

**Purpose**: Interactively generate local file scaffolding for a new secret.

**Files to create**:
- `cmd/sh-generate/main.go` - Entry point
- `cmd/sh-generate/config.go` - CLI configuration
- `generate/generate.go` - Core generation logic (new package)
- `generate/prompts.go` - Interactive prompts (new package)

**CLI flags**:
```
-debug bool  Enable debug logging
```

**Flow**:
1. Prompt: "Select secret type:"
   - Display numbered list:
     ```
     1. rdspostgres
     2. snowflake
     3. ssl_certificate
     4. jsondoc
     5. text_file
     ```
   - Read user selection (1-5)

2. Based on type, prompt for required metadata fields:
   - **All types**: Environment (e.g., dev, staging, production)
   - **rdspostgres**: Instance, Database, Access
   - **snowflake**: Warehouse, Access
   - **sslcert**: CommonName
   - **jsondoc**: Access
   - **textfile**: Access

3. Generate files with REPLACE-ME placeholders:

*For jsondoc*:
```json
// jsondoc.{env}.{access}.metadata.json
{
  "resourceType": "jsondoc",
  "environment": "dev",
  "access": "myaccess",
  "JSONSha256Sum": "REPLACE-ME"
}

// jsondoc.{env}.{access}.contents.json
{}
```

*For textfile*:
```json
// textfile.{env}.{access}.metadata.json
{
  "resourceType": "text_file",
  "environment": "prod",
  "access": "myfile",
  "sha256Sum": "REPLACE-ME"
}

// textfile.{env}.{access}.contents.txt
(empty file)
```

*For sslcert*:
```json
// sslcert.{env}.{commonname}.metadata.json
{
  "resourceType": "ssl_certificate",
  "environment": "staging",
  "commonName": "example.com",
  "expirationDate": "REPLACE-ME",
  "modulus": "REPLACE-ME",
  "certificateSha256": "REPLACE-ME",
  "privateKeySha256": "REPLACE-ME"
}

// sslcert.{env}.{commonname}.crt
(empty file)

// sslcert.{env}.{commonname}.key
(empty file)
```

*For rdspostgres*:
```json
// rdspostgres.{env}.{instance}.{database}.{access}.json
{
  "metadata": {
    "resourceType": "rdspostgres",
    "environment": "dev",
    "instance": "myinstance",
    "database": "mydb",
    "access": "master"
  },
  "data": {
    "password": "REPLACE-ME",
    "engine": "REPLACE-ME",
    "port": 5432,
    "dbInstanceIdentifier": "REPLACE-ME",
    "host": "REPLACE-ME",
    "username": "REPLACE-ME"
  }
}
```

*For snowflake*:
```json
// snowflake.{env}.{warehouse}.{access}.json
{
  "metadata": {
    "resourceType": "snowflake",
    "environment": "dev",
    "warehouse": "mywarehouse",
    "access": "readwrite"
  },
  "data": {
    "password": "REPLACE-ME",
    "accountName": "REPLACE-ME",
    "warehouse": "mywarehouse",
    "username": "REPLACE-ME"
  }
}
```

4. Log created files and next steps:
   ```
   Generated files:
     - jsondoc.dev.myaccess.metadata.json
     - jsondoc.dev.myaccess.contents.json
   
   Next steps:
     1. Edit the contents.json file with your JSON document
     2. Run: sh-push -metadata=jsondoc.dev.myaccess.metadata.json
   ```

**Reuse existing code**:
- `tools.WriteStringToFile()` - Write generated files
- Metadata field names and SecretID format patterns from type packages

**New utilities needed**:
- `generate/prompts.go`: `PromptForChoice(prompt string, options []string) int` and `PromptForString(prompt string) string`

---

## Critical Files

### Existing files to reference (read-only):
- `/home/nmarks/projects/secret-hoard/tools/helper.go` - AWS Secrets Manager operations
- `/home/nmarks/projects/secret-hoard/tools/file.go` - File I/O utilities
- `/home/nmarks/projects/secret-hoard/tools/config.go` - Config and logger patterns
- `/home/nmarks/projects/secret-hoard/get/get.go` - Download handlers (reference for pull)
- `/home/nmarks/projects/secret-hoard/rdspostgres/rdspostgres.go` - Secret struct and Update method
- `/home/nmarks/projects/secret-hoard/snowflake/snowflake.go` - Secret struct and Update method
- `/home/nmarks/projects/secret-hoard/jsondoc/jsondoc.go` - Secret struct and Update method
- `/home/nmarks/projects/secret-hoard/textfile/textfile.go` - Secret struct and Update method
- `/home/nmarks/projects/secret-hoard/sslcert/sslcert.go` - Secret struct and Update method
- `/home/nmarks/projects/secret-hoard/sslcert/sslcertcsv.go` - Certificate utilities (modulus, expiration)

### New files to create:
1. `cmd/sh-pull/main.go` - Pull executable entry point
2. `cmd/sh-pull/config.go` - Pull CLI configuration
3. `pull/pull.go` - Pull logic package
4. `cmd/sh-push/main.go` - Push executable entry point
5. `cmd/sh-push/config.go` - Push CLI configuration
6. `push/push.go` - Push logic package
7. `push/diff.go` - Diff generation
8. `push/confirm.go` - Confirmation prompt
9. `cmd/sh-generate/main.go` - Generate executable entry point
10. `cmd/sh-generate/config.go` - Generate CLI configuration
11. `generate/generate.go` - Generation logic package
12. `generate/prompts.go` - Interactive prompts

### Files to modify:
- `/home/nmarks/projects/secret-hoard/Makefile` - Add sh-pull, sh-push, sh-generate to EXECUTABLES list

---

## Verification

### Test sh-pull:
```bash
# Pull a jsondoc secret
sh-pull -type=jsondoc -env=dev -access=idemia -debug

# Verify files created:
ls -la jsondoc.dev.idemia.metadata.json
ls -la jsondoc.dev.idemia.contents.json

# Pull an sslcert secret
sh-pull -type=ssl_certificate -env=staging -commonname=example.com -debug

# Verify files created:
ls -la sslcert.staging.example.com.metadata.json
ls -la sslcert.staging.example.com.crt
ls -la sslcert.staging.example.com.key

# Pull an rdspostgres secret
sh-pull -type=rdspostgres -env=dev -instance=myinstance -database=mydb -access=master -debug

# Verify single file created:
ls -la rdspostgres.dev.myinstance.mydb.master.json
```

### Test sh-generate:
```bash
# Generate jsondoc scaffolding
sh-generate -debug
# Select: 4 (jsondoc)
# Enter environment: dev
# Enter access: test

# Verify files created with REPLACE-ME:
cat jsondoc.dev.test.metadata.json
cat jsondoc.dev.test.contents.json
```

### Test sh-push:
```bash
# Edit generated or pulled files
vim jsondoc.dev.idemia.contents.json

# Push with diff and confirmation
sh-push -metadata=jsondoc.dev.idemia.metadata.json -debug
# Should display diff
# Should prompt for random 4-char string
# Type string to confirm

# Verify secret updated in AWS:
sh-pull -type=jsondoc -env=dev -access=idemia
```

### Test all types:
1. Generate scaffolding for each type (rdspostgres, snowflake, sslcert, jsondoc, textfile)
2. Fill in REPLACE-ME values appropriately
3. Push each type and verify confirmation prompt works
4. Pull each type back and verify file structure matches expectations
5. Verify SHA256 sums are validated correctly for jsondoc, textfile, sslcert

### Integration test:
```bash
# Full workflow: generate → edit → push → pull → edit → push
sh-generate  # Create new jsondoc secret scaffolding
# Edit files
sh-push -metadata=jsondoc.dev.newtest.metadata.json  # Upload
sh-pull -type=jsondoc -env=dev -access=newtest  # Download to verify
# Edit contents.json
sh-push -metadata=jsondoc.dev.newtest.metadata.json  # Update with diff
```

---

## Summary

This plan implements three interactive CLI tools that complement the existing batch CSV workflow:
- **sh-pull**: Download secrets to editable local files with automatic file splitting for content-based types
- **sh-push**: Upload with safety (diff display + random confirmation string)
- **sh-generate**: Scaffold new secrets interactively without CSV

The design reuses existing Secret structs, Update methods, AWS SDK wrappers, and file utilities from the codebase. File naming follows a consistent pattern based on secret metadata, making it easy to identify which secret each file belongs to.
