package main

import (
	"github.com/natemarks/secret-hoard/tools"
	"github.com/natemarks/secret-hoard/uploader"
	"github.com/rs/zerolog"
)

// GetCSVProcessor returns the appropriate CSVProcessor based on the config
func GetCSVProcessor(cfg tools.Config, log *zerolog.Logger) uploader.CSVProcessor {
	switch csvType := cfg.CSVType(); csvType {
	case "rdspostgres":
		return uploader.RDSPostgresProcessor{}
	case "snowflake":
		return uploader.SnowflakeProcessor{}
	case "text_file":
		return uploader.TextFileProcessor{}
	case "jsondoc":
		return uploader.JSONDocProcessor{}
	case "ssl_certificate":
		return uploader.SSLCertProcessor{}
	default:
		log.Fatal().Msgf("unknown CSV type: %s", csvType)
		return nil

	}
}
func main() {
	cfg, err := tools.GetConfig()
	if err != nil {
		panic(err)
	}
	log := cfg.GetLogger()
	log.Info().Msgf("config: %+v", cfg)
	processor := GetCSVProcessor(cfg, &log)
	processor.Process(cfg, &log)

}
