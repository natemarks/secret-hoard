package main

import (
	"os"

	"github.com/natemarks/secret-hoard/generate"
)

func main() {
	cfg, err := GetConfig()
	if err != nil {
		panic(err)
	}

	log := cfg.GetLogger()
	log.Info().Msgf("config: %+v", cfg)

	err = generate.GenerateSecretFiles(&log)
	if err != nil {
		log.Fatal().Err(err).Msg("GenerateSecretFiles() error")
		os.Exit(1)
	}
}
