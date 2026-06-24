package main

import (
	"fmt"
	"os"

	"github.com/natemarks/secret-hoard/generate"
	"github.com/natemarks/secret-hoard/version"
)

func main() {
	fmt.Fprintf(os.Stderr, "sh-generate version: %s\n", version.GetVersion())

	cfg, err := GetConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	log := cfg.GetLogger()
	log.Debug("config: %+v", cfg)

	err = generate.SecretFiles(log)
	if err != nil {
		log.Error("SecretFiles() error: %v", err)
		os.Exit(1)
	}
}
