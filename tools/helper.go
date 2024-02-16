package tools

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/natemarks/secret-hoard/version"
	"github.com/rs/zerolog"

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
func DeleteSecrets(secretIDs []string) {
	ctx := context.Background()
	cfg, _ := config.LoadDefaultConfig(ctx)

	client := secretsmanager.NewFromConfig(cfg)

	for _, secretID := range secretIDs {
		_, _ = client.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
			SecretId:                   &secretID,
			ForceDeleteWithoutRecovery: aws.Bool(true),
		})

	}
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

func TestLogger() (log zerolog.Logger) {
	log = zerolog.New(os.Stdout).With().Str("version", version.Version).Timestamp().Logger()
	log = log.With().Str("aws_account_number", GetAWSAccountNumber()).Logger()
	log = log.Level(zerolog.DebugLevel)
	return log
}
