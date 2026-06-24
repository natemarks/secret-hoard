package main

import (
	"flag"

	"github.com/natemarks/secret-hoard/tools"
	"github.com/rs/zerolog"
)

// Config is the configuration for the sh-generate command
type Config struct {
	Debug bool
}

// GetConfig parses command line flags and returns a Config
func GetConfig() (config Config, err error) {
	debugPtr := flag.Bool("debug", false, "Enable Debug mode")
	flag.Parse()

	config.Debug = *debugPtr

	return config, nil
}

// GetLogger returns a configured logger
func (c Config) GetLogger() zerolog.Logger {
	// Use simple logger without AWS account number since sh-generate is local-only
	log := tools.SimpleLogger()
	if !c.Debug {
		log = log.Level(zerolog.InfoLevel)
	}
	return log
}
