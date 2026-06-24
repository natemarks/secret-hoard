package main

import (
	"fmt"
	"os"

	"github.com/natemarks/secret-hoard/pull"
	"github.com/natemarks/secret-hoard/version"
)

func main() {
	fmt.Fprintf(os.Stderr, "sh-pull version: %s\n", version.GetVersion())

	cfg, err := GetConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	log := cfg.GetLogger()
	log.Debug("config: %+v", cfg)

	pullMeta := cfg.ToPullMetadata()
	err = pull.Secret(pullMeta, log)
	if err != nil {
		log.Error("Secret() error: %v", err)
		os.Exit(1)
	}
}
