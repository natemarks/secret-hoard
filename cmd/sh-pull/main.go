package main

import (
	"os"

	"github.com/natemarks/secret-hoard/pull"
)

func main() {
	cfg, err := GetConfig()
	if err != nil {
		panic(err)
	}

	log := cfg.GetLogger()
	log.Info().Msgf("config: %+v", cfg)

	pullMeta := cfg.ToPullMetadata()
	err = pull.PullSecret(pullMeta, &log)
	if err != nil {
		log.Fatal().Err(err).Msg("PullSecret() error")
		os.Exit(1)
	}
}
