# Quick Start Guide: Secret Synchronization Tools

## Overview

Three new tools for interactive secret management:
- `sh-generate` - Create local file templates
- `sh-pull` - Download secrets to local files
- `sh-push` - Upload local files to AWS (with safety checks)

**Working Directory**: All files are managed in `$HOME/.secret-hoard/` by default. This directory is created automatically on first use.

## Common Workflows

### Workflow 1: Create a New Secret

```bash
# Step 1: Generate template files
sh-generate
# Select: 4 (jsondoc)
# Environment: dev
# Access: myapp
# Files created in ~/.secret-hoard/

# Step 2: Edit the generated files
vim ~/.secret-hoard/jsondoc.dev.myapp.contents.json

# Step 3: Push to AWS (creates new secret)
sh-push -metadata=jsondoc.dev.myapp.metadata.json
# Review diff, type confirmation string
```

### Workflow 2: Edit an Existing Secret

```bash
# Step 1: Pull current version (interactive)
sh-pull
# Select: 4 (jsondoc)
# Environment: dev
# Access: myapp
# Downloads to ~/.secret-hoard/

# Step 2: Edit local files
vim ~/.secret-hoard/jsondoc.dev.myapp.contents.json

# Step 3: Push changes
sh-push -metadata=jsondoc.dev.myapp.metadata.json
# Review diff, confirm changes
```

### Workflow 3: Clone Secret to Another Environment

```bash
# Step 1: Pull from prod
sh-pull -type=jsondoc -env=prod -access=myapp

# Step 2: Rename files for dev
mv jsondoc.prod.myapp.metadata.json jsondoc.dev.myapp.metadata.json
mv jsondoc.prod.myapp.contents.json jsondoc.dev.myapp.contents.json

# Step 3: Update metadata file
sed -i 's/"prod"/"dev"/g' jsondoc.dev.myapp.metadata.json

# Step 4: Push to dev
sh-push -metadata=jsondoc.dev.myapp.metadata.json
```

## Command Reference

### sh-generate

**Purpose**: Interactively create local file scaffolding

**Usage**:
```bash
sh-generate [-debug]
```

**Interactive Prompts**:
1. Select secret type (1-5)
2. Environment
3. Type-specific fields

**Output**: Local files with REPLACE-ME placeholders

---

### sh-pull

**Purpose**: Download a secret and create local editable files

**Interactive Mode** (recommended):
```bash
# Run without flags to be prompted
sh-pull

# Interactive prompts:
# 1. Select secret type (1-5)
# 2. Environment
# 3. Type-specific fields (instance, database, access, etc.)
```

**Flag Mode** (for scripting):
```bash
# JSON document
sh-pull -type=jsondoc -env=dev -access=idemia

# Text file
sh-pull -type=text_file -env=prod -access=config

# SSL certificate
sh-pull -type=ssl_certificate -env=staging -commonname=example.com

# RDS Postgres
sh-pull -type=rdspostgres -env=dev -instance=mydb -database=app -access=readonly

# Snowflake
sh-pull -type=snowflake -env=prod -warehouse=analytics -access=admin
```

**Flags**:
- `-type` - Secret type: rdspostgres, snowflake, ssl_certificate, jsondoc, text_file
- `-env` - Environment
- `-access` - Access type (jsondoc, text_file, snowflake, rdspostgres)
- `-instance` - Instance (rdspostgres)
- `-database` - Database (rdspostgres)
- `-warehouse` - Warehouse (snowflake)
- `-commonname` - Common name (ssl_certificate)
- `-debug` - Enable debug logging

**Note**: If no `-type` flag is provided, interactive mode is used automatically.

---

### sh-push

**Purpose**: Upload local files with diff and confirmation

**Usage**:
```bash
sh-push -metadata=<metadata-file> [-debug]
```

**Examples**:
```bash
# Push jsondoc
sh-push -metadata=jsondoc.dev.idemia.metadata.json

# Push with debug logging
sh-push -metadata=textfile.prod.config.metadata.json -debug
```

**Safety Features**:
- Shows diff between local and remote
- Skips if no changes detected
- Requires typing random 4-character string to confirm

**Flags**:
- `-metadata` - Path to metadata JSON file (required)
- `-debug` - Enable debug logging

---

## File Structures by Type

### jsondoc
```
jsondoc.{env}.{access}.metadata.json  # Metadata + SHA256
jsondoc.{env}.{access}.contents.json  # Actual JSON document
```

### text_file
```
textfile.{env}.{access}.metadata.json  # Metadata + SHA256
textfile.{env}.{access}.contents.txt   # Actual text content
```

### ssl_certificate
```
sslcert.{env}.{commonname}.metadata.json  # All computed fields
sslcert.{env}.{commonname}.crt            # Certificate PEM
sslcert.{env}.{commonname}.key            # Private key PEM
```

### rdspostgres (single file)
```
rdspostgres.{env}.{instance}.{database}.{access}.json
```
Contains:
```json
{
  "metadata": {...},
  "data": {
    "password": "...",
    "engine": "postgres",
    "port": 5432,
    "host": "...",
    "username": "..."
  }
}
```

### snowflake (single file)
```
snowflake.{env}.{warehouse}.{access}.json
```
Contains:
```json
{
  "metadata": {...},
  "data": {
    "password": "...",
    "accountName": "...",
    "warehouse": "...",
    "username": "..."
  }
}
```

---

## Tips & Tricks

### 1. Validate Before Push
```bash
# Check JSON syntax before pushing
jq . jsondoc.dev.myapp.contents.json

# Verify certificate and key match
openssl x509 -noout -modulus -in sslcert.staging.example.com.crt | openssl md5
openssl rsa -noout -modulus -in sslcert.staging.example.com.key | openssl md5
```

### 2. Bulk Operations
```bash
# Pull all dev jsondoc secrets (if you know the names)
for access in app1 app2 app3; do
  sh-pull -type=jsondoc -env=dev -access=$access
done

# Generate multiple secrets
# (Note: sh-generate is interactive, use scripted approach instead)
```

### 3. Backup Before Major Changes
```bash
# Pull all secrets to backup directory
mkdir backup-$(date +%Y%m%d)
cd backup-$(date +%Y%m%d)
# Pull secrets...
```

### 4. Compare Environments
```bash
# Pull from both environments
sh-pull -type=jsondoc -env=dev -access=config
sh-pull -type=jsondoc -env=prod -access=config

# Compare
diff jsondoc.dev.config.contents.json jsondoc.prod.config.contents.json
```

### 5. Version Control Integration
```bash
# Store metadata files in git (without sensitive contents)
git add *.metadata.json
git commit -m "Update secret metadata"

# Add contents files to .gitignore
echo "*.contents.json" >> .gitignore
echo "*.contents.txt" >> .gitignore
echo "*.crt" >> .gitignore
echo "*.key" >> .gitignore
```

---

## Troubleshooting

### "Secret not found" error
- Verify secret exists in AWS Secrets Manager
- Check secret ID format matches: `{type}/{env}/{metadata...}`
- Ensure AWS credentials are configured

### "SHA256 verification failed"
- Contents file was modified after download
- Network corruption during download (rare)
- Re-pull the secret to get fresh copy

### "Certificate and private key moduli do not match"
- Wrong private key for certificate
- Files were modified incorrectly
- Ensure you're using matching cert/key pair

### Push shows "Update cancelled"
- You typed the wrong confirmation string
- Re-run the command and type the exact string shown

### No diff shown, but changes expected
- Local and remote are identical
- Check if you edited the right file
- Verify metadata file points to correct contents file

---

## Best Practices

1. **Always review diffs carefully** before confirming push
2. **Test in dev first** before pushing to prod
3. **Keep metadata files** even if you delete contents (easy to re-pull)
4. **Use descriptive access names** (e.g., `app-readonly` not just `ro`)
5. **Document your secrets** in a separate README (what they're for, who uses them)
6. **Rotate secrets regularly** using this workflow:
   ```bash
   sh-pull -type=... (get current)
   # Update password/credentials
   sh-push -metadata=... (upload new version)
   ```
