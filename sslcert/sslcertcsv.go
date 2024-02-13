package sslcert

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/natemarks/secret-hoard/tools"
	"github.com/rs/zerolog"
)

// Record is the struct of the SSL certificate record
type Record struct {
	ResourceType    string `json:"resourceType"`    // record[0] : ssl_certificate
	Environment     string `json:"environment"`     // record[1] : dev, integration, staging, production
	CommonName      string `json:"commonName"`      // record[2] : \*.my.domain.com | server.my.domain.com
	CertificateFile string `json:"certificateFile"` // record[3] : /path/to/certificate.crt
	PrivateKeyFile  string `json:"privateKeyFile"`  // record[4] : /path/to/private.key
}

// CSVColumns Usage output describing the CSV structure
func (scr Record) CSVColumns() string {
	result := "ResourceType,Environment,CommonName,CertificateFile,PrivateKeyFile\n"
	result += "ssl_certificate,testenv,my.domain.com,/path/to/certificate.crt,/path/to/private.key\n"
	return result
}

// CertificateModulus returns the modulus of the certificate
func (scr Record) CertificateModulus() (modulus string, err error) {
	// Read the certificate file
	certPEM, err := os.ReadFile(scr.CertificateFile)
	if err != nil {
		return "", err
	}

	// Decode PEM-encoded certificate
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return "", fmt.Errorf("failed to decode certificate PEM")
	}

	// Parse certificate
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", err
	}

	// Extract RSA public key from certificate
	rsaPublicKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("certificate public key is not RSA")
	}

	return rsaPublicKey.N.String(), nil
}

// PrivateKeyModulus returns the modulus of the private key
func (scr Record) PrivateKeyModulus() (modulus string, err error) {
	// Read the private key file
	privateKeyPEM, err := os.ReadFile(scr.PrivateKeyFile)
	if err != nil {
		return "", err
	}

	// Decode PEM-encoded private key
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return "", fmt.Errorf("failed to decode private key PEM")
	}

	// Parse private key
	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}

	// Extract RSA private key from parsed private key
	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("private key is not RSA")
	}
	return rsaPrivateKey.N.String(), nil
}

// CertificateSha256Sum returns the SHA256 sum of the certificate
func (scr Record) CertificateSha256Sum() (sum string, err error) {
	sum, err = tools.GetSHA256Sum(scr.CertificateFile)
	return sum, err
}

// PrivateKeySha256Sum returns the SHA256 sum of the certificate
func (scr Record) PrivateKeySha256Sum() (sum string, err error) {
	sum, err = tools.GetSHA256Sum(scr.PrivateKeyFile)
	return sum, err
}

// Expiration returns the expiration date of the certificate in ISO 3339 format
func (scr Record) Expiration() (expiration string, err error) {
	certData, err := os.ReadFile(scr.CertificateFile)
	if err != nil {
		return expiration, err
	}

	block, _ := pem.Decode(certData)
	if block == nil {
		return expiration, fmt.Errorf("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return expiration, err
	}

	return cert.NotAfter.Format(time.RFC3339), nil
}

// CertificateContents returns the contents of the certificate file
func (scr Record) CertificateContents() (contents string, err error) {
	contents, err = tools.ReadFileToString(scr.CertificateFile)
	return contents, err
}

// PrivateKeyContents returns the contents of the private key file
func (scr Record) PrivateKeyContents() (contents string, err error) {
	contents, err = tools.ReadFileToString(scr.PrivateKeyFile)
	return contents, err
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
		result = append(result, Record{
			ResourceType:    record[0],
			Environment:     record[1],
			CommonName:      record[2],
			CertificateFile: record[3],
			PrivateKeyFile:  record[4],
		})

	}
	return result, nil
}
