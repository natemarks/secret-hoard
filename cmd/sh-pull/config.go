package main

import (
	"flag"
	"fmt"

	"github.com/natemarks/secret-hoard/generate"
	"github.com/natemarks/secret-hoard/pull"
	"github.com/natemarks/secret-hoard/tools"
	"github.com/rs/zerolog"
)

// Config is the configuration for the sh-pull command
type Config struct {
	Type        string
	Env         string
	Access      string
	Instance    string
	Database    string
	Warehouse   string
	CommonName  string
	Debug       bool
	Interactive bool
}

// GetConfig parses command line flags and returns a Config
// If no type is provided, enters interactive mode
func GetConfig() (config Config, err error) {
	typePtr := flag.String("type", "", "Secret type: rdspostgres, snowflake, ssl_certificate, jsondoc, text_file")
	envPtr := flag.String("env", "", "Environment")
	accessPtr := flag.String("access", "", "Access type (for jsondoc, text_file, snowflake, rdspostgres)")
	instancePtr := flag.String("instance", "", "Instance (for rdspostgres)")
	databasePtr := flag.String("database", "", "Database (for rdspostgres)")
	warehousePtr := flag.String("warehouse", "", "Warehouse (for snowflake)")
	commonnamePtr := flag.String("commonname", "", "Common name (for ssl_certificate)")
	debugPtr := flag.Bool("debug", false, "Enable Debug mode")
	flag.Parse()

	config.Type = *typePtr
	config.Env = *envPtr
	config.Access = *accessPtr
	config.Instance = *instancePtr
	config.Database = *databasePtr
	config.Warehouse = *warehousePtr
	config.CommonName = *commonnamePtr
	config.Debug = *debugPtr

	// If no type provided, enter interactive mode
	if config.Type == "" {
		config.Interactive = true
		return promptForConfig()
	}

	// Validate flag-based input
	if config.Env == "" {
		return config, fmt.Errorf(`environment is required

Usage:
  sh-pull -type=<type> -env=<env> [type-specific flags]

Example:
  sh-pull -type=jsondoc -env=dev -access=app-config

Tip: Run 'sh-pull' without flags for interactive mode`)
	}

	// Validate type-specific required flags
	switch config.Type {
	case "rdspostgres":
		if config.Instance == "" || config.Database == "" || config.Access == "" {
			return config, fmt.Errorf(`rdspostgres requires additional flags

Usage:
  sh-pull -type=rdspostgres -env=<env> -instance=<inst> -database=<db> -access=<access>

Example:
  sh-pull -type=rdspostgres -env=dev -instance=mydb -database=appdb -access=readonly`)
		}
	case "snowflake":
		if config.Warehouse == "" || config.Access == "" {
			return config, fmt.Errorf(`snowflake requires additional flags

Usage:
  sh-pull -type=snowflake -env=<env> -warehouse=<warehouse> -access=<access>

Example:
  sh-pull -type=snowflake -env=prod -warehouse=analytics -access=readonly`)
		}
	case "ssl_certificate":
		if config.CommonName == "" {
			return config, fmt.Errorf(`ssl_certificate requires -commonname flag

Usage:
  sh-pull -type=ssl_certificate -env=<env> -commonname=<domain>

Example:
  sh-pull -type=ssl_certificate -env=prod -commonname=example.com`)
		}
	case "jsondoc", "text_file":
		if config.Access == "" {
			return config, fmt.Errorf(`%s requires -access flag

Usage:
  sh-pull -type=%s -env=<env> -access=<access>

Example:
  sh-pull -type=%s -env=dev -access=app-config`, config.Type, config.Type, config.Type)
		}
	default:
		return config, fmt.Errorf(`unknown secret type: %q

Valid types:
  - rdspostgres
  - snowflake
  - ssl_certificate
  - jsondoc
  - text_file

Example:
  sh-pull -type=jsondoc -env=dev -access=app-config`, config.Type)
	}

	return config, nil
}

// promptForConfig prompts the user for configuration interactively
func promptForConfig() (config Config, err error) {
	secretTypes := []string{"rdspostgres", "snowflake", "ssl_certificate", "jsondoc", "text_file"}
	typeIndex := generate.PromptForChoice("Select secret type:", secretTypes)
	config.Type = secretTypes[typeIndex]

	// Common prompt for environment
	config.Env = generate.PromptForString("Environment: ")

	// Type-specific prompts
	switch config.Type {
	case "rdspostgres":
		config.Instance = generate.PromptForString("Instance: ")
		config.Database = generate.PromptForString("Database: ")
		config.Access = generate.PromptForString("Access: ")
	case "snowflake":
		config.Warehouse = generate.PromptForString("Warehouse: ")
		config.Access = generate.PromptForString("Access: ")
	case "ssl_certificate":
		config.CommonName = generate.PromptForString("Common Name: ")
	case "jsondoc", "text_file":
		config.Access = generate.PromptForString("Access: ")
	}

	config.Interactive = true
	return config, nil
}

// GetLogger returns a configured logger
func (c Config) GetLogger() zerolog.Logger {
	log := tools.TestLogger()
	if !c.Debug {
		log = log.Level(zerolog.InfoLevel)
	}
	return log
}

// ToPullMetadata converts Config to pull.PullMetadata
func (c Config) ToPullMetadata() pull.PullMetadata {
	return pull.PullMetadata{
		Type:       c.Type,
		Env:        c.Env,
		Access:     c.Access,
		Instance:   c.Instance,
		Database:   c.Database,
		Warehouse:  c.Warehouse,
		CommonName: c.CommonName,
	}
}
