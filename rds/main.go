package rds

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rds"
)

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
