package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"

	"github.com/natemarks/secret-hoard/awssm"
	"github.com/natemarks/secret-hoard/csv"
	"github.com/natemarks/secret-hoard/version"
	"github.com/rs/zerolog"
)

type config struct {
	overwrite  bool
	secretType string
	filePath   string
}

func getConfig() (config config, err error) {

	secretTypePtr := flag.String("type", "", "secret type: rds, snowflake")
	filePathPtr := flag.String("file", "", "path to csv file")
	overwritePtr := flag.Bool("overwrite", false, "overwrite existing secrets")
	flag.Parse()

	switch *secretTypePtr {
	case "rds":
		config.secretType = "rds"
	case "snowflake":
		config.secretType = "snowflake"
	default:
		return config, fmt.Errorf("invalid secret type: %s", *secretTypePtr)
	}
	if _, err := os.Stat(*filePathPtr); err == nil {
		config.filePath = *filePathPtr
	} else {
		return config, fmt.Errorf("invalid file path: %s", *filePathPtr)
	}
	config.overwrite = *overwritePtr
	return config, nil
}
func main() {

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	logger := zerolog.New(os.Stderr).With().Str("version", version.Version).Timestamp().Logger()
	config, err := getConfig()
	if err != nil {
		logger.Fatal().Err(err).Msgf("error getting config: %v", err)
	}
	logger.Info().Msgf("parsing file: %v", config.filePath)
	switch config.secretType {
	case "rds":
		secrets, err := csv.ReadRDSSecrets(config.filePath, &logger)
		if err != nil {
			logger.Fatal().Err(err).Msgf("error reading RDS secrets from file %s", config.filePath)
		}
		err = awssm.CreateRDSSecrets(secrets, &logger)
		if err != nil {
			logger.Fatal().Err(err).Msg("error creating RDS secrets")
		}
		log.Info().Msg("RDS secrets created successfully")
	case "snowflake":
		secrets, err := csv.ReadSnowflakeSecrets(config.filePath, &logger)
		if err != nil {
			logger.Fatal().Err(err).Msgf("error reading Snowflake secrets from file %s", config.filePath)
		}
		err = awssm.CreateSnowflakeSecrets(secrets, &logger)
		if err != nil {
			logger.Fatal().Err(err).Msg("error creating Snowflake secrets")
		}
		log.Info().Msg("Snowflake secrets created successfully")
	default:
		logger.Fatal().Msgf("invalid secret type: %s", config.secretType)
	}

}
