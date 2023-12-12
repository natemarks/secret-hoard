package aws

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	smtypes "github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
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

// GetSecretsMatchingTags returns a list of secret names that match the given tags
func GetSecretsMatchingTags(tagFilters map[string]string) ([]string, error) {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	client := secretsmanager.NewFromConfig(cfg)

	params := &secretsmanager.ListSecretsInput{}
	paginator := secretsmanager.NewListSecretsPaginator(client, params)

	var matchingSecrets []string

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, secret := range page.SecretList {

			if tagsMatch(tagFilters, secret.Tags) {
				matchingSecrets = append(matchingSecrets, *secret.Name)
			}
		}
	}

	return matchingSecrets, nil
}

func tagsMatch(expectedTags map[string]string, actualTags []smtypes.Tag) bool {
	for key, value := range expectedTags {
		found := false
		for _, tag := range actualTags {
			if *tag.Key == key && *tag.Value == value {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// CredsOK checks that the AWS credentials are valid
func CredsOK() (err error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return err
	}

	svc := sts.NewFromConfig(cfg)
	result, err := svc.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
	if err != nil {
		return err
	}
	if *result.Account == "709310380790" {
		return nil
	}
	return fmt.Errorf("AWS Account ID is not 709310380790, got %s", *result.Account)
}

func existingTestSecrets() (result []string, err error) {
	result, err = GetSecretsMatchingTags(map[string]string{
		"Environment": "testenv",
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func deleteSecrets(secretIDs []string) error {
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

func setup(t *testing.T) error {
	secretIDs, err := existingTestSecrets()
	if err != nil {
		return err
	}
	for _, secretID := range secretIDs {
		t.Logf("deleting secret: %s", secretID)
	}
	err = deleteSecrets(secretIDs)
	if err != nil {
		return err
	}
	if len(secretIDs) > 0 {
		t.Logf("waiting 30 seconds for secrets to be deleted")
		time.Sleep(30 * time.Second)
	}
	return nil
}
