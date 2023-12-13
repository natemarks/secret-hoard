package store

import (
	"encoding/csv"
	"strconv"
	"strings"

	"github.com/natemarks/secret-hoard/types"
	"github.com/rs/zerolog"
)

// ReadRDSSecrets reads a CSV file and returns a slice of RDSSecrets
func ReadRDSSecrets(filename string, log *zerolog.Logger) ([]types.RDSSecret, error) {
	csvContents, err := readFileToString(filename)
	if err != nil {
		log.Error().Err(err).Msgf("error reading store contents from file %s", filename)
		return nil, err
	}
	secrets, err := RDSSecretsFromCSVString(csvContents, log)
	if err != nil {
		log.Error().Err(err).Msgf("error converting CSV contents to RDSSecrets")
		return nil, err
	}
	return secrets, nil
}

// RDSSecretsFromCSVString reads string contents in CSV format and returns a slice of RDSSecrets
func RDSSecretsFromCSVString(csvData string, log *zerolog.Logger) ([]types.RDSSecret, error) {

	reader := csv.NewReader(strings.NewReader(csvData))
	records, err := reader.ReadAll()
	if err != nil {
		log.Error().Err(err).Msg("error reading store contents string")
		return nil, err
	}

	var secrets []types.RDSSecret

	for _, record := range records {
		// Assuming CSV columns are in order: ResourceType, Environment, Instance, Database, Access,
		// Password, Engine, Port, DbInstanceIdentifier, Host, Username
		if strings.ToLower(record[0]) == "resourcetype" {
			continue
		}
		port, err := strconv.Atoi(record[7])
		if err != nil {
			log.Error().Err(err).Msgf("error converting port %s to int", record[6])
			continue
		}
		secret := types.RDSSecret{
			Data: types.RDSSecretData{
				Password:             record[5],
				Engine:               record[6],
				Port:                 port,
				DbInstanceIdentifier: record[8],
				Host:                 record[9],
				Username:             record[10],
			},
			Metadata: types.RDSSecretMetadata{
				ResourceType: record[0],
				Environment:  record[1],
				Instance:     record[2],
				Database:     record[3],
				Access:       record[4],
			},
		}
		// try to override the endpoint by looking up the RDS instance
		host, err := GetRDSEndpoint(record[8])
		if err != nil {
			log.Error().Err(err).Msgf("error getting RDS endpoint for %s", record[8])
		} else {
			secret.Data.Host = host
		}
		secrets = append(secrets, secret)
	}

	return secrets, nil
}

// RDSSecretsToCSVString writes a slice of RDSSecrets to a CSV file
func RDSSecretsToCSVString(secrets []types.RDSSecret, log *zerolog.Logger) (csvData string, err error) {
	var csvString strings.Builder
	writer := csv.NewWriter(&csvString)

	// Write CSV header
	err = writer.Write([]string{
		"ResourceType", "Environment", "Instance", "Database", "Access",
		"Password", "Engine", "Port", "DbInstanceIdentifier", "Host", "Username",
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
			secret.Metadata.Instance,
			secret.Metadata.Database,
			secret.Metadata.Access,
			secret.Data.Password,
			secret.Data.Engine,
			strconv.Itoa(secret.Data.Port),
			secret.Data.DbInstanceIdentifier,
			secret.Data.Host,
			secret.Data.Username,
		}
		err := writer.Write(record)
		if err != nil {
			log.Error().Err(err).Msgf("error writing record %v", record)
			return "", err
		}
	}

	writer.Flush()
	csvData = csvString.String()
	return csvData, writer.Error()
}
