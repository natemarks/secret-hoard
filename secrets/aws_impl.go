package secrets

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/natemarks/secret-hoard/tools"
)

// AWSSecretsManager implements Manager using AWS SDK
type AWSSecretsManager struct {
	client *secretsmanager.Client
}

// NewAWSSecretsManager creates a new AWS Secrets Manager client
func NewAWSSecretsManager() (*AWSSecretsManager, error) {
	cfg, err := tools.LoadAWSConfig(context.Background())
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
		// Check if it's a ResourceNotFoundException by type comparison
		// Note: We check the type without calling .Error() on the AWS SDK error
		// because some error implementations may have nil internal fields
		var notFound *types.ResourceNotFoundException
		if errors.As(err, &notFound) {
			// Secret doesn't exist - this is expected, not an error
			return false, nil
		}
		// Some other AWS error occurred - this is a real error
		// DO NOT wrap the error with fmt.Errorf here because the AWS SDK error
		// may have nil internal fields that cause panics when .Error() is called
		return false, err
	}
	return true, nil
}

// CreateSecret creates a new secret with tags
func (sm *AWSSecretsManager) CreateSecret(ctx context.Context, secretID string, value any, tags map[string]string) error {
	// Convert value to JSON string
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("error marshaling secret value: %w", err)
	}

	// Convert tags map to AWS tags format
	// Note: The Source tag should already be included in the tags map
	// by the secret type's Metadata.Map() method
	awsTags := make([]types.Tag, 0, len(tags))
	for k, v := range tags {
		awsTags = append(awsTags, types.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}

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
func (sm *AWSSecretsManager) UpdateSecret(ctx context.Context, secretID string, value any) error {
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
