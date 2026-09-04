package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/natemarks/secret-hoard/tools"
)

// Config is the configuration for the sh-contents command
type Config struct {
	SecretID  string
	TargetDir string
	Debug     bool
}

// GetConfig parses command line flags and returns a Config
func GetConfig() (config Config, err error) {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] <secret-id> [target-directory]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Download secret contents.\n\n")
		fmt.Fprintf(os.Stderr, "  With target-directory: write to files in that directory (all types).\n")
		fmt.Fprintf(os.Stderr, "  Without target-directory: write content to stdout (jsondoc, textfile only).\n\n")
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  secret-id          Secret ID (e.g., jsondoc/dev/app-config)\n")
		fmt.Fprintf(os.Stderr, "  target-directory   Directory to write files (required for sslcert)\n\n")
		fmt.Fprintf(os.Stderr, "Supported secret types:\n")
		fmt.Fprintf(os.Stderr, "  - jsondoc          Stdout or writes .contents.json file\n")
		fmt.Fprintf(os.Stderr, "  - textfile         Stdout or writes .contents.txt file\n")
		fmt.Fprintf(os.Stderr, "  - sslcert          Requires target-directory; writes .crt and .key files\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s jsondoc/dev/app-config > config.json\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s textfile/prod/api-key > api-key.txt\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s jsondoc/dev/app-config $HOME\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s sslcert/prod/example.com /etc/ssl\n", os.Args[0])
	}

	debugPtr := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	config.Debug = *debugPtr

	// Check for positional arguments
	args := flag.Args()
	if len(args) == 1 {
		config.SecretID = args[0]
		return config, nil
	}

	if len(args) == 2 {
		config.SecretID = args[0]
		config.TargetDir = args[1]

		// Convert to absolute path
		absPath, err := filepath.Abs(config.TargetDir)
		if err != nil {
			return config, fmt.Errorf("error resolving target directory: %w", err)
		}
		config.TargetDir = absPath

		// Check if directory exists
		if !tools.FileExists(config.TargetDir) {
			return config, fmt.Errorf("target directory does not exist: %s", config.TargetDir)
		}

		return config, nil
	}

	flag.Usage()
	return config, fmt.Errorf("requires 1 or 2 arguments: secret-id [target-directory]")
}

// GetLogger returns a configured logger
func (c Config) GetLogger() *tools.Logger {
	return tools.NewLogger(c.Debug)
}
