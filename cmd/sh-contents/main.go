package main

import (
	"fmt"
	"os"

	"github.com/natemarks/secret-hoard/contents"
	"github.com/natemarks/secret-hoard/version"
)

func main() {
	cfg, err := GetConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	log := cfg.GetLogger()

	// Only show version in debug mode
	if cfg.Debug {
		fmt.Fprintf(os.Stderr, "sh-contents version: %s\n", version.GetVersion())
		log.Debug("config: %+v", cfg)
	}

	if cfg.TargetDir == "" {
		if err := contents.EmitSecretContents(cfg.SecretID, os.Stdout, log); err != nil {
			log.Error("EmitSecretContents() error: %v", err)
			os.Exit(1)
		}
		return
	}

	paths, err := contents.WriteSecretContents(cfg.SecretID, cfg.TargetDir, log)
	if err != nil {
		log.Error("WriteSecretContents() error: %v", err)
		os.Exit(1)
	}

	// Print paths to stdout (one per line) so they can be captured in bash scripts
	for _, path := range paths {
		fmt.Println(path)
	}
}
