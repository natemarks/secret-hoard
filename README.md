# secret-hoard

AWS Secrets Manager management tool with interactive workflows and safety checks.

## Overview

secret-hoard manages AWS Secrets Manager secrets with support for five secret types:
- **rdspostgres** - Database credentials (RDS PostgreSQL)
- **snowflake** - Data warehouse credentials
- **ssl_certificate** - TLS certificates and private keys
- **jsondoc** - JSON documents
- **text_file** - Text file contents

Each secret is tagged with metadata and stored in a consistent format for easy retrieval and rotation.

## Philosophy

**Safety First**: All uploads require review and confirmation. No bulk uploads without human oversight.

**Interactive & Scriptable**: Commands support both interactive prompts and command-line flags.

**Working Directory**: All files stored in `~/.secret-hoard/` with predictable naming.

## Commands

secret-hoard provides 4 commands that work for all secret types:

---

### sh-generate - Create Scaffolding

Interactively generate local file templates for new secrets.

```bash
sh-generate
```

**Interactive prompts:**
1. Select secret type (1-5)
2. Enter environment (dev, staging, prod, etc.)
3. Enter type-specific metadata (instance, database, access, etc.)

**Output:**
Creates template files in `~/.secret-hoard/` with REPLACE-ME placeholders:
- `<type>.<env>.<metadata>.metadata.json` - Metadata file
- `<type>.<env>.<metadata>.<ext>` - Contents file(s)

**Example:**
```bash
sh-generate
# Select: 1 (rdspostgres)
# Environment: dev
# Instance: mydb
# Database: appdb
# Access: readonly

# Creates: ~/.secret-hoard/rdspostgres.dev.mydb.appdb.readonly.json
```

**Next steps:**
1. Edit the generated file(s)
2. Replace REPLACE-ME values with actual data
3. Use `sh-push` to upload

---

### sh-pull - Download Secrets

Download a secret to local editable files in `~/.secret-hoard/`.

**Interactive mode:**
```bash
sh-pull

# Prompts:
# Select secret type: 1-5
# Environment: dev
# (Type-specific fields: instance, database, access, etc.)
```

**Flag mode (scriptable):**
```bash
# RDS PostgreSQL
sh-pull -type=rdspostgres -env=dev -instance=mydb -database=appdb -access=readonly

# Snowflake
sh-pull -type=snowflake -env=prod -warehouse=analytics -access=readonly

# SSL Certificate
sh-pull -type=ssl_certificate -env=prod -commonname=example.com

# JSON Document
sh-pull -type=jsondoc -env=dev -access=app-config

# Text File
sh-pull -type=text_file -env=prod -access=api-key
```

**Output:**
Creates files in `~/.secret-hoard/` with predictable names:
- Metadata file with secret metadata and checksums
- Content file(s) with actual secret data

**Example:**
```bash
sh-pull -type=jsondoc -env=dev -access=app-config

# Creates:
#   ~/.secret-hoard/jsondoc.dev.app-config.metadata.json
#   ~/.secret-hoard/jsondoc.dev.app-config.contents.json
```

---

### sh-push - Upload with Diff & Confirmation

Upload local files to AWS with diff review and confirmation.

```bash
sh-push -metadata=<path-to-metadata-file>
```

**Features:**
- **Auto-creates** if secret doesn't exist (with proper tags)
- **Shows diff** if secret exists (highlighting changes)
- **Requires confirmation** with random code (prevents accidents)
- Works with all 5 secret types

**Example (create new secret):**
```bash
sh-push -metadata=~/.secret-hoard/jsondoc.dev.app-config.metadata.json

# Output:
# Secret does not exist. Creating new secret: jsondoc/dev/app-config
#
# New secret contents:
# {
#   "JSONContents": "{...}",
#   "JSONSha256Sum": "abc123..."
# }
#
# Type 'aB3x' to confirm: aB3x
#
# Secret created successfully: jsondoc/dev/app-config
```

**Example (update existing secret):**
```bash
sh-push -metadata=~/.secret-hoard/jsondoc.dev.app-config.metadata.json

# Output:
# Comparing local files to remote secret: jsondoc/dev/app-config
#
# {
#   "JSONContents": "...",
# - "JSONSha256Sum": "abc123...",
# + "JSONSha256Sum": "def456..."
# }
#
# Type 'xY7z' to confirm: xY7z
#
# Secret updated successfully: jsondoc/dev/app-config
```

---

### sh-contents - Download Secret Contents

Download secret contents directly to files for use in scripts. Outputs absolute paths to stdout for easy scripting.

**Supported types:** jsondoc, textfile, sslcert

**Secret ID Formats:**
- User-friendly: `textfile/`, `sslcert/` (recommended)
- AWS native: `text_file/`, `ssl_certificate/` (also supported)
- Both formats work identically - the tool normalizes automatically

**Usage:**
```bash
sh-contents [OPTIONS] <secret-id> <target-directory>
```

**Options:**
- `-debug` - Show verbose output (version, progress, debug info)

**Output Modes:**
- **Normal mode (default):** Prints only the absolute path(s) to stdout - perfect for scripts
- **Debug mode (-debug):** Shows version, progress messages, and paths

**Examples:**

```bash
# Normal mode - minimal output (only path)
$ sh-contents jsondoc/dev/app-config /tmp
/tmp/jsondoc.dev.app-config.contents.json

# Debug mode - verbose output
$ sh-contents -debug jsondoc/dev/app-config /tmp
sh-contents version: abc123
Fetching secret: jsondoc/dev/app-config
Writing to: /tmp/jsondoc.dev.app-config.contents.json
Successfully wrote jsondoc contents
/tmp/jsondoc.dev.app-config.contents.json

# Text File - outputs single file path
$ sh-contents textfile/prod/api-key /tmp
/tmp/textfile.prod.api-key.contents.txt

# SSL Certificate - outputs base path (append .crt or .key)
$ sudo sh-contents sslcert/prod/example.com /etc/ssl
/etc/ssl/sslcert.prod.example.com

# Files created on disk:
$ ls -la /etc/ssl/sslcert.prod.example.com.*
-rw-r--r-- 1 root root 1234 Jun 25 10:00 /etc/ssl/sslcert.prod.example.com.crt
-rw------- 1 root root 1679 Jun 25 10:00 /etc/ssl/sslcert.prod.example.com.key

# Note: AWS formats also work (automatically normalized)
$ sh-contents text_file/prod/api-key /tmp
$ sh-contents ssl_certificate/prod/example.com /etc/ssl
```

**Bash scripting:**

The minimal output mode makes scripting simple - just capture the path:

```bash
# JSON config - no need to filter stderr, output is already clean
CONFIG=$(sh-contents jsondoc/dev/app-config /tmp)
cat "$CONFIG"

# Text file content
API_KEY_FILE=$(sh-contents textfile/prod/api-key /tmp)
export API_KEY=$(cat "$API_KEY_FILE")

# SSL certificate - append extensions
CERT=$(sudo sh-contents sslcert/prod/example.com /etc/nginx/ssl)
cat > /etc/nginx/conf.d/ssl.conf <<EOF
ssl_certificate ${CERT}.crt;
ssl_certificate_key ${CERT}.key;
EOF

# Error handling
if CONFIG=$(sh-contents jsondoc/dev/app-config /tmp 2>/dev/null); then
    echo "Success: $CONFIG"
else
    echo "Failed to fetch secret" >&2
    exit 1
fi
```

**Format Flexibility:**

The `sh-contents` command accepts secret IDs in either user-friendly or AWS-native formats:

| Type | User-Friendly | AWS Format | Both Work |
|------|---------------|------------|-----------|
| Text File | `textfile/` | `text_file/` | ✅ |
| SSL Certificate | `sslcert/` | `ssl_certificate/` | ✅ |
| JSON Document | `jsondoc/` | `jsondoc/` | N/A (same) |

The tool automatically normalizes to AWS format before API calls. Use whichever format you prefer - they work identically.

**Data Integrity:**

Every file written by `sh-contents` is automatically verified using SHA256 checksums:
- Checksum calculated before writing
- File written to disk
- Checksum verified after writing
- Operation fails with exit code 1 if corruption detected

This protects against:
- Disk corruption or hardware failures
- Filesystem issues
- Partial writes due to disk space issues
- Concurrent modification

If verification fails, the error shows both expected and actual checksums for debugging.

**SSL Certificate Security:**
- Certificate file (`.crt`): 644 permissions (world-readable)
- Private key file (`.key`): 600 permissions (owner-only)
- If run as root: files owned by root:root
- Private keys are automatically secured with restrictive permissions

**Key differences from sh-pull:**
- sh-pull: Downloads to `~/.secret-hoard/` for editing
- sh-contents: Downloads to specified directory for immediate use
- sh-contents: Designed for automation and scripts
- sh-contents: Outputs paths to stdout (stderr for logs)

---

## Workflows

### New Secret (Interactive)

```bash
# 1. Generate scaffolding
sh-generate
# Select type, enter metadata

# 2. Edit the generated files
cd ~/.secret-hoard
vim <generated-file>

# 3. Upload to AWS (auto-creates with tags)
sh-push -metadata=<generated-metadata-file>
```

### Update Existing Secret

```bash
# 1. Download current version
sh-pull -type=jsondoc -env=dev -access=app-config

# 2. Edit the files
cd ~/.secret-hoard
vim jsondoc.dev.app-config.contents.json

# 3. Upload with diff review
sh-push -metadata=jsondoc.dev.app-config.metadata.json
# Reviews diff, confirms, updates
```

### Bulk Operations

For uploading multiple secrets, use a shell script with sh-push:

```bash
#!/bin/bash
# bulk-upload.sh - Upload multiple secrets with safety checks

SECRETS=(
  "jsondoc.dev.app-config.metadata.json"
  "jsondoc.dev.database-config.metadata.json"
  "rdspostgres.dev.mydb.appdb.readonly.json"
)

for secret in "${SECRETS[@]}"; do
  echo "═══════════════════════════════════════"
  echo "Uploading: $secret"
  echo "═══════════════════════════════════════"
  sh-push -metadata="$HOME/.secret-hoard/$secret"
  
  if [ $? -ne 0 ]; then
    echo "✗ Failed to upload $secret"
    exit 1
  fi
  echo "✓ Uploaded $secret"
  echo ""
done

echo "✓ All secrets uploaded successfully"
```

**Benefits of scripted approach:**
- Each secret reviewed individually (safer)
- Clear audit trail of what was uploaded
- Easy to skip or retry individual secrets
- No CSV format to maintain

---

## Secret Types

### RDS PostgreSQL (rdspostgres)

**Metadata:** environment, instance, database, access  
**Secret ID:** `rdspostgres/<env>/<instance>/<database>/<access>`  
**Data:** password, engine, port, dbInstanceIdentifier, host, username

**Example:**
```
Secret ID: rdspostgres/dev/mydb/appdb/readonly
Tags: ResourceType=rdspostgres, Environment=dev, Instance=mydb, Database=appdb, Access=readonly
Data: {"password": "...", "engine": "postgres", "port": 5432, ...}
```

### Snowflake (snowflake)

**Metadata:** environment, warehouse, access  
**Secret ID:** `snowflake/<env>/<warehouse>/<access>`  
**Data:** password, accountName, warehouse, username

**Example:**
```
Secret ID: snowflake/prod/analytics/readonly
Tags: ResourceType=snowflake, Environment=prod, Warehouse=analytics, Access=readonly
Data: {"password": "...", "accountName": "mycompany.us-east-1", ...}
```

### SSL Certificate (ssl_certificate)

**Metadata:** environment, commonName  
**Secret ID:** `ssl_certificate/<env>/<commonName>`  
**Data:** certificate (PEM), privateKey (PEM), expirationDate, modulus, certificateSha256, privateKeySha256

**Example:**
```
Secret ID: ssl_certificate/prod/example.com
Tags: ResourceType=ssl_certificate, Environment=prod, CommonName=example.com
Data: {"certificate": "-----BEGIN CERTIFICATE-----\n...", "privateKey": "-----BEGIN PRIVATE KEY-----\n...", ...}
```

### JSON Document (jsondoc)

**Metadata:** environment, access  
**Secret ID:** `jsondoc/<env>/<access>`  
**Data:** JSONContents (as string), JSONSha256Sum

**Example:**
```
Secret ID: jsondoc/dev/app-config
Tags: ResourceType=jsondoc, Environment=dev, Access=app-config
Data: {"JSONContents": "{\"key\": \"value\"}", "JSONSha256Sum": "abc123..."}
```

### Text File (text_file)

**Metadata:** environment, access  
**Secret ID:** `text_file/<env>/<access>`  
**Data:** Contents (as string), Sha256Sum

**Example:**
```
Secret ID: text_file/prod/api-key
Tags: ResourceType=text_file, Environment=prod, Access=api-key
Data: {"Contents": "sk-1234567890abcdef", "Sha256Sum": "def456..."}
```

---

## Working Directory

All commands use `~/.secret-hoard/` as the default working directory for local files. This directory is created automatically if it doesn't exist.

**File naming convention:**
- Metadata: `<type>.<env>.<metadata-fields>.metadata.json`
- Contents: `<type>.<env>.<metadata-fields>.<ext>`

**Examples:**
```
~/.secret-hoard/
  rdspostgres.dev.mydb.appdb.readonly.json
  jsondoc.dev.app-config.metadata.json
  jsondoc.dev.app-config.contents.json
  sslcert.prod.example.com.metadata.json
  sslcert.prod.example.com.crt
  sslcert.prod.example.com.key
```

---

## Installation

### Build from source

```bash
# Build all commands
make build

# Or build individually
go build -o bin/sh-pull ./cmd/sh-pull
go build -o bin/sh-push ./cmd/sh-push
go build -o bin/sh-generate ./cmd/sh-generate
go build -o bin/sh-contents ./cmd/sh-contents

# Add to PATH
export PATH=$PATH:$(pwd)/bin
```

### Prerequisites

- Go 1.21 or later
- AWS credentials configured (for sh-upload, sh-pull, sh-push)
- No AWS credentials needed for sh-generate

---

## Development

### Run tests

```bash
# Unit tests (no AWS required, fast)
make unittest

# All tests
make test

# Static analysis
make static

# Install pre-commit hook
./scripts/git-commithook.sh install
```

### Project structure

```
secret-hoard/
├── cmd/                    # Command-line executables
│   ├── sh-pull/           # Interactive/flag download
│   ├── sh-push/           # Upload with diff
│   ├── sh-generate/       # Scaffolding generator
│   └── sh-contents/       # Download contents for scripts
├── jsondoc/               # JSON document secret type
├── rdspostgres/           # RDS PostgreSQL secret type
├── snowflake/             # Snowflake secret type
├── sslcert/               # SSL certificate secret type
├── textfile/              # Text file secret type
├── contents/              # Contents download logic
├── pull/                  # Pull logic
├── push/                  # Push logic
├── generate/              # Generation logic
├── secretlogic/           # Pure business logic (testable)
└── tools/                 # Shared utilities
```

---

## Examples

See the `examples/` directory for sample CSV files for each secret type.

---

## Migration Guide

### From Old Commands (v1.x)

| Old Command | New Command | Notes |
|-------------|-------------|-------|
| `sh-jsondoc -file=...` | `sh-generate` + `sh-push` | Generate files, then push individually |
| `sh-rdsinstance -file=...` | `sh-generate` + `sh-push` | Generate files, then push individually |
| `sh-snowflake -file=...` | `sh-generate` + `sh-push` | Generate files, then push individually |
| `sh-sslcert -file=...` | `sh-generate` + `sh-push` | Generate files, then push individually |
| `sh-textfile -file=...` | `sh-generate` + `sh-push` | Generate files, then push individually |
| `sh-upload -file=...` | Script with `sh-push` | Use shell script for bulk (see above) |
| `sh-download -id=...` | `sh-pull -type=... -env=...` | Interactive or flag-based pull |

### Why Remove sh-upload?

The old CSV-based batch upload had several issues:
- **No review**: Uploaded all secrets without showing what changed
- **No safety**: Easy to accidentally overwrite production secrets
- **Complex CSV format**: Hard to maintain and error-prone

**New approach**: Generate files individually, review each, then push with confirmation. For bulk operations, use a simple shell script that calls `sh-push` in a loop - each secret gets reviewed individually.

### Migrating CSV Workflows

**Old CSV workflow:**
```bash
# secrets.csv with 10 secrets
sh-upload -file=secrets.csv -overwrite  # uploads all at once
```

**New safe workflow:**
```bash
# 1. Generate files for each secret
sh-generate  # interactive for each

# 2. Edit files as needed
cd ~/.secret-hoard
vim *.json

# 3. Upload with review (one at a time or scripted)
for f in *.metadata.json; do
  sh-push -metadata="$f"
done
```

**Benefits:**
- Each secret reviewed before upload
- Clear diff shown for changes
- Random confirmation prevents accidents
- Safer for production environments

---

## Contributing

See [CLEANUP.md](CLEANUP.md) for planned improvements and contribution opportunities.

## License

[Add your license here]
