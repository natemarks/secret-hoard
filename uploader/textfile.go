package uploader

import (
	"strings"

	"github.com/natemarks/secret-hoard/textfile"
	"github.com/natemarks/secret-hoard/tools"
	"github.com/rs/zerolog"
)

// TextFileProcessor implement CSVProcessor for rdspostgres secrets
type TextFileProcessor struct{}

// Process handles the jsondoc secrets CSV files
func (t TextFileProcessor) Process(cfg tools.Config, log *zerolog.Logger) {
	var secrets []textfile.Secret
	records, err := textfile.RecordsFromCSV(cfg.FilePath, log)
	if err != nil {
		log.Fatal().Err(err).Msgf("error reading secrets from file %s", cfg.FilePath)
	}
	for _, record := range records {
		// skip header row
		if strings.ToLower(record.ResourceType) == "resourcetype" {
			continue
		}
		secret, err := textfile.FromCSVRecord(record, log)
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
