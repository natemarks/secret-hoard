package tools

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

// FileExists checks if a file exists
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	return true
}

// ReadFileToString reads a file and returns the contents as a string
// read as byte slice then convert to string
func ReadFileToString(filepath string) (string, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// WriteStringToFile writes a string to a file.
func WriteStringToFile(content string, filename string) error {
	// Write the string content to the file
	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return err
	}

	return nil
}

// GetSHA256Sum returns the SHA256 sum of a file
func GetSHA256Sum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			return
		}
	}(file)

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	// Convert the hash to a hexadecimal string
	hashInBytes := hash.Sum(nil)
	sha256sum := hex.EncodeToString(hashInBytes)
	return sha256sum, nil
}

// GetCSVRecordsFromFile reads a CSV file and returns a slice of string slices
// This will include the header row which can be identified by the first element
// (record[0]) being "ResourceType"
func GetCSVRecordsFromFile(csvFile string) (records [][]string, err error) {
	csvContents, err := ReadFileToString(csvFile)
	if err != nil {
		return nil, err
	}
	reader := csv.NewReader(strings.NewReader(csvContents))
	records, err = reader.ReadAll()
	if err != nil {
		return nil, err
	}
	return records, nil
}

// CheckSha256Sum checks if the SHA256 sum of a file matches the expected value
func CheckSha256Sum(filePath string, expected string) (err error) {
	sha256Sum, err := GetSHA256Sum(filePath)
	if err != nil {
		return err
	}
	if sha256Sum != expected {
		return fmt.Errorf("sha256sum mismatch: expected %s, got %s", expected, sha256Sum)
	}
	return nil
}
