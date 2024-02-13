package get

import (
	"os"
	"testing"
	"time"

	"github.com/natemarks/secret-hoard/dbinstance"

	"github.com/natemarks/secret-hoard/tools"
	"github.com/natemarks/secret-hoard/version"

	"github.com/rs/zerolog"
)

func TestDBInstanceSecret(t *testing.T) {
	var record = dbinstance.Record{
		ResourceType:         "rdspostgres",
		Environment:          "testenv",
		Instance:             "myinstance",
		Database:             "mydb",
		Access:               "mytype",
		Password:             "password",
		Engine:               "postgres",
		Port:                 5432,
		DbInstanceIdentifier: "dbInstanceIdentifier",
		Host:                 "host",
		Username:             "username",
	}
	downloadFile := t.TempDir() + "/dbinstance_test.json"
	log := zerolog.New(os.Stdout).With().Str("version", version.Version).Timestamp().Logger()
	log = log.Level(zerolog.DebugLevel)
	secret, err := dbinstance.FromCSVRecord(record, &log)
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
