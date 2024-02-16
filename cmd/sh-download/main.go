package main

import (
	"os"

	"github.com/natemarks/secret-hoard/get"
)

func main() {
	cfg, err := GetConfig()
	if err != nil {
		panic(err)
	}
	log := cfg.GetLogger()
	log.Info().Msgf("config: %+v", cfg)
	err = get.DownloadSecret(cfg.SecretID, cfg.FilePath, &log)
	if err != nil {
		log.Fatal().Err(err).Msg("DownloadSecret() error")
		os.Exit(1)
	}
}
