package main

import (
	"flag"
	"fmt"
	"os"

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

}
