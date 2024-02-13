package get

import (
	"os"
	"testing"
	"time"

	"github.com/natemarks/secret-hoard/sslcert"

	"github.com/natemarks/secret-hoard/tools"
	"github.com/natemarks/secret-hoard/version"
	"github.com/rs/zerolog"
)

func TestSSLCertSecret(t *testing.T) {
	var record = sslcert.Record{
		ResourceType:    "ssl_certificate",
		Environment:     "testenv",
		CommonName:      "my.domain.com",
		CertificateFile: "../examples/certificate.crt",
		PrivateKeyFile:  "../examples/private_key.key",
	}
	// this will creaste two files in the temp directory: sslcert_test.crt and sslcert_test.key
	downloadFile := t.TempDir() + "/sslcert_test"
	log := zerolog.New(os.Stdout).With().Str("version", version.Version).Timestamp().Logger()
	log = log.Level(zerolog.DebugLevel)
	secret, err := sslcert.FromCSVRecord(record, &log)
	if err != nil {
		t.Errorf("FromCSVRecord() error = %v", err)
	}
	if secret.Exists(&log) {
		tools.DeleteSecrets([]string{secret.Metadata.SecretID()})
		t.Logf("waiting 30 seconds for secret deletion: %s", secret.Metadata.SecretID())
		time.Sleep(30 * time.Second)
	}
	t.Logf("creating secret: %s", secret.Metadata.SecretID())
	secret.Create(&log)
	t.Logf("updating secret - overwrite FALSE: %s", secret.Metadata.SecretID())
	secret.Update(false, &log)
	t.Logf("updating secret - overwrite TRUE: %s", secret.Metadata.SecretID())
	secret.Update(true, &log)
	t.Logf("downloading secret  (%s) to %s", secret.Metadata.SecretID(), downloadFile)
	err = DownloadSecret(secret.Metadata.SecretID(), downloadFile, &log)
	if err != nil {
		t.Errorf("DownloadSecret() error = %v", err)
	}
	tools.DeleteSecrets([]string{secret.Metadata.SecretID()})
}
