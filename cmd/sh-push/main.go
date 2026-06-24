package main

import (
	"os"

	"github.com/natemarks/secret-hoard/push"
)

func main() {
	cfg, err := GetConfig()
	if err != nil {
		panic(err)
	}

	log := cfg.GetLogger()
	log.Info().Msgf("config: %+v", cfg)

	err = push.PushSecret(cfg.MetadataFile, &log)
	if err != nil {
		log.Fatal().Err(err).Msg("PushSecret() error")
		os.Exit(1)
	}
}
