# Zerolog Removal - Simplification Summary

## Overview

Replaced zerolog structured JSON logging with standard Go logging throughout the project. For interactive CLI tools, structured JSON logging is unnecessary overhead.

## What Changed

### Created New Simple Logger

**tools/logger.go** - New simple logger implementation:
```go
type Logger struct {
    debug   *log.Logger  // DEBUG: prefix
    info    *log.Logger  // No prefix
    error   *log.Logger  // ERROR: prefix  
}

func NewLogger(debugMode bool) *Logger
func (l *Logger) Debug(format string, v ...any)
func (l *Logger) Info(format string, v ...any)
func (l *Logger) Error(format string, v ...any)
func (l *Logger) Fatal(format string, v ...any)
```

### Files Modified

**Removed zerolog imports and updated logger usage in:**

#### Commands (4 files)
- cmd/sh-pull/config.go - GetLogger() now returns *tools.Logger
- cmd/sh-push/config.go - GetLogger() now returns *tools.Logger  
- cmd/sh-generate/config.go - GetLogger() now returns *tools.Logger
- cmd/sh-upload/main.go - Updated GetCSVProcessor() signature

#### Business Logic (3 files)
- pull/pull.go - All functions use *tools.Logger
- push/push.go - All functions use *tools.Logger
- generate/generate.go - All functions use *tools.Logger

#### Secret Types (10 files)
- jsondoc/jsondoc.go & jsondoc/jsondoccsv.go
- textfile/textfile.go & textfile/textfilecsv.go
- sslcert/sslcert.go & sslcert/sslcertcsv.go
- rdspostgres/rdspostgres.go & rdspostgres/rdspostgrescsv.go
- snowflake/snowflake.go & snowflake/snowflakecsv.go

#### Uploaders (6 files)
- uploader/processor.go - Interface updated
- uploader/jsondoc.go
- uploader/textfile.go
- uploader/sslcert.go
- uploader/rdspostgres.go
- uploader/snowflake.go

#### Core Infrastructure (4 files)
- tools/helper.go - Removed TestLogger() and SimpleLogger()
- tools/config.go - GetLogger() now returns *tools.Logger
- secrets/interfaces.go - Secret interface uses *tools.Logger
- secrets/operations.go - Generic operations use *tools.Logger

#### Main Entry Points (4 files)
- cmd/sh-pull/main.go
- cmd/sh-push/main.go
- cmd/sh-generate/main.go
- cmd/sh-upload/main.go

### Log Call Transformations

**Zerolog fluent API → Simple printf-style:**

```go
// Before
log.Info().Msgf("Secret ID: %s", secretID)
log.Debug().Msgf("Using directory: %s", dir)
log.Error().Err(err).Msg("Failed to create secret")
log.Fatal().Err(err).Msgf("Cannot load config: %v", err)
*log = log.With().Str("env", env).Logger()  // Context addition

// After
log.Info("Secret ID: %s", secretID)
log.Debug("Using directory: %s", dir)
log.Error("Failed to create secret: %v", err)
log.Fatal("Cannot load config: %v", err)
// Removed - no context addition needed for simple CLI tools
```

### Removed Functions

From tools/helper.go:
- `TestLogger()` - Created structured logger with AWS account number
- `SimpleLogger()` - Created structured logger without AWS account

These are replaced by a single `tools.NewLogger(debugMode bool)` function.

## Benefits

### 1. Simplicity
- **Before**: `log.Info().Msgf("message: %s", value)` (fluent API, chaining)
- **After**: `log.Info("message: %s", value)` (familiar printf-style)
- Developers already know printf-style formatting

### 2. Reduced Dependencies
- Removed zerolog dependency from go.mod
- One less external package to maintain
- Faster builds

### 3. Appropriate for Use Case
- Interactive CLI tools don't need JSON logging
- Users read terminal output directly
- Structured logs are for machine parsing (not needed here)

### 4. Better Performance
- No JSON marshaling overhead
- No reflection for field extraction
- Direct writes to stdout/stderr

### 5. Code Size Reduction
- Removed complex logger initialization code
- Removed AWS account number lookup for logger context
- Simpler configuration

## Testing

✅ All 4 commands build successfully:
- sh-pull
- sh-push
- sh-generate  
- sh-upload

✅ All existing tests pass (55 tests in secretlogic/)

✅ No breaking changes to command-line interfaces

## Output Comparison

### Before (zerolog JSON)
```json
{"level":"info","version":"1.0.0","aws_account_number":"123456789","time":"2026-01-15T10:30:00Z","message":"Secret created: jsondoc/dev/app"}
```

### After (simple logging)
```
Secret created: jsondoc/dev/app
```

For debug mode:
```
DEBUG: 2026/01/15 10:30:00 config: {MetadataFile:jsondoc.dev.app.metadata.json Debug:true}
```

Much more readable for interactive CLI usage!

## Migration Notes

The logger interface changed slightly:

| Old (zerolog) | New (tools.Logger) |
|--------------|-------------------|
| `log.Info().Msg("text")` | `log.Info("text")` |
| `log.Info().Msgf("text: %s", v)` | `log.Info("text: %s", v)` |
| `log.Error().Err(err).Msg("text")` | `log.Error("text: %v", err)` |
| `log.With().Str("k", v).Logger()` | *(removed - not needed)* |
| `tools.TestLogger()` | `tools.NewLogger(debug)` |
| `tools.SimpleLogger()` | `tools.NewLogger(debug)` |

## Future Considerations

If structured logging becomes necessary later (e.g., for log aggregation), we can:
1. Keep the same `*tools.Logger` interface
2. Add a backend flag to switch between simple and JSON output
3. The calling code wouldn't need to change

## Files Count

**Total files modified: 31**
- 4 command configs
- 4 command main files  
- 3 business logic packages
- 10 secret type files
- 6 uploader files
- 4 core infrastructure files

**Lines removed: ~100 lines of logging boilerplate**
- Complex logger initialization
- AWS account number lookup
- Context addition with .With()
- Fluent API chaining

**Lines added: ~60 lines**
- tools/logger.go implementation
- Simple, focused logger

**Net reduction: ~40 lines**

## Conclusion

Replaced structured JSON logging with simple printf-style logging appropriate for interactive CLI tools. The codebase is now simpler, has fewer dependencies, and is more maintainable while providing better user experience through readable terminal output.
