package types

import "github.com/rs/zerolog"

type Secret interface {
	SecretID() string
	Map() map[string]string
	Reachable() bool
	CreateAWSSecret() error
	UpdateAWSSecret() error
}

type CSVReader interface {
	ReadCSV(filename string, log *zerolog.Logger) ([]Secret, error)
}
