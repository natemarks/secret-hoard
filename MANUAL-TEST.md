# Manual Testing Guide

This document provides step-by-step manual tests for the complete workflow with each secret type. All files are managed in `$HOME/.secret-hoard/` by default.

## Important: Auto-Create Feature

**`sh-push` now automatically creates secrets if they don't exist!**

You no longer need to manually create secrets in AWS Console or use `sh-upload` first. The workflow is now:

1. **`sh-generate`** - Generate scaffolding files
2. **Edit files** - Fill in values
3. **`sh-push`** - Creates secret with proper tags automatically (or updates if exists)

The push command will:
- Check if the secret exists in AWS
- If it doesn't exist: show contents, prompt for confirmation, create with proper tags
- If it exists: show diff, prompt for confirmation, update

This applies to all five secret types (jsondoc, text_file, ssl_certificate, rdspostgres, snowflake).

## Prerequisites

1. Build the executables:
```bash
make build
# Or build individually for testing
go build -o /tmp/sh-generate ./cmd/sh-generate
go build -o /tmp/sh-pull ./cmd/sh-pull
go build -o /tmp/sh-push ./cmd/sh-push
```

2. Ensure AWS credentials are configured:
```bash
aws sts get-caller-identity
```

3. Verify working directory:
```bash
# Directory will be created automatically on first run
ls -la ~/.secret-hoard/
```

---

## Test 1: JSONDOC Secret - Complete Workflow

### Generate Scaffolding
```bash
# Run generate command
sh-generate

# Interactive prompts:
# Select secret type: 4 (jsondoc)
# Environment: test
# Access: manual-test-app
```

**Expected Output:**
```
Generated files:
  - /home/<user>/.secret-hoard/jsondoc.test.manual-test-app.metadata.json
  - /home/<user>/.secret-hoard/jsondoc.test.manual-test-app.contents.json
```

**Verify Files Created:**
```bash
ls -la ~/.secret-hoard/jsondoc.test.manual-test-app.*
cat ~/.secret-hoard/jsondoc.test.manual-test-app.metadata.json
cat ~/.secret-hoard/jsondoc.test.manual-test-app.contents.json
```

### Edit Contents
```bash
# Edit the contents file with actual JSON
cat > ~/.secret-hoard/jsondoc.test.manual-test-app.contents.json <<'EOF'
{
  "app_name": "manual-test-app",
  "version": "1.0.0",
  "config": {
    "timeout": 30,
    "retry": 3,
    "endpoint": "https://example.com/api"
  }
}
EOF
```

### Push to AWS (Create or Update Secret)
```bash
sh-push -metadata=jsondoc.test.manual-test-app.metadata.json -debug
```

**Expected Behavior - If Secret Doesn't Exist (First Time):**
```
Secret does not exist. Creating new secret: jsondoc/test/manual-test-app

New secret contents:
{
  "JSONContents": "{\"app_name\":\"manual-test-app\",\"version\":\"1.0.0\"...}",
  "JSONSha256Sum": "abc123..."
}

Type 'Xr4p' to confirm: 
```

Type the confirmation string. The secret will be created with proper tags automatically.

**Expected Output After Confirmation:**
```
Secret created successfully: jsondoc/test/manual-test-app
```

**Expected Behavior - If Secret Already Exists (Update):**
```
Comparing local files to remote secret: jsondoc/test/manual-test-app

{
  "JSONContents": "...",
- "JSONSha256Sum": "old_hash...",
+ "JSONSha256Sum": "new_hash..."
}

Type 'Xr4p' to confirm overwrite:
```

Type the confirmation string to proceed with update.

### Pull Secret Back

**Option 1: Interactive Mode (Recommended)**
```bash
# Clean local files first
rm ~/.secret-hoard/jsondoc.test.manual-test-app.*

# Pull from AWS interactively
sh-pull -debug

# Interactive prompts:
# Select secret type: 4 (jsondoc)
# Environment: test
# Access: manual-test-app
```

**Option 2: Flag Mode (For Scripting)**
```bash
# Clean local files first
rm ~/.secret-hoard/jsondoc.test.manual-test-app.*

# Pull from AWS with flags
sh-pull -type=jsondoc -env=test -access=manual-test-app -debug
```

**Expected Output:**
```
Successfully pulled secret: jsondoc/test/manual-test-app
  - /home/<user>/.secret-hoard/jsondoc.test.manual-test-app.metadata.json
  - /home/<user>/.secret-hoard/jsondoc.test.manual-test-app.contents.json
```

**Verify Contents Match:**
```bash
cat ~/.secret-hoard/jsondoc.test.manual-test-app.contents.json
# Should match what you pushed
```

### Modify and Re-push
```bash
# Edit contents
cat > ~/.secret-hoard/jsondoc.test.manual-test-app.contents.json <<'EOF'
{
  "app_name": "manual-test-app",
  "version": "1.0.1",
  "config": {
    "timeout": 60,
    "retry": 5,
    "endpoint": "https://example.com/api/v2"
  }
}
EOF

# Push update
sh-push -metadata=jsondoc.test.manual-test-app.metadata.json
```

**Expected Behavior:**
- Shows diff with changes (- old values, + new values)
- Prompts for confirmation
- Updates secret on confirmation

### Cleanup
```bash
# Delete the test secret from AWS
aws secretsmanager delete-secret \
  --secret-id jsondoc/test/manual-test-app \
  --force-delete-without-recovery

# Remove local files
rm ~/.secret-hoard/jsondoc.test.manual-test-app.*
```

---

## Test 2: TEXT_FILE Secret - Complete Workflow

### Generate Scaffolding
```bash
sh-generate

# Interactive prompts:
# Select secret type: 5 (text_file)
# Environment: test
# Access: config-file
```

**Verify Files:**
```bash
ls -la ~/.secret-hoard/textfile.test.config-file.*
```

### Edit Contents
```bash
# Add text content
cat > ~/.secret-hoard/textfile.test.config-file.contents.txt <<'EOF'
# Application Configuration
SERVER_HOST=localhost
SERVER_PORT=8080
DATABASE_URL=postgresql://localhost:5432/mydb
LOG_LEVEL=info
MAX_CONNECTIONS=100
EOF
```

### Push to AWS (Create Secret)
```bash
# Push to create the secret (auto-creates with proper tags)
sh-push -metadata=textfile.test.config-file.metadata.json
```

**Expected Output:**
```
Secret does not exist. Creating new secret: text_file/test/config-file

New secret contents:
{
  "Contents": "# Application Configuration\nSERVER_HOST=localhost...",
  "Sha256Sum": "abc123..."
}

Type 'Xr4p' to confirm: 
```

### Push Update (After Modification)
```bash
# Modify the contents
echo "# Updated configuration" | cat - ~/.secret-hoard/textfile.test.config-file.contents.txt > /tmp/temp && mv /tmp/temp ~/.secret-hoard/textfile.test.config-file.contents.txt

# Push changes (shows diff and updates)
sh-push -metadata=textfile.test.config-file.metadata.json
```

### Pull and Verify
```bash
rm ~/.secret-hoard/textfile.test.config-file.*
sh-pull -type=text_file -env=test -access=config-file
cat ~/.secret-hoard/textfile.test.config-file.contents.txt
```

### Cleanup
```bash
aws secretsmanager delete-secret \
  --secret-id text_file/test/config-file \
  --force-delete-without-recovery
rm ~/.secret-hoard/textfile.test.config-file.*
```

---

## Test 3: SSL_CERTIFICATE Secret - Complete Workflow

### Generate Scaffolding
```bash
sh-generate

# Interactive prompts:
# Select secret type: 3 (ssl_certificate)
# Environment: test
# Common Name: test.example.com
```

**Verify Files:**
```bash
ls -la ~/.secret-hoard/sslcert.test.test.example.com.*
```

### Generate Test Certificate and Key
```bash
# Generate self-signed certificate for testing
cd ~/.secret-hoard

openssl req -x509 -newkey rsa:2048 -nodes \
  -keyout sslcert.test.test.example.com.key \
  -out sslcert.test.test.example.com.crt \
  -days 365 \
  -subj "/CN=test.example.com"
```

### Push to AWS (Create Secret)
```bash
# Push to create the secret (auto-creates with proper tags)
sh-push -metadata=sslcert.test.test.example.com.metadata.json
```

**Expected Output:**
```
Secret does not exist. Creating new secret: ssl_certificate/test/test.example.com

New secret contents:
{
  "Certificate": "-----BEGIN CERTIFICATE-----\n...",
  "PrivateKey": "-----BEGIN PRIVATE KEY-----\n...",
  "ExpirationDate": "2027-06-24T...",
  "Modulus": "ABC123...",
  "CertificateSha256": "def456...",
  "PrivateKeySha256": "ghi789..."
}

Type 'Xr4p' to confirm: 
```

### Pull Secret Back to Verify
```bash
# Remove local files first
rm ~/.secret-hoard/sslcert.test.test.example.com.*

# Pull from AWS
sh-pull -type=ssl_certificate -env=test -commonname=test.example.com -debug
```

**Verify Files:**
```bash
ls -la ~/.secret-hoard/sslcert.test.test.example.com.*
cat ~/.secret-hoard/sslcert.test.test.example.com.metadata.json
# Should contain expiration, modulus, and SHA256 sums

# Verify certificate
openssl x509 -in ~/.secret-hoard/sslcert.test.test.example.com.crt -text -noout | head -20
```

### Generate New Certificate and Push Update
```bash
# Generate new certificate
openssl req -x509 -newkey rsa:2048 -nodes \
  -keyout ~/.secret-hoard/sslcert.test.test.example.com.key \
  -out ~/.secret-hoard/sslcert.test.test.example.com.crt \
  -days 730 \
  -subj "/CN=test.example.com/O=Test Org/C=US"

# Push update
sh-push -metadata=sslcert.test.test.example.com.metadata.json
```

**Expected Behavior:**
- Shows diff with certificate and key changes
- Updates expiration date
- Prompts for confirmation

### Cleanup
```bash
aws secretsmanager delete-secret \
  --secret-id ssl_certificate/test/test.example.com \
  --force-delete-without-recovery
rm ~/.secret-hoard/sslcert.test.test.example.com.*
```

---

## Test 4: RDSPOSTGRES Secret - Complete Workflow

### Generate Scaffolding
```bash
sh-generate

# Interactive prompts:
# Select secret type: 1 (rdspostgres)
# Environment: test
# Instance: test-db
# Database: myapp
# Access: readonly
```

**Verify File:**
```bash
ls -la ~/.secret-hoard/rdspostgres.test.test-db.myapp.readonly.json
cat ~/.secret-hoard/rdspostgres.test.test-db.myapp.readonly.json
```

### Edit Configuration
```bash
# Replace REPLACE-ME values
cat > ~/.secret-hoard/rdspostgres.test.test-db.myapp.readonly.json <<'EOF'
{
  "metadata": {
    "resourceType": "rdspostgres",
    "environment": "test",
    "instance": "test-db",
    "database": "myapp",
    "access": "readonly"
  },
  "data": {
    "password": "test_readonly_pass_12345",
    "engine": "postgres",
    "port": 5432,
    "dbInstanceIdentifier": "test-db-instance",
    "host": "test-db.123456789012.us-east-1.rds.amazonaws.com",
    "username": "readonly_user"
  }
}
EOF
```

### Push to AWS (Create Secret)
```bash
# Push to create the secret (auto-creates with proper tags)
sh-push -metadata=rdspostgres.test.test-db.myapp.readonly.json
```

**Expected Output:**
```
Secret does not exist. Creating new secret: rdspostgres/test/test-db/myapp/readonly

New secret contents:
{
  "password": "test_readonly_pass_12345",
  "engine": "postgres",
  "port": 5432,
  "dbInstanceIdentifier": "test-db-instance",
  "host": "test-db.123456789012.us-east-1.rds.amazonaws.com",
  "username": "readonly_user"
}

Type 'Xr4p' to confirm: 
```

### Pull Secret Back to Verify
```bash
rm ~/.secret-hoard/rdspostgres.test.test-db.myapp.readonly.json
sh-pull -type=rdspostgres -env=test -instance=test-db -database=myapp -access=readonly
cat ~/.secret-hoard/rdspostgres.test.test-db.myapp.readonly.json | jq .
```

### Update and Push
```bash
# Update password and host
jq '.data.password = "new_test_password_67890" | .data.host = "test-db-new.123456789012.us-east-1.rds.amazonaws.com"' \
  ~/.secret-hoard/rdspostgres.test.test-db.myapp.readonly.json > /tmp/temp.json && \
  mv /tmp/temp.json ~/.secret-hoard/rdspostgres.test.test-db.myapp.readonly.json

# Push update
sh-push -metadata=rdspostgres.test.test-db.myapp.readonly.json
```

### Cleanup
```bash
aws secretsmanager delete-secret \
  --secret-id rdspostgres/test/test-db/myapp/readonly \
  --force-delete-without-recovery
rm ~/.secret-hoard/rdspostgres.test.test-db.myapp.readonly.json
```

---

## Test 5: SNOWFLAKE Secret - Complete Workflow

### Generate Scaffolding
```bash
sh-generate

# Interactive prompts:
# Select secret type: 2 (snowflake)
# Environment: test
# Warehouse: analytics
# Access: developer
```

**Verify File:**
```bash
ls -la ~/.secret-hoard/snowflake.test.analytics.developer.json
cat ~/.secret-hoard/snowflake.test.analytics.developer.json
```

### Edit Configuration
```bash
cat > ~/.secret-hoard/snowflake.test.analytics.developer.json <<'EOF'
{
  "metadata": {
    "resourceType": "snowflake",
    "environment": "test",
    "warehouse": "analytics",
    "access": "developer"
  },
  "data": {
    "password": "snowflake_dev_password_123",
    "accountName": "mycompany.us-east-1",
    "warehouse": "analytics",
    "username": "dev_user"
  }
}
EOF
```

### Push to AWS (Create Secret)
```bash
# Push to create the secret (auto-creates with proper tags)
sh-push -metadata=snowflake.test.analytics.developer.json
```

**Expected Output:**
```
Secret does not exist. Creating new secret: snowflake/test/analytics/developer

New secret contents:
{
  "password": "snowflake_dev_password_123",
  "accountName": "mycompany.us-east-1",
  "warehouse": "analytics",
  "username": "dev_user"
}

Type 'Xr4p' to confirm: 
```

### Pull Secret Back to Verify
```bash
rm ~/.secret-hoard/snowflake.test.analytics.developer.json
sh-pull -type=snowflake -env=test -warehouse=analytics -access=developer
cat ~/.secret-hoard/snowflake.test.analytics.developer.json | jq .
```

### Update and Push
```bash
# Update password
jq '.data.password = "new_snowflake_password_456"' \
  ~/.secret-hoard/snowflake.test.analytics.developer.json > /tmp/temp.json && \
  mv /tmp/temp.json ~/.secret-hoard/snowflake.test.analytics.developer.json

# Push update
sh-push -metadata=snowflake.test.analytics.developer.json
```

### Cleanup
```bash
aws secretsmanager delete-secret \
  --secret-id snowflake/test/analytics/developer \
  --force-delete-without-recovery
rm ~/.secret-hoard/snowflake.test.analytics.developer.json
```

---

## Test 6: Interactive Mode for sh-pull

### Test Interactive Prompts for Each Type

**JSONDOC Interactive:**
```bash
sh-pull -debug

# Prompts:
# Select secret type: 4
# Environment: test
# Access: myapp

# Expected: Downloads jsondoc/test/myapp to ~/.secret-hoard/
```

**TEXT_FILE Interactive:**
```bash
sh-pull

# Prompts:
# Select secret type: 5
# Environment: prod
# Access: config

# Expected: Downloads text_file/prod/config to ~/.secret-hoard/
```

**SSL_CERTIFICATE Interactive:**
```bash
sh-pull

# Prompts:
# Select secret type: 3
# Environment: staging
# Common Name: example.com

# Expected: Downloads ssl_certificate/staging/example.com to ~/.secret-hoard/
```

**RDSPOSTGRES Interactive:**
```bash
sh-pull

# Prompts:
# Select secret type: 1
# Environment: dev
# Instance: mydb
# Database: app
# Access: readonly

# Expected: Downloads rdspostgres/dev/mydb/app/readonly to ~/.secret-hoard/
```

**SNOWFLAKE Interactive:**
```bash
sh-pull

# Prompts:
# Select secret type: 2
# Environment: prod
# Warehouse: analytics
# Access: developer

# Expected: Downloads snowflake/prod/analytics/developer to ~/.secret-hoard/
```

### Verify Prompts Match Requirements

Each type should only prompt for required metadata:
- **jsondoc**: environment, access (2 fields)
- **text_file**: environment, access (2 fields)
- **ssl_certificate**: environment, commonname (2 fields)
- **snowflake**: environment, warehouse, access (3 fields)
- **rdspostgres**: environment, instance, database, access (4 fields)

---

## Integration Test: Complete Multi-Type Workflow

This test validates the entire workflow across multiple secret types in sequence.

### Setup
```bash
# Clean working directory
rm -rf ~/.secret-hoard/*

# Verify working directory gets created
ls -la ~/.secret-hoard/ || echo "Directory will be created on first run"
```

### Test Sequence

1. **Generate all types:**
```bash
# Generate jsondoc
sh-generate # Select 4, env: integration, access: app1

# Generate textfile
sh-generate # Select 5, env: integration, access: config1

# Generate rdspostgres
sh-generate # Select 1, env: integration, instance: db1, database: app, access: ro

# Generate snowflake
sh-generate # Select 2, env: integration, warehouse: wh1, access: dev

# Verify all scaffolding created
ls -la ~/.secret-hoard/
```

2. **Edit all files with test data** (use examples from individual tests above)

3. **Create secrets in AWS** (use CSV upload method for initial creation)

4. **Pull all secrets back:**
```bash
rm ~/.secret-hoard/*
sh-pull -type=jsondoc -env=integration -access=app1
sh-pull -type=text_file -env=integration -access=config1
sh-pull -type=rdspostgres -env=integration -instance=db1 -database=app -access=ro
sh-pull -type=snowflake -env=integration -warehouse=wh1 -access=dev

# Verify all files present
ls -la ~/.secret-hoard/
```

5. **Modify each and push updates:**
```bash
# Modify each file, then push
sh-push -metadata=jsondoc.integration.app1.metadata.json
sh-push -metadata=textfile.integration.config1.metadata.json
sh-push -metadata=rdspostgres.integration.db1.app.ro.json
sh-push -metadata=snowflake.integration.wh1.dev.json
```

6. **Pull again to verify:**
```bash
rm ~/.secret-hoard/*
# Pull all again and verify contents match what was pushed
```

7. **Cleanup:**
```bash
# Delete all test secrets
aws secretsmanager delete-secret --secret-id jsondoc/integration/app1 --force-delete-without-recovery
aws secretsmanager delete-secret --secret-id text_file/integration/config1 --force-delete-without-recovery
aws secretsmanager delete-secret --secret-id rdspostgres/integration/db1/app/ro --force-delete-without-recovery
aws secretsmanager delete-secret --secret-id snowflake/integration/wh1/dev --force-delete-without-recovery

rm -rf ~/.secret-hoard/*
```

---

## Validation Checklist

For each test, verify:

- [ ] Working directory `~/.secret-hoard/` is created automatically
- [ ] Generate creates correct file structure with proper naming
- [ ] Generated files contain correct metadata and REPLACE-ME placeholders
- [ ] Push shows diff comparing local vs remote
- [ ] Push prompts with random 4-character confirmation string
- [ ] Push only proceeds when correct string is entered
- [ ] Push aborts when wrong string or no input
- [ ] Pull creates files in `~/.secret-hoard/`
- [ ] Pull validates SHA256 for content-based types (jsondoc, textfile, sslcert)
- [ ] Pulled files match what was pushed
- [ ] File paths are always in `~/.secret-hoard/` by default
- [ ] Debug flag shows additional logging
- [ ] Error messages are clear and actionable

---

## Error Condition Tests

### Test Invalid Secret Type
```bash
sh-pull -type=invalid -env=test -access=test
# Expected: Error message about unknown secret type
```

### Test Missing Required Flags
```bash
sh-pull -type=jsondoc -env=test
# Expected: Error about missing -access flag

sh-pull -type=rdspostgres -env=test -instance=db1
# Expected: Error about missing -database and -access flags
```

### Test Non-Existent Secret
```bash
sh-pull -type=jsondoc -env=nonexistent -access=nonexistent
# Expected: Clear error about secret not found
```

### Test Push Without Metadata File
```bash
sh-push -metadata=nonexistent.json
# Expected: Error about file not existing
```

### Test SHA256 Mismatch
```bash
# Pull a jsondoc secret
sh-pull -type=jsondoc -env=test -access=test

# Manually corrupt the metadata SHA256
sed -i 's/"JSONSha256Sum": "[^"]*"/"JSONSha256Sum": "invalid_hash"/' \
  ~/.secret-hoard/jsondoc.test.test.metadata.json

# Try to push
sh-push -metadata=jsondoc.test.test.metadata.json
# Expected: Should still work but computed hash won't match metadata value
```

### Test Certificate/Key Mismatch (sslcert)
```bash
# Generate mismatched cert and key
openssl req -x509 -newkey rsa:2048 -nodes \
  -keyout ~/.secret-hoard/test.key \
  -out ~/.secret-hoard/test.crt \
  -days 365 -subj "/CN=test.com"

# Generate different key
openssl genrsa -out ~/.secret-hoard/test2.key 2048

# Try to use mismatched pair
mv ~/.secret-hoard/test2.key ~/.secret-hoard/sslcert.test.test.com.key
mv ~/.secret-hoard/test.crt ~/.secret-hoard/sslcert.test.test.com.crt

sh-push -metadata=sslcert.test.test.com.metadata.json
# Expected: Error about moduli not matching
```

---

## Performance Test

Test with multiple secrets to verify directory performance:

```bash
# Generate 20 jsondoc secrets
for i in {1..20}; do
  echo "Creating secret $i..."
  # Use sh-generate or create files programmatically
done

# Pull all 20
# Measure time
time for i in {1..20}; do
  sh-pull -type=jsondoc -env=test -access=app$i
done

# Verify all files in working directory
ls ~/.secret-hoard/ | wc -l
# Expected: 40 files (metadata + contents for each)
```

---

## Notes

1. **Working Directory**: All operations default to `$HOME/.secret-hoard/`. The directory is created automatically on first use.

2. **Path Resolution**: When using `-metadata` flag in `sh-push`, relative paths are resolved relative to `~/.secret-hoard/`.

3. **Initial Creation**: These tools require secrets to exist in AWS first (create via CSV upload or AWS console). They are designed for pull/edit/push workflows, not initial creation.

4. **Confirmation String**: The random 4-character string is case-sensitive and must match exactly.

5. **Debug Mode**: Use `-debug` flag to see detailed logging including working directory path and AWS operations.
