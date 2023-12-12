package store

import (
	"encoding/csv"
	"strings"

	"github.com/natemarks/secret-hoard/types"
	"github.com/rs/zerolog"
)

// ReadSnowflakeSecrets reads a CSV file and returns a slice of Snowflake
func ReadSnowflakeSecrets(filename string, log *zerolog.Logger) ([]types.SnowflakeSecret, error) {
	csvContents, err := readFileToString(filename)
	if err != nil {
		log.Error().Err(err).Msgf("error reading store contents from file %s", filename)
		return nil, err
	}
	secrets, err := SnowflakeSecretsFromCSVString(csvContents, log)
	if err != nil {
		log.Error().Err(err).Msgf("error converting CSV contents to SnowflakeSecrets")
		return nil, err
	}
	return secrets, nil
}

// ReadSnowflakeSecrets reads a CSV file and returns a slice of Snowflake
func SnowflakeSecretsFromCSVString(csvData string, log *zerolog.Logger) ([]types.SnowflakeSecret, error) {

	reader := csv.NewReader(strings.NewReader(csvData))
	records, err := reader.ReadAll()
	if err != nil {
		log.Error().Err(err).Msg("error reading store contents string")
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

// SnowflakeSecretsToCSvString converts a slice of SnowflakeSecrets to a CSV string
func SnowflakeSecretsToCSvString(secrets []types.SnowflakeSecret, log *zerolog.Logger) (csvData string, err error) {
	var csvString strings.Builder
	writer := csv.NewWriter(&csvString)

	// Write CSV header
	err = writer.Write([]string{
		"ResourceType", "Environment", "Warehouse", "Access",
		"Password", "AccountName", "Username",
	})
	if err != nil {
		log.Error().Err(err).Msgf("error writing header")
		return "", err
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
			secret.Data.Username,
		}
		err := writer.Write(record)
		if err != nil {
			log.Error().Err(err).Msgf("error writing record %v ", record)
			return "", err
		}
	}

	writer.Flush()
	csvData = csvString.String()
	return csvData, writer.Error()
}
