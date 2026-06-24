# secret-hoard

AWS Secrets Manager management tool with batch and interactive workflows.

## Overview

secret-hoard manages AWS Secrets Manager secrets with support for five secret types:
- **rdspostgres** - Database credentials (RDS PostgreSQL)
- **snowflake** - Data warehouse credentials
- **ssl_certificate** - TLS certificates and private keys
- **jsondoc** - JSON documents
- **text_file** - Text file contents

Each secret is tagged with metadata and stored in a consistent format for easy retrieval and rotation.

## Commands

secret-hoard provides 4 commands that work for all secret types:

### sh-upload - Batch Upload from CSV

Upload multiple secrets of any type from a single CSV file.

```bash
sh-upload -file=secrets.csv -overwrite
```

The CSV format varies by secret type (ResourceType column determines the type):

**RDS PostgreSQL:**
```csv
ResourceType,Environment,Instance,Database,Access,Password,Engine,Port,DbInstanceIdentifier,Host,Username
rdspostgres,dev,mydb,appdb,readonly,pass123,postgres,5432,mydb-dev,mydb.aws.com,readonly_user
```

**Snowflake:**
```csv
ResourceType,Environment,Warehouse,Access,AccountName,Username,Password
snowflake,prod,analytics,readonly,mycompany.us-east-1,analytics_ro,pass456
```

**SSL Certificate:**
```csv
ResourceType,Environment,CommonName,CertificateFile,PrivateKeyFile
ssl_certificate,prod,example.com,/path/to/cert.crt,/path/to/key.key
```

**JSON Document:**
```csv
ResourceType,Environment,Access,FilePath
jsondoc,dev,app-config,/path/to/config.json
```

**Text File:**
```csv
ResourceType,Environment,Access,FilePath
text_file,prod,api-key,/path/to/key.txt
```

**Features:**
- Auto-detects type from ResourceType column
- Creates or updates secrets automatically
- Tags secrets with metadata
- `-overwrite` flag to update existing secrets

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

### Batch Operations

```bash
# Create CSV with many secrets
cat > secrets.csv <<EOF
ResourceType,Environment,Instance,Database,Access,Password,Engine,Port,DbInstanceIdentifier,Host,Username
rdspostgres,dev,db1,app1,readonly,pass1,postgres,5432,db1-dev,db1.aws.com,ro_user
rdspostgres,dev,db1,app1,readwrite,pass2,postgres,5432,db1-dev,db1.aws.com,rw_user
rdspostgres,prod,db2,app2,readonly,pass3,postgres,5432,db2-prod,db2.aws.com,ro_user
EOF

# Upload all at once
sh-upload -file=secrets.csv -overwrite
```

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
go build -o bin/sh-upload ./cmd/sh-upload
go build -o bin/sh-pull ./cmd/sh-pull
go build -o bin/sh-push ./cmd/sh-push
go build -o bin/sh-generate ./cmd/sh-generate

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
│   ├── sh-upload/         # Batch CSV upload
│   ├── sh-pull/           # Interactive/flag download
│   ├── sh-push/           # Upload with diff
│   └── sh-generate/       # Scaffolding generator
├── jsondoc/               # JSON document secret type
├── rdspostgres/           # RDS PostgreSQL secret type
├── snowflake/             # Snowflake secret type
├── sslcert/               # SSL certificate secret type
├── textfile/              # Text file secret type
├── pull/                  # Pull logic
├── push/                  # Push logic
├── generate/              # Generation logic
├── uploader/              # CSV upload logic
├── secretlogic/           # Pure business logic (testable)
└── tools/                 # Shared utilities
```

---

## Examples

See the `examples/` directory for sample CSV files for each secret type.

---

## Migration from Old Commands

If you were using the old type-specific commands, here's the migration:

| Old Command | New Command |
|-------------|-------------|
| `sh-jsondoc -file=...` | `sh-upload -file=...` |
| `sh-rdsinstance -file=...` | `sh-upload -file=...` |
| `sh-snowflake -file=...` | `sh-upload -file=...` |
| `sh-sslcert -file=...` | `sh-upload -file=...` |
| `sh-textfile -file=...` | `sh-upload -file=...` |
| `sh-download -id=... -file=...` | `sh-pull -type=... -env=...` |

All type-specific CSV upload commands have been replaced by the unified `sh-upload` command. The old `sh-download` command has been replaced by the more capable `sh-pull` command.

---

## Contributing

See [CLEANUP.md](CLEANUP.md) for planned improvements and contribution opportunities.

## License

[Add your license here]
