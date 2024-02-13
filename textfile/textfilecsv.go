package textfile

import (
	"strings"

	"github.com/natemarks/secret-hoard/tools"
	"github.com/rs/zerolog"
)

// Record is the struct of the text file record
type Record struct {
	ResourceType string `json:"resourceType"` // record[0] : text_file
	Environment  string `json:"environment"`  // record[1] : dev, integration, staging, production
	Access       string `json:"access"`       // record[2] : access type provides by the secret
	FilePath     string `json:"filePath"`     // record[3] : /path/to/file
}

// CSVColumns Usage output describing the CSV structure
func (r Record) CSVColumns() string {
	result := "ResourceType,Environment,Access,FilePath\n"
	result += "text_file,testenv,my_file_type,/path/to/file\n"
	return result
}

// Sha256Sum returns the SHA256 sum of the certificate
func (r Record) Sha256Sum() (sum string, err error) {
	sum, err = tools.GetSHA256Sum(r.FilePath)
	return sum, err
}

// Contents returns the contents of the file
func (r Record) Contents() (contents string, err error) {
	contents, err = tools.ReadFileToString(r.FilePath)
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
			FilePath:     record[3],
		})

	}
	return result, nil
}
