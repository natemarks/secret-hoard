package uploader

import (
	"strings"

	"github.com/natemarks/secret-hoard/snowflake"
	"github.com/natemarks/secret-hoard/tools"

)

// SnowflakeProcessor implement CSVProcessor for snowflake secrets
type SnowflakeProcessor struct{}

// Process handles the snowflake secrets CSV files
func (s SnowflakeProcessor) Process(cfg tools.Config, log *tools.Logger) {
	var secrets []snowflake.Secret
	records, err := snowflake.RecordsFromCSV(cfg.FilePath, log)
	if err != nil {
		log.Fatal("error reading secrets from file %s", cfg.FilePath)
	}
	for _, record := range records {
		// skip header row
		if strings.ToLower(record.ResourceType) == "resourcetype" {
			continue
		}
		secret, err := snowflake.FromCSVRecord(record, log)
		if err != nil {
			log.Error("error converting record to secret: %v", record)
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
