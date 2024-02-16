package tools

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"

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
	log = log.With().Str("aws_account_number", GetAWSAccountNumber()).Logger()
	log = log.Level(zerolog.InfoLevel)
	if c.Debug {
		log = log.Level(zerolog.DebugLevel)
	}
	return log
}

// CSVType returns the type of the CSV file
func (c Config) CSVType() (result string) {
	records, err := GetCSVRecordsFromFile(c.FilePath)
	if err != nil {
		panic(err)
	}
	for _, record := range records {
		// skip the header row
		if strings.ToLower(record[0]) == "resourcetype" {
			continue
		}
		return record[0]
	}
	panic(fmt.Errorf("no records found in CSV file: %s", c.FilePath))
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

// GetAWSAccountNumber retrieves the AWS account number associated with the credentials.
func GetAWSAccountNumber() string {
	// Load AWS SDK configuration
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(fmt.Errorf("unable to load SDK config, %v", err))
	}

	// Create an STS client
	stsClient := sts.NewFromConfig(cfg)

	// Call GetCallerIdentity API to get account details
	resp, err := stsClient.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
	if err != nil {
		panic(fmt.Errorf("unable to get caller identity, %v", err))
	}

	// Extract and return the account number
	return *resp.Account
}
