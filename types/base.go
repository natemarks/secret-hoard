package types

import "github.com/rs/zerolog"

// Secret is an interface for a secret
type Secret interface {
	SecretID() string
	Map() map[string]string
	Reachable() bool
	CreateAWSSecret() error
	UpdateAWSSecret() error
}

// CSVReader is an interface for reading CSV files
type CSVReader interface {
	ReadCSV(filename string, log *zerolog.Logger) ([]Secret, error)
}
