package secrets

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
)

// AWSSecretsManager implements SecretsManager using AWS SDK
type AWSSecretsManager struct {
	client *secretsmanager.Client
}

// NewAWSSecretsManager creates a new AWS Secrets Manager client
func NewAWSSecretsManager() (*AWSSecretsManager, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	return &AWSSecretsManager{
		client: secretsmanager.NewFromConfig(cfg),
	}, nil
}

// DescribeSecret checks if a secret exists
func (sm *AWSSecretsManager) DescribeSecret(ctx context.Context, secretID string) (bool, error) {
	_, err := sm.client.DescribeSecret(ctx, &secretsmanager.DescribeSecretInput{
		SecretId: aws.String(secretID),
	})
	if err != nil {
		var notFound *types.ResourceNotFoundException
		if err.Error() == notFound.Error() || err.Error() == "ResourceNotFoundException" {
			return false, nil
		}
		return false, fmt.Errorf("error describing secret: %w", err)
	}
	return true, nil
}

// CreateSecret creates a new secret with tags
func (sm *AWSSecretsManager) CreateSecret(ctx context.Context, secretID string, value interface{}, tags map[string]string) error {
	// Convert value to JSON string
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("error marshaling secret value: %w", err)
	}

	// Convert tags map to AWS tags format
	awsTags := make([]types.Tag, 0, len(tags))
	for k, v := range tags {
		awsTags = append(awsTags, types.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}

	// Add Source tag
	awsTags = append(awsTags, types.Tag{
		Key:   aws.String("Source"),
		Value: aws.String("secret-hoard"),
	})

	_, err = sm.client.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
		Name:         aws.String(secretID),
		SecretString: aws.String(string(valueBytes)),
		Tags:         awsTags,
	})
	if err != nil {
		return fmt.Errorf("error creating secret: %w", err)
	}

	return nil
}

// UpdateSecret updates an existing secret
func (sm *AWSSecretsManager) UpdateSecret(ctx context.Context, secretID string, value interface{}) error {
	// Convert value to JSON string
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("error marshaling secret value: %w", err)
	}

	_, err = sm.client.UpdateSecret(ctx, &secretsmanager.UpdateSecretInput{
		SecretId:     aws.String(secretID),
		SecretString: aws.String(string(valueBytes)),
	})
	if err != nil {
		return fmt.Errorf("error updating secret: %w", err)
	}

	return nil
}

// GetSecretValue retrieves a secret's value
func (sm *AWSSecretsManager) GetSecretValue(ctx context.Context, secretID string) (string, error) {
	result, err := sm.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretID),
	})
	if err != nil {
		return "", fmt.Errorf("error getting secret value: %w", err)
	}

	if result.SecretString == nil {
		return "", fmt.Errorf("secret has no string value")
	}

	return *result.SecretString, nil
}
