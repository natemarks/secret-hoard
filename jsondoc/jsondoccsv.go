package jsondoc

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/natemarks/secret-hoard/tools"
	"github.com/rs/zerolog"
)

// Record is the struct of the SSL certificate record
type Record struct {
	ResourceType string `json:"resourceType"` // record[0] : json_document
	Environment  string `json:"environment"`  // record[1] : dev, integration, staging, production
	Access       string `json:"access"`       // record[2] : access type provides by the secret
	JSONFilePath string `json:"jsonFilePath"` // record[3] : /path/to/file.json
}

// CSVColumns Usage output describing the CSV structure
func (r Record) CSVColumns() string {
	result := "ResourceType,Environment,Access,JSONFilePath\n"
	result += "jsondoc,testenv,my_endpoints,/path/to/file.json\n"
	return result
}

// Sha256Sum returns the SHA256 sum of the certificate
func (r Record) Sha256Sum() (sum string, err error) {
	sum, err = tools.GetSHA256Sum(r.JSONFilePath)
	return sum, err
}

// Unmarshall returns the contents of the json file as a map
func (r Record) Unmarshall() (result map[string]interface{}, err error) {
	content, err := os.ReadFile(r.JSONFilePath)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(content, &result)
	return result, err
}

// IsValidJSON returns true if the json file is valid
func (r Record) IsValidJSON() bool {
	_, err := r.Unmarshall()
	if err != nil {
		return false
	}
	return true
}

// JSONContents returns the contents of the private key file
func (r Record) JSONContents() (contents string, err error) {
	contents, err = tools.ReadFileToString(r.JSONFilePath)
	return contents, err
}

// RecordsFromCSV reads a CSV file and returns a slice of Records
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
			Access:       record[2],
			JSONFilePath: record[3],
		})

	}
	return result, nil
}
