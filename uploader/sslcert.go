package uploader

import (
	"strings"

	"github.com/natemarks/secret-hoard/sslcert"
	"github.com/natemarks/secret-hoard/tools"
	"github.com/rs/zerolog"
)

// SSLCertProcessor implement CSVProcessor for rdspostgres secrets
type SSLCertProcessor struct{}

// Process handles the jsondoc secrets CSV files
func (s SSLCertProcessor) Process(cfg tools.Config, log *zerolog.Logger) {
	var secrets []sslcert.Secret
	records, err := sslcert.RecordsFromCSV(cfg.FilePath, log)
	if err != nil {
		log.Fatal().Err(err).Msgf("error reading secrets from file %s", cfg.FilePath)
	}
	for _, record := range records {
		// skip header row
		if strings.ToLower(record.ResourceType) == "resourcetype" {
			continue
		}
		secret, err := sslcert.FromCSVRecord(record, log)
		if err != nil {
			log.Error().Err(err).Msgf("error converting record to secret: %v", record)
			continue
		}
		secrets = append(secrets, secret)
	}

	for _, secret := range secrets {
		if secret.Exists(log) {
			log.Debug().Msgf("secret already exists: %s", secret.Metadata.SecretID())
			secret.Update(cfg.Overwrite, log)
			continue
		}
		log.Debug().Msgf("secret does not exist: %s", secret.Metadata.SecretID())
		secret.Create(log)
	}
}
