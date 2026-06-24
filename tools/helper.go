package tools

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// GetSecretValue retrieves the value of a secret
func GetSecretValue(secretID string) (string, error) {
	// Load AWS SDK configuration
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return "", err
	}

	// Create Secrets Manager client
	client := secretsmanager.NewFromConfig(cfg)

	// Prepare input parameters
	input := &secretsmanager.GetSecretValueInput{
		SecretId: &secretID,
	}

	// Retrieve secret value
	result, err := client.GetSecretValue(context.TODO(), input)
	if err != nil {
		return "", err
	}

	// Check if secret value is present
	if result.SecretString == nil {
		return "", fmt.Errorf("secret value is nil")
	}

	// Return the secret value
	return *result.SecretString, nil
}
