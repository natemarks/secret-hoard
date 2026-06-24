package secrets

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
)

// GenericExists checks if a secret exists using the SecretsManager interface
func GenericExists(s Secret, log *zerolog.Logger, sm SecretsManager) bool {
	exists, err := sm.DescribeSecret(context.Background(), s.SecretID())
	if err != nil {
		log.Error().Err(err).Msg("error checking if secret exists")
		return false
	}
	return exists
}

// GenericCreate creates a secret using the SecretsManager interface
func GenericCreate(s Secret, log *zerolog.Logger, sm SecretsManager) error {
	tags := s.Metadata()
	err := sm.CreateSecret(context.Background(), s.SecretID(), s.Data(), tags)
	if err != nil {
		log.Error().Err(err).Msgf("error creating secret: %s", s.SecretID())
		return err
	}
	log.Info().Msgf("created secret: %s", s.SecretID())
	return nil
}

// GenericUpdate updates a secret using the SecretsManager interface
func GenericUpdate(s Secret, log *zerolog.Logger, sm SecretsManager, overwrite bool) error {
	if !overwrite {
		log.Info().Msgf("skipping update (overwrite=false): %s", s.SecretID())
		return nil
	}

	err := sm.UpdateSecret(context.Background(), s.SecretID(), s.Data())
	if err != nil {
		log.Error().Err(err).Msgf("error updating secret: %s", s.SecretID())
		return err
	}
	log.Info().Msgf("updated secret: %s", s.SecretID())
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
