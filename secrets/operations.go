package secrets

import (
	"context"
	"fmt"

	"github.com/natemarks/secret-hoard/tools"
)

// GenericExists checks if a secret exists using the SecretsManager interface
func GenericExists(s Secret, log *tools.Logger, sm SecretsManager) bool {
	exists, err := sm.DescribeSecret(context.Background(), s.SecretID())
	if err != nil {
		log.Error("error checking if secret exists: %v", err)
		return false
	}
	return exists
}

// GenericCreate creates a secret using the SecretsManager interface
func GenericCreate(s Secret, log *tools.Logger, sm SecretsManager) error {
	tags := s.Metadata()
	err := sm.CreateSecret(context.Background(), s.SecretID(), s.Data(), tags)
	if err != nil {
		log.Error("error creating secret %s: %v", s.SecretID(), err)
		return err
	}
	log.Info("created secret: %s", s.SecretID())
	return nil
}

// GenericUpdate updates a secret using the SecretsManager interface
func GenericUpdate(s Secret, log *tools.Logger, sm SecretsManager, overwrite bool) error {
	if !overwrite {
		log.Info("skipping update (overwrite=false): %s", s.SecretID())
		return nil
	}

	err := sm.UpdateSecret(context.Background(), s.SecretID(), s.Data())
	if err != nil {
		log.Error("error updating secret %s: %v", s.SecretID(), err)
		return err
	}
	log.Info("updated secret: %s", s.SecretID())
	return nil
}

// DefaultSecretsManager returns the default AWS Secrets Manager implementation
// This is used by secret type packages that don't need custom implementations
var defaultSM *AWSSecretsManager

func DefaultSecretsManager() SecretsManager {
	if defaultSM == nil {
		sm, err := NewAWSSecretsManager()
		if err != nil {
			panic(fmt.Errorf("failed to create default secrets manager: %w", err))
		}
		defaultSM = sm
	}
	return defaultSM
}
