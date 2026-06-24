package main

import (
	"fmt"
	"os"

	"github.com/natemarks/secret-hoard/tools"
	"github.com/natemarks/secret-hoard/uploader"
)

// GetCSVProcessor returns the appropriate CSVProcessor based on the config
func GetCSVProcessor(cfg tools.Config, log *tools.Logger) uploader.CSVProcessor {
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
		log.Error("unknown CSV type: %s", csvType)
		os.Exit(1)
		return nil
	}
}

func main() {
	cfg, err := tools.GetConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}
	log := cfg.GetLogger()
	log.Debug("config: %+v", cfg)
	processor := GetCSVProcessor(cfg, log)
	processor.Process(cfg, log)

}
