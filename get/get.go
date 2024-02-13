package get

import (
	"encoding/json"
	"fmt"

	"github.com/natemarks/secret-hoard/textfile"

	"github.com/natemarks/secret-hoard/jsondoc"

	"github.com/natemarks/secret-hoard/sslcert"
	"github.com/natemarks/secret-hoard/tools"
	"github.com/rs/zerolog"
)

// DownloadSecret returns a secret from the secret store
// The execution varies depending on the resources type:
// rdspostgres: Download the data required for a connection string to json file
// snowflake: Download the data required for a connection string to json file
// jsondoc: Download the json file
// ssl_certificate: Download the certificate and private key files to filePath.crt  and filePath.key files
func DownloadSecret(secretID, filePath string, log *zerolog.Logger) (err error) {
	resourceType, err := tools.GetResourceTypeFromSecretID(secretID)
	if err != nil {
		return err
	}
	// use switch to handle different resource types
	switch resourceType {
	case "rdspostgres":
		return DownloadValue(secretID, filePath, log)
	case "snowflake":
		return DownloadValue(secretID, filePath, log)
	case "jsondoc":
		return DownloadJSONContents(secretID, filePath, log)
	case "ssl_certificate":
		return DownloadCertAndKeyFiles(secretID, filePath, log)
	case "text_file":
		return DownloadTextContents(secretID, filePath, log)
	default:
		return fmt.Errorf("resource type not supported: %s", resourceType)
	}
}

// DownloadValue download the secret value to a JSON file
func DownloadValue(secretID string, filePath string, log *zerolog.Logger) (err error) {
	log.Info().Msgf("getting secret value: %s", secretID)
	secretValue, err := tools.GetSecretValue(secretID)
	if err != nil {
		log.Error().Err(err).Msgf("error getting secret value: %sd", secretID)
	}
	log.Debug().Msgf("got secret value: %s", secretID)
	// do not unmarshall the value. download the whole JSON value to the file contents for rdspostgres
	log.Debug().Msgf("writing secret data to file: %s", filePath)
	err = tools.WriteStringToFile(secretValue, filePath)
	if err != nil {
		log.Error().Err(err).Msgf("error writing secret data to file: %s", filePath)
		return err
	}
	log.Debug().Msgf("wrote secret data to file: %s", filePath)
	return nil
}

// DownloadJSONContents download Data.JSONContents to a file and use Data.JSONSha256Sum to verify integrity
func DownloadJSONContents(secretID string, filePath string, log *zerolog.Logger) (err error) {
	var result jsondoc.Data
	log.Info().Msgf("getting secret value: %s", secretID)
	secretValue, err := tools.GetSecretValue(secretID)
	if err != nil {
		log.Error().Err(err).Msgf("error getting secret value: %sd", secretID)
	}
	log.Debug().Msgf("got secret value: %s", secretID)

	// unmarshall the value to get the certificate and private key file contents
	err = json.Unmarshal([]byte(secretValue), &result)
	if err != nil {
		log.Error().Err(err).Msgf("error unmarshalling secret value: %s", secretID)
		return err
	}
	log.Debug().Msgf("unmarshalled secret value: %s", secretID)

	err = tools.WriteStringToFile(result.JSONContents, filePath)
	if err != nil {
		log.Error().Err(err).Msgf("error writing json document to file: %s", filePath)
		return err
	}
	log.Debug().Msgf("wrote valid json document to file: %s", filePath)

	err = tools.CheckSha256Sum(filePath, result.JSONSha256Sum)
	if err != nil {
		log.Error().Err(err).Msgf("error checking sha256sum of json document: %s", filePath)
		return err
	}
	log.Debug().Msgf("valid sha256sum (%s): %s", result.JSONSha256Sum, filePath)

	return nil
}

// DownloadTextContents download Data.Contents to a file and use Data.Sha256Sum to verify integrity
func DownloadTextContents(secretID string, filePath string, log *zerolog.Logger) (err error) {
	var result textfile.Data
	log.Info().Msgf("getting secret value: %s", secretID)
	secretValue, err := tools.GetSecretValue(secretID)
	if err != nil {
		log.Error().Err(err).Msgf("error getting secret value: %sd", secretID)
	}
	log.Debug().Msgf("got secret value: %s", secretID)

	// unmarshall the value to get the certificate and private key file contents
	err = json.Unmarshal([]byte(secretValue), &result)
	if err != nil {
		log.Error().Err(err).Msgf("error unmarshalling secret value: %s", secretID)
		return err
	}
	log.Debug().Msgf("unmarshalled secret value: %s", secretID)

	err = tools.WriteStringToFile(result.Contents, filePath)
	if err != nil {
		log.Error().Err(err).Msgf("error writing json document to file: %s", filePath)
		return err
	}
	log.Debug().Msgf("wrote valid json document to file: %s", filePath)

	err = tools.CheckSha256Sum(filePath, result.Sha256Sum)
	if err != nil {
		log.Error().Err(err).Msgf("error checking sha256sum of json document: %s", filePath)
		return err
	}
	log.Debug().Msgf("valid sha256sum (%s): %s", result.Sha256Sum, filePath)

	return nil
}

// DownloadCertAndKeyFiles download the certificate and private key files to filePath.crt  and filePath.key files
func DownloadCertAndKeyFiles(secretID string, filePath string, log *zerolog.Logger) (err error) {
	var result sslcert.Data
	log.Info().Msgf("getting snowflake secret value: %s", secretID)
	secretValue, err := tools.GetSecretValue(secretID)
	if err != nil {
		log.Error().Err(err).Msgf("error getting secret value: %sd", secretID)
	}
	log.Debug().Msgf("got secret value: %s", secretID)

	// unmarshall the value to get the certificate and private key file contents
	err = json.Unmarshal([]byte(secretValue), &result)
	if err != nil {
		log.Error().Err(err).Msgf("error unmarshalling secret value: %s", secretID)
		return err
	}
	log.Debug().Msgf("unmarshalled secret value: %s", secretID)

	err = tools.WriteStringToFile(result.Certificate, filePath+".crt")
	if err != nil {
		log.Error().Err(err).Msgf("error writing certificate to file: %s", filePath+".crt")
		return err
	}
	log.Debug().Msgf("wrote certificate to file: %s", filePath+".crt")
	err = tools.CheckSha256Sum(filePath+".crt", result.CertificateSha256)
	if err != nil {
		log.Error().Err(err).Msgf("error checking sha256sum of certificate: %s", filePath+".crt")
		return err

	}
	log.Debug().Msgf("valid sha256sum (%s): %s", result.CertificateSha256, filePath+".crt")

	err = tools.WriteStringToFile(result.PrivateKey, filePath+".key")
	if err != nil {
		log.Error().Err(err).Msgf("error writing private key to file: %s", filePath+".key")
		return err
	}
	log.Debug().Msgf("wrote private key  to file: %s", filePath+".key")
	err = tools.CheckSha256Sum(filePath+".key", result.PrivateKeySha256)
	if err != nil {
		log.Error().Err(err).Msgf("error checking sha256sum of private key: %s", filePath+".key")
		return err
	}
	log.Debug().Msgf("valid sha256sum (%s): %s", result.PrivateKeySha256, filePath+".key")
	return nil
}
