package main

import (
	"flag"
	"fmt"

	"github.com/natemarks/secret-hoard/tools"
)

// Config is the configuration for the sh-push command
type Config struct {
	MetadataFile string
	Debug        bool
}

// GetConfig parses command line flags and returns a Config
func GetConfig() (config Config, err error) {
	metadataPtr := flag.String("metadata", "", "Path to metadata JSON file (relative to $HOME/.secret-hoard/ or absolute)")
	debugPtr := flag.Bool("debug", false, "Enable Debug mode")
	flag.Parse()

	config.MetadataFile = *metadataPtr
	config.Debug = *debugPtr

	// Validate required flags
	if config.MetadataFile == "" {
		return config, fmt.Errorf(`metadata file is required

Usage:
  sh-push -metadata=<file>

Examples:
  sh-push -metadata=jsondoc.dev.app.metadata.json
  sh-push -metadata=/path/to/rdspostgres.prod.db.json

Tip: Files are typically in ~/.secret-hoard/`)
	}

	// Resolve path - if not absolute, try relative to working directory
	if config.MetadataFile[0] != '/' {
		workingDir, err := tools.GetWorkingDir()
		if err != nil {
			return config, err
		}
		config.MetadataFile = fmt.Sprintf("%s/%s", workingDir, config.MetadataFile)
	}

	if !tools.FileExists(config.MetadataFile) {
		workingDir, _ := tools.GetWorkingDir()
		return config, fmt.Errorf(`metadata file not found: %s

Tip: Check files in your working directory:
  ls %s

Or use an absolute path:
  sh-push -metadata=/full/path/to/file.json`, config.MetadataFile, workingDir)
	}

	return config, nil
}

// GetLogger returns a configured logger
func (c Config) GetLogger() *tools.Logger {
	return tools.NewLogger(c.Debug)
}
