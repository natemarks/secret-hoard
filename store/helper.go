package store

import (
	"context"
	"fmt"
	"os"

	"github.com/rs/zerolog"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rds"

	"github.com/rs/zerolog/log"
)

func readFileToString(filepath string) (string, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// writeStringToFile writes a string to a file
func writeStringToFile(filename, contents string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Error().Err(err).Msgf("error closing file %s", filename)
		}
	}(file)

	_, err = file.WriteString(contents)
	if err != nil {
		return err
	}
	return nil

}

// GetRDSEndpoint returns the endpoint of the RDS instance
func GetRDSEndpoint(instanceID string) (string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return "", err
	}

	client := rds.NewFromConfig(cfg)

	params := &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: &instanceID,
	}

	resp, err := client.DescribeDBInstances(context.TODO(), params)
	if err != nil {
		return "", err
	}

	if len(resp.DBInstances) == 0 {
		return "", fmt.Errorf("no RDS instance found with ID: %s", instanceID)
	}

	return *resp.DBInstances[0].Endpoint.Address, nil
}

// GetTestLogger returns a logger for testing
func GetTestLogger() *zerolog.Logger {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	return &logger
}
