package dbinstance

import (
	"strconv"
	"strings"

	"github.com/natemarks/secret-hoard/tools"
	"github.com/rs/zerolog"
)

// Record is the struct of the rdspostgres record
type Record struct {
	ResourceType         string `json:"resourceType"`         // record[0] : rdspostgres
	Environment          string `json:"environment"`          // record[1] : dev, integration, staging, production
	Instance             string `json:"instance"`             // record[2] : RDS instance db identifier
	Database             string `json:"database"`             // record[3] : database name in the instance
	Access               string `json:"access"`               // record[4] : app_readwrite, app_readonly, etc.
	Password             string `json:"password"`             // record[5] : password
	Engine               string `json:"engine"`               // record[6] : ex. postgres
	Port                 int    `json:"port"`                 // record[7] : 5432
	DbInstanceIdentifier string `json:"dbInstanceIdentifier"` // record[8] : dbInstanceIdentifier
	Host                 string `json:"host"`                 // record[9] : host
	Username             string `json:"username"`             // record[10] : username
}

// CSVColumns Usage output describing the CSV structure
func (r Record) CSVColumns() string {
	result := "ResourceType,Environment,Instance,Database,Access,Password"
	result += ",Engine,Port,DbInstanceIdentifier,Host,Username\n"
	result += "rdspostgres,myenvironment,myinstance,mydatabase,mytype,mypassword"
	result += ",postgres,5432,mydbInstanceIdentifier,myhost,myusername\n"
	return result
}

// RecordsFromCSV reads a CSV file and returns a slice of SSLCertRecords
func RecordsFromCSV(csvFile string, log *zerolog.Logger) (result []Record, err error) {
	records, err := tools.GetCSVRecordsFromFile(csvFile)
	if err != nil {
		log.Error().Err(err).Msg("error reading store contents string")
		return result, err
	}

	for _, record := range records {
		// skip the header row
		if strings.ToLower(record[0]) == "resourcetype" {
			continue
		}
		port, err := strconv.Atoi(record[7])
		if err != nil {
			log.Error().Err(err).Msgf("error converting port %s to int", record[6])
			continue
		}
		result = append(result, Record{
			ResourceType:         record[0],
			Environment:          record[1],
			Instance:             record[2],
			Database:             record[3],
			Access:               record[4],
			Password:             record[5],
			Engine:               record[6],
			Port:                 port,
			DbInstanceIdentifier: record[8],
			Host:                 record[9],
			Username:             record[10],
		})

	}
	return result, nil
}
