package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	smtypes "github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
)

// ConvertMapToTags Convert a map to a list of tags
func ConvertMapToTags(tags map[string]string) []smtypes.Tag {
	var tagList []smtypes.Tag
	for key, value := range tags {
		tag := smtypes.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		}
		tagList = append(tagList, tag)
	}
	return tagList
}

// DeleteSecrets deletes the given secrets
func DeleteSecrets(secretIDs []string) error {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}

	client := secretsmanager.NewFromConfig(cfg)

	for _, secretID := range secretIDs {
		_, err := client.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
			SecretId:                   &secretID,
			ForceDeleteWithoutRecovery: aws.Bool(true),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

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

// GetResourceTypeFromSecretID returns the resource type from a secret ID
func GetResourceTypeFromSecretID(secretID string) (result string, err error) {
	parts := strings.Split(secretID, "/")
	if len(parts) < 3 {
		return "", fmt.Errorf("invalid secret ID: %s", secretID)
	}
	return parts[0], nil
}
