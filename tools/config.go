package tools

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// LoadAWSConfig loads AWS configuration with region fallback logic:
// 1. Uses current session credentials (never specifies profile)
// 2. Gets region from environment variables (AWS_REGION, AWS_DEFAULT_REGION)
// 3. Falls back to us-east-1 if no region is configured
func LoadAWSConfig(ctx context.Context) (aws.Config, error) {
	// Load default config (uses current session, respects env vars)
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return aws.Config{}, err
	}

	// If no region was detected, check env vars explicitly and fallback to us-east-1
	if cfg.Region == "" {
		if region := os.Getenv("AWS_REGION"); region != "" {
			cfg.Region = region
		} else if region := os.Getenv("AWS_DEFAULT_REGION"); region != "" {
			cfg.Region = region
		} else {
			cfg.Region = "us-east-1"
		}
	}

	return cfg, nil
}
