package main

import (
	"strings"

	"github.com/natemarks/secret-hoard/rdspostgres"
	"github.com/natemarks/secret-hoard/tools"
)

var secrets []rdspostgres.Secret

func main() {
	cfg, err := tools.GetConfig()
	if err != nil {
		panic(err)
	}
	log := cfg.GetLogger()
	log.Info().Msgf("config: %+v", cfg)
	records, err := rdspostgres.RecordsFromCSV(cfg.FilePath, &log)
	if err != nil {
		log.Fatal().Err(err).Msgf("error reading secrets from file %s", cfg.FilePath)
	}
	for _, record := range records {
		// skip header row
		if strings.ToLower(record.ResourceType) == "resourcetype" {
			continue
		}
		secret, err := rdspostgres.FromCSVRecord(record, &log)
		if err != nil {
			log.Error().Err(err).Msgf("error converting record to secret: %v", record)
			continue
		}
		secrets = append(secrets, secret)
	}

	for _, secret := range secrets {
		if secret.Exists(&log) {
			log.Debug().Msgf("secret already exists: %s", secret.Metadata.SecretID())
			secret.Update(cfg.Overwrite, &log)
			continue
		}
		log.Debug().Msgf("secret does not exist: %s", secret.Metadata.SecretID())
		secret.Create(&log)
	}
}
