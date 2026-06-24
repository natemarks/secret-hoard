package main

import (
	"fmt"
	"os"

	"github.com/natemarks/secret-hoard/push"
)

func main() {
	cfg, err := GetConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	log := cfg.GetLogger()
	log.Debug("config: %+v", cfg)

	err = push.PushSecret(cfg.MetadataFile, log)
	if err != nil {
		log.Error("PushSecret() error: %v", err)
		os.Exit(1)
	}
}
