package main

import (
	"fmt"
	"os"

	"github.com/natemarks/secret-hoard/pull"
)

func main() {
	cfg, err := GetConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	log := cfg.GetLogger()
	log.Info().Msgf("config: %+v", cfg)

	pullMeta := cfg.ToPullMetadata()
	err = pull.PullSecret(pullMeta, &log)
	if err != nil {
		log.Error().Err(err).Msg("PullSecret() error")
		os.Exit(1)
	}
}
