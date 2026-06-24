package generate

import (
	"encoding/json"
	"fmt"

	"github.com/natemarks/secret-hoard/tools"
)

const placeholder = tools.PlaceholderValue

// SecretFiles creates local file scaffolding for a new secret
func SecretFiles(log *tools.Logger) error {
	secretTypes := []string{"rdspostgres", "snowflake", "ssl_certificate", "jsondoc", "text_file"}
	typeIndex := PromptForChoice("Select secret type:", secretTypes)
	secretType := secretTypes[typeIndex]

	environment := PromptForString("Environment: ")

	switch secretType {
	case "rdspostgres":
		return generateRDSPostgres(environment, log)
	case "snowflake":
		return generateSnowflake(environment, log)
	case "ssl_certificate":
		return generateSSLCert(environment, log)
	case "jsondoc":
		return generateJSONDoc(environment, log)
	case "text_file":
		return generateTextFile(environment, log)
	default:
		return fmt.Errorf("unknown secret type: %s", secretType)
	}
}

func generateRDSPostgres(environment string, log *tools.Logger) error {
	instance := PromptForString("Instance: ")
	database := PromptForString("Database: ")
	access := PromptForString("Access: ")

	workingDir, err := tools.GetWorkingDir()
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("%s/rdspostgres.%s.%s.%s.%s.json", workingDir, environment, instance, database, access)

	data := map[string]interface{}{
		"metadata": map[string]string{
			"resourceType": "rdspostgres",
			"environment":  environment,
			"instance":     instance,
			"database":     database,
			"access":       access,
		},
		"data": map[string]interface{}{
			"password":             placeholder,
			"engine":               placeholder,
			"port":                 5432,
			"dbInstanceIdentifier": placeholder,
			"host":                 placeholder,
			"username":             placeholder,
		},
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	err = tools.WriteStringToFile(string(jsonBytes), filename)
	if err != nil {
		return err
	}

	log.Info("Generated file: %s", filename)
	fmt.Printf("\nGenerated files:\n  - %s\n\n", filename)
	fmt.Printf("Next steps:\n")
	fmt.Printf("  1. Edit the file and replace REPLACE-ME values\n")
	fmt.Printf("  2. Run: sh-push -metadata=%s\n", filename)

	return nil
}

func generateSnowflake(environment string, log *tools.Logger) error {
	warehouse := PromptForString("Warehouse: ")
	access := PromptForString("Access: ")

	workingDir, err := tools.GetWorkingDir()
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("%s/snowflake.%s.%s.%s.json", workingDir, environment, warehouse, access)

	data := map[string]interface{}{
		"metadata": map[string]string{
			"resourceType": "snowflake",
			"environment":  environment,
			"warehouse":    warehouse,
			"access":       access,
		},
		"data": map[string]interface{}{
			"password":    placeholder,
			"accountName": placeholder,
			"warehouse":   warehouse,
			"username":    placeholder,
		},
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	err = tools.WriteStringToFile(string(jsonBytes), filename)
	if err != nil {
		return err
	}

	log.Info("Generated file: %s", filename)
	fmt.Printf("\nGenerated files:\n  - %s\n\n", filename)
	fmt.Printf("Next steps:\n")
	fmt.Printf("  1. Edit the file and replace REPLACE-ME values\n")
	fmt.Printf("  2. Run: sh-push -metadata=%s\n", filename)

	return nil
}

func generateSSLCert(environment string, log *tools.Logger) error {
	commonName := PromptForString("Common Name: ")

	workingDir, err := tools.GetWorkingDir()
	if err != nil {
		return err
	}

	metadataFile := fmt.Sprintf("%s/sslcert.%s.%s.metadata.json", workingDir, environment, commonName)
	certFile := fmt.Sprintf("%s/sslcert.%s.%s.crt", workingDir, environment, commonName)
	keyFile := fmt.Sprintf("%s/sslcert.%s.%s.key", workingDir, environment, commonName)

	metadata := map[string]string{
		"resourceType":      "ssl_certificate",
		"environment":       environment,
		"commonName":        commonName,
		"expirationDate":    placeholder,
		"modulus":           placeholder,
		"certificateSha256": placeholder,
		"privateKeySha256":  placeholder,
	}

	jsonBytes, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	err = tools.WriteStringToFile(string(jsonBytes), metadataFile)
	if err != nil {
		return err
	}

	err = tools.WriteStringToFile("", certFile)
	if err != nil {
		return err
	}

	err = tools.WriteStringToFile("", keyFile)
	if err != nil {
		return err
	}

	log.Info("Generated files: %s, %s, %s", metadataFile, certFile, keyFile)
	fmt.Printf("\nGenerated files:\n  - %s\n  - %s\n  - %s\n\n", metadataFile, certFile, keyFile)
	fmt.Printf("Next steps:\n")
	fmt.Printf("  1. Add certificate content to %s\n", certFile)
	fmt.Printf("  2. Add private key content to %s\n", keyFile)
	fmt.Printf("  3. Run: sh-push -metadata=%s\n", metadataFile)

	return nil
}

func generateJSONDoc(environment string, log *tools.Logger) error {
	access := PromptForString("Access: ")

	workingDir, err := tools.GetWorkingDir()
	if err != nil {
		return err
	}

	metadataFile := fmt.Sprintf("%s/jsondoc.%s.%s.metadata.json", workingDir, environment, access)
	contentsFile := fmt.Sprintf("%s/jsondoc.%s.%s.contents.json", workingDir, environment, access)

	metadata := map[string]string{
		"resourceType":  "jsondoc",
		"environment":   environment,
		"access":        access,
		"JSONSha256Sum": placeholder,
	}

	jsonBytes, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	err = tools.WriteStringToFile(string(jsonBytes), metadataFile)
	if err != nil {
		return err
	}

	err = tools.WriteStringToFile("{}", contentsFile)
	if err != nil {
		return err
	}

	log.Info("Generated files: %s, %s", metadataFile, contentsFile)
	fmt.Printf("\nGenerated files:\n  - %s\n  - %s\n\n", metadataFile, contentsFile)
	fmt.Printf("Next steps:\n")
	fmt.Printf("  1. Edit the %s file with your JSON document\n", contentsFile)
	fmt.Printf("  2. Run: sh-push -metadata=%s\n", metadataFile)

	return nil
}

func generateTextFile(environment string, log *tools.Logger) error {
	access := PromptForString("Access: ")

	workingDir, err := tools.GetWorkingDir()
	if err != nil {
		return err
	}

	metadataFile := fmt.Sprintf("%s/textfile.%s.%s.metadata.json", workingDir, environment, access)
	contentsFile := fmt.Sprintf("%s/textfile.%s.%s.contents.txt", workingDir, environment, access)

	metadata := map[string]string{
		"resourceType": "text_file",
		"environment":  environment,
		"access":       access,
		"sha256Sum":    placeholder,
	}

	jsonBytes, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	err = tools.WriteStringToFile(string(jsonBytes), metadataFile)
	if err != nil {
		return err
	}

	err = tools.WriteStringToFile("", contentsFile)
	if err != nil {
		return err
	}

	log.Info("Generated files: %s, %s", metadataFile, contentsFile)
	fmt.Printf("\nGenerated files:\n  - %s\n  - %s\n\n", metadataFile, contentsFile)
	fmt.Printf("Next steps:\n")
	fmt.Printf("  1. Edit the %s file with your text content\n", contentsFile)
	fmt.Printf("  2. Run: sh-push -metadata=%s\n", metadataFile)

	return nil
}
