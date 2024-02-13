package snowflake

import (
	"strings"

	"github.com/natemarks/secret-hoard/tools"
	"github.com/rs/zerolog"
)

// Record is the struct of the snowflake record
type Record struct {
	ResourceType string `json:"resourceType"` // record[0] : rdspostgres
	Environment  string `json:"environment"`  // record[1] : dev, integration, staging, production
	Warehouse    string `json:"warehouse"`    // record[2] : snowflake warehouse
	Access       string `json:"access"`       // record[3] : app_readwrite, app_readonly, etc.
	AccountName  string `json:"accountName"`  // record[4] : snowflake account name
	Username     string `json:"username"`     // record[5] : username
	Password     string `json:"password"`     // record[6] : password

}

// CSVColumns Usage output describing the CSV structure
func (scr Record) CSVColumns() string {
	result := "ResourceType,Environment,Warehouse,Access,AccountName,Username,Password\n"
	result += "snowflake,myenvironment,mywarehouse,mytype,myAccountname,myusername,mypassword\n"
	return result
}

// RecordsFromCSV reads a CSV file and returns a slice of Record
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

		result = append(result, Record{
			ResourceType: record[0],
			Environment:  record[1],
			Warehouse:    record[2],
			Access:       record[3],
			AccountName:  record[4],
			Username:     record[5],
			Password:     record[6],
		})

	}
	return result, nil
}
