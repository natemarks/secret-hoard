package csv

import (
	"encoding/csv"
	"os"
	"strconv"
	"strings"

	"github.com/natemarks/secret-hoard/rds"

	"github.com/natemarks/secret-hoard/types"
	"github.com/rs/zerolog"
)

// ReadRDSSecrets reads a CSV file and returns a slice of RDSSecrets
func ReadRDSSecrets(filename string, log *zerolog.Logger) ([]types.RDSSecret, error) {
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
		host, err := rds.GetRDSEndpoint(record[8])
		if err != nil {
			log.Error().Err(err).Msgf("error getting RDS endpoint for %s", record[8])
		} else {
			secret.Data.Host = host
		}
		secrets = append(secrets, secret)
	}

	return secrets, nil
}

// WriteRDSSecrets writes a slice of RDSSecrets to a CSV file
func WriteRDSSecrets(filename string, secrets []types.RDSSecret, log *zerolog.Logger) error {
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
		"ResourceType", "Environment", "Instance", "Database", "Access",
		"Password", "Engine", "Port", "DbInstanceIdentifier", "Host", "Username",
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
			log.Error().Err(err).Msgf("error writing record %v to file %s", record, filename)
			return err
		}
	}

	writer.Flush()
	return writer.Error()
}
