package aws

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/rds"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/natemarks/secret-hoard/types"
	"github.com/rs/zerolog"
)

// CreateRDSSecrets creates RDS secrets in AWS Secrets Manager
func CreateRDSSecrets(secrets []types.RDSSecret, log *zerolog.Logger) error {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	client := secretsmanager.NewFromConfig(cfg)

	for _, secret := range secrets {
		// Convert RDSSecretData to JSON string
		secretValue, err := json.Marshal(secret.Data)
		if err != nil {
			log.Error().Err(err).Msg("error marshalling secret data")
			continue
		}

		// Convert RDSSecretMetadata to tags
		tags := map[string]string{
			"ResourceType": secret.Metadata.ResourceType,
			"Environment":  secret.Metadata.Environment,
			"Instance":     secret.Metadata.Instance,
			"Database":     secret.Metadata.Database,
			"Access":       secret.Metadata.Access,
			"Source":       "secret-hoard",
		}

		// Create the secret
		createSecretInput := &secretsmanager.CreateSecretInput{
			Name:         aws.String(fmt.Sprint(secret.Metadata.SecretID())),
			SecretString: aws.String(string(secretValue)),
			Tags:         ConvertMapToTags(tags),
		}

		_, err = client.CreateSecret(ctx, createSecretInput)
		if err != nil {
			log.Error().Err(err).Msg("error creating secret")
			continue
		}
		log.Info().Msgf("secret created successfully: %s", *createSecretInput.Name)
	}

	return nil
}

// GetRDSEndpoint returns the endpoint of the RDS instance
func GetRDSEndpoint(instanceID string) (string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return "", err
	}

	client := rds.NewFromConfig(cfg)

	params := &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: &instanceID,
	}

	resp, err := client.DescribeDBInstances(context.TODO(), params)
	if err != nil {
		return "", err
	}

	if len(resp.DBInstances) == 0 {
		return "", fmt.Errorf("no RDS instance found with ID: %s", instanceID)
	}

	return *resp.DBInstances[0].Endpoint.Address, nil
}
