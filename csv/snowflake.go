package csv

import (
	"encoding/csv"
	"os"
	"strings"

	"github.com/natemarks/secret-hoard/types"
	"github.com/rs/zerolog"
)

// ReadSnowflakeSecrets reads a CSV file and returns a slice of Snowflake
func ReadSnowflakeSecrets(filename string, log *zerolog.Logger) ([]types.SnowflakeSecret, error) {
	file, err := os.Open(filename)
	if err != nil {
		log.Error().Err(err).Msgf("error opening file %s", filename)
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Error().Err(err).Msgf("error closing file %s", filename)
		}
	}(file)

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Error().Err(err).Msgf("error reading file %s", filename)
		return nil, err
	}

	var secrets []types.SnowflakeSecret

	for _, record := range records {
		// Assuming CSV columns are in order: ResourceType, Environment, Warehouse, Access,
		// Password, AccountName, Username
		if strings.ToLower(record[0]) == "resourcetype" {
			continue
		}
		secret := types.SnowflakeSecret{
			Data: types.SnowflakeSecretData{
				Password:    record[4],
				AccountName: record[5],
				Warehouse:   record[2],
				Username:    record[6],
			},
			Metadata: types.SnowflakeSecretMetadata{
				ResourceType: record[0],
				Environment:  record[1],
				Warehouse:    record[2],
				Access:       record[3],
			},
		}
		secrets = append(secrets, secret)
	}

	return secrets, nil
}

// WriteSnowflakeSecrets writes a slice of SnowflakeSecrets to a CSV file
func WriteSnowflakeSecrets(filename string, secrets []types.SnowflakeSecret, log *zerolog.Logger) error {
	file, err := os.Create(filename)
	if err != nil {
		log.Error().Err(err).Msgf("error creating file %s", filename)
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	writer := csv.NewWriter(file)

	// Write CSV header
	err = writer.Write([]string{
		"ResourceType", "Environment", "Warehouse", "Access",
		"Password", "AccountName", "Username",
	})
	if err != nil {
		log.Error().Err(err).Msgf("error writing header to file %s", filename)
		return err
	}

	// Write data rows
	for _, secret := range secrets {
		record := []string{
			secret.Metadata.ResourceType,
			secret.Metadata.Environment,
			secret.Metadata.Warehouse,
			secret.Metadata.Access,
			secret.Data.Password,
			secret.Data.AccountName,
			secret.Data.Warehouse,
			secret.Data.Username,
		}
		err := writer.Write(record)
		if err != nil {
			log.Error().Err(err).Msgf("error writing record %v to file %s", record, filename)
			return err
		}
	}

	writer.Flush()
	return writer.Error()
}
