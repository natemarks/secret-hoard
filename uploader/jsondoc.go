package uploader

import (
	"strings"

	"github.com/natemarks/secret-hoard/jsondoc"
	"github.com/natemarks/secret-hoard/tools"
	"github.com/rs/zerolog"
)

// JSONDocProcessor implement CSVProcessor for rdspostgres secrets
type JSONDocProcessor struct{}

// Process handles the jsondoc secrets CSV files
func (j JSONDocProcessor) Process(cfg tools.Config, log *zerolog.Logger) {
	var secrets []jsondoc.Secret
	records, err := jsondoc.RecordsFromCSV(cfg.FilePath, log)
	if err != nil {
		log.Fatal().Err(err).Msgf("error reading secrets from file %s", cfg.FilePath)
	}
	for _, record := range records {
		// skip header row
		if strings.ToLower(record.ResourceType) == "resourcetype" {
			continue
		}
		secret, err := jsondoc.FromCSVRecord(record, log)
		if err != nil {
			log.Error().Err(err).Msgf("error converting record to secret: %v", record)
			continue
		}
		secrets = append(secrets, secret)
	}

	for _, secret := range secrets {
		if secret.Exists(log) {
			secret.Update(cfg.Overwrite, log)
			continue
		}
		secret.Create(log)
	}
}
