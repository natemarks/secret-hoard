package sslcert

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/natemarks/secret-hoard/tools"

	"github.com/rs/zerolog"
)

// Metadata server certificate secret metadata for tagging
type Metadata struct {
	ResourceType string `json:"resourceType"` // ssl_certificate
	Environment  string `json:"environment"`  // dev, integration, staging, production
	CommonName   string `json:"commonName"`   // \*.my.domain.com | server.my.domain.com
}

// Map converts RDSSecretMetadata to a map of strings to simplify tagging
func (sfm Metadata) Map() map[string]string {
	attributes := map[string]string{
		"ResourceType": sfm.ResourceType,
		"Environment":  sfm.Environment,
		"CommonName":   sfm.CommonName,
		"Source":       "secret-hoard",
	}
	return attributes
}

// SecretID returns the secret id for the rdsSecret
func (sfm Metadata) SecretID() string {
	return fmt.Sprintf("%v/%v/%v", sfm.ResourceType, sfm.Environment, sfm.CommonName)
}

// Data is the struct of the secret for s snowflake connection
type Data struct {
	Certificate       string `json:"certificate"`       // Certificate file contents as string
	PrivateKey        string `json:"key"`               // PrivateKey file contents as string
	ExpirationDate    string `json:"expirationDate"`    // Expiration date in ISO 3339 format
	Modulus           string `json:"modulus"`           // Modulus of the certificate and PrivateKey
	CertificateSha256 string `json:"certificateSha256"` // SHA256 hash of the certificate file
	PrivateKeySha256  string `json:"privateKeySha256"`  // SHA256 hash of the PrivateKey file
}

// Secret is the struct of the secret for snowflake
type Secret struct {
	Data     Data
	Metadata Metadata
}

// Exists checks if the secret exists in Secrets Manager
func (s Secret) Exists(log *zerolog.Logger) bool {

	// Load AWS SDK configuration
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("unable to load SDK config")
	}

	// Create Secrets Manager client
	client := secretsmanager.NewFromConfig(cfg)

	// Input parameters for DescribeSecret API call
	input := &secretsmanager.DescribeSecretInput{
		SecretId: aws.String(s.Metadata.SecretID()),
	}

	// Call DescribeSecret API to check if the secret exists
	_, err = client.DescribeSecret(context.Background(), input)
	if err != nil {
		var e *types.ResourceNotFoundException
		if errors.As(err, &e) {
			log.Debug().Msgf("secret does not exist: %s", *input.SecretId)
			return false
		}
	}
	log.Debug().Msgf("secret exists: %s", *input.SecretId)
	return true
}

// Create the Secret
func (s Secret) Create(log *zerolog.Logger) {
	log.Debug().Msgf("creating RDS rdsSecret: %s", s.Metadata.SecretID())
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	client := secretsmanager.NewFromConfig(cfg)

	// Convert RDSSecretData to JSON string
	secretValue, err := json.Marshal(s.Data)
	if err != nil {
		log.Error().Err(err).Msg("error marshalling secret data")
		return
	}

	// Convert RDSSecretMetadata to tags
	tags := s.Metadata.Map()

	// Create the rdsSecret
	createSecretInput := &secretsmanager.CreateSecretInput{
		Name:         aws.String(fmt.Sprint(s.Metadata.SecretID())),
		SecretString: aws.String(string(secretValue)),
		Tags:         tools.ConvertMapToTags(tags),
	}
	_, err = client.CreateSecret(ctx, createSecretInput)
	// If the secret already exists and overwrite is true, update it
	if err != nil {
		log.Error().Err(err).Msgf("error creating rdsSecret: %s", *createSecretInput.Name)
		return
	}
	log.Info().Msgf("secret created successfully: %s", *createSecretInput.Name)
}

// Update the rdsSecret
func (s Secret) Update(overwrite bool, log *zerolog.Logger) {
	if !overwrite {
		log.Debug().Msgf("overwrite is false, skipping update for %s", s.Metadata.SecretID())
		return
	}
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	client := secretsmanager.NewFromConfig(cfg)

	// Convert RDSSecretData to JSON string
	secretValue, err := json.Marshal(s.Data)
	if err != nil {
		log.Error().Err(err).Msg("error marshalling secret data")
		return
	}

	// Convert RDSSecretMetadata to tags
	tags := s.Metadata.Map()

	// Create the rdsSecret
	// Update the secret string value
	updateSecretInput := &secretsmanager.UpdateSecretInput{
		SecretId:     aws.String(fmt.Sprint(s.Metadata.SecretID())),
		SecretString: aws.String(string(secretValue)),
	}
	_, err = client.UpdateSecret(ctx, updateSecretInput)
	// If the secret already exists and overwrite is true, update it
	if err != nil {
		log.Error().Err(err).Msgf("error updating secret value: %s", *updateSecretInput.SecretId)
		return
	}

	// Update the secret tags
	tagResourceInput := &secretsmanager.TagResourceInput{
		SecretId: aws.String(fmt.Sprint(s.Metadata.SecretID())),
		Tags:     tools.ConvertMapToTags(tags),
	}
	_, err = client.TagResource(ctx, tagResourceInput)
	if err != nil {
		log.Error().Err(err).Msgf("error updating secret tags: %s", *updateSecretInput.SecretId)
		return
	}
	log.Info().Msgf("secret update successfully: %s", *updateSecretInput.SecretId)
}

// FromCSVRecord converts a CSV record to a valid Secret
func FromCSVRecord(record Record, log *zerolog.Logger) (secret Secret, err error) {
	// set logger context for this record
	*log = log.With().Str("environment", record.Environment).Str("commonName", record.CommonName).Logger()
	*log = log.With().Str("certificateFile", record.CertificateFile).Logger()
	*log = log.With().Str("privateKeyFile", record.PrivateKeyFile).Logger()

	//compare certificate and privateKey moduli
	certModulus, err := record.CertificateModulus()
	if err != nil {
		log.Error().Err(err).Msg("error getting certificate modulus")
		return secret, err
	}
	privateKeyModulus, err := record.PrivateKeyModulus()
	if err != nil {
		log.Error().Err(err).Msg("error getting privateKey modulus")
		return secret, err
	}
	if certModulus != privateKeyModulus {
		log.Error().Msg("certificate and privateKey moduli do not match")
		return secret, fmt.Errorf("certificate and privateKey moduli do not match")
	}
	log.Debug().Msgf("certificate and privateKey moduli match: %s", certModulus)

	expiration, err := record.Expiration()
	if err != nil {
		log.Error().Err(err).Msg("error getting certificate expiration")
		return secret, err
	}
	log.Debug().Msgf("certificate expiration: %s", expiration)

	certificateSha256Sum, err := record.CertificateSha256Sum()
	if err != nil {
		log.Error().Err(err).Msg("error getting certificate sha256 sum")
		return secret, err
	}

	privateKeySha256Sum, err := record.PrivateKeySha256Sum()
	if err != nil {
		log.Error().Err(err).Msg("error getting privateKey sha256 sum")
		return secret, err
	}

	certificateContents, err := record.CertificateContents()
	if err != nil {
		log.Error().Err(err).Msg("error getting certificate contents")
		return secret, err
	}

	privateKeyContents, err := record.PrivateKeyContents()
	if err != nil {
		log.Error().Err(err).Msg("error getting privateKey contents")
		return secret, err
	}

	secret = Secret{
		Data: Data{
			Certificate:       certificateContents,
			PrivateKey:        privateKeyContents,
			ExpirationDate:    expiration,
			Modulus:           certModulus,
			CertificateSha256: certificateSha256Sum,
			PrivateKeySha256:  privateKeySha256Sum,
		},
		Metadata: Metadata{
			ResourceType: record.ResourceType,
			Environment:  record.Environment,
			CommonName:   record.CommonName,
		},
	}
	log.Debug().Msgf("new secret from CSV: %v", secret.Metadata.SecretID())
	return secret, err
}
