package tools

import (
	"flag"
	"fmt"
	"os"

	"github.com/natemarks/secret-hoard/version"
	"github.com/rs/zerolog"
)

// Config is the configuration for the application
type Config struct {
	Overwrite bool
	FilePath  string
	Debug     bool
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
	filePtr := flag.String("file", "", "Path to the file")
	overwritePtr := flag.Bool("Overwrite", false, "Overwrite the secret value if it exists")
	debugPtr := flag.Bool("debug", false, "Enable Debug mode")

	// Parse command line arguments
	flag.Parse()
	config.FilePath = *filePtr
	config.Overwrite = *overwritePtr
	config.Debug = *debugPtr

	if !FileExists(config.FilePath) {
		return config, fmt.Errorf("invalid file path: %s", config.FilePath)
	}
	return config, nil
}
