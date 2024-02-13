package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/natemarks/secret-hoard/tools"

	"github.com/natemarks/secret-hoard/version"
	"github.com/rs/zerolog"
)

// Config is the configuration for the application
type Config struct {
	SecretID string // the secret ID to download
	FilePath string // the file path to write the secret to
	Debug    bool   // enable debug mode
}

// GetLogger returns a logger for the application
func (c Config) GetLogger() (log zerolog.Logger) {
	log = zerolog.New(os.Stdout).With().Str("version", version.Version).Timestamp().Logger()
	log = log.Level(zerolog.InfoLevel)
	if c.Debug {
		log = log.Level(zerolog.DebugLevel)
	}
	return log
}

// GetConfig returns the configuration for the application
func GetConfig() (config Config, err error) {
	// Define flags
	secretIDPtr := flag.String("secret", "", "Secret ID to get")
	filePtr := flag.String("file", "", "Path to the file")
	debugPtr := flag.Bool("debug", false, "Enable Debug mode")

	// Parse command line arguments
	flag.Parse()
	config.FilePath = *filePtr
	config.SecretID = *secretIDPtr
	config.Debug = *debugPtr

	if !tools.FileExists(config.FilePath) {
		return config, fmt.Errorf("file already exists: %s", config.FilePath)
	}
	return config, nil
}
