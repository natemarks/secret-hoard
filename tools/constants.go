package tools

import "time"

// File permissions
const (
	// DefaultFilePermissions is the permission mode for regular files (0644 = rw-r--r--)
	DefaultFilePermissions = 0644
	// DefaultDirPermissions is the permission mode for directories (0755 = rwxr-xr-x)
	DefaultDirPermissions = 0755
)

// AWS and Testing
const (
	// AWSSecretDeletionWaitTime is how long to wait for AWS secret deletion in tests
	AWSSecretDeletionWaitTime = 30 * time.Second
)

// Template placeholders
const (
	// PlaceholderValue is used in generated template files to indicate values that must be replaced
	PlaceholderValue = "REPLACE-ME"
)

// Confirmation
const (
	// ConfirmationStringLength is the length of random confirmation codes
	ConfirmationStringLength = 4
)
