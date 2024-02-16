package get

import (
	"testing"
	"time"

	"github.com/natemarks/secret-hoard/rdspostgres"

	"github.com/natemarks/secret-hoard/tools"
)

func TestDBInstanceSecret(t *testing.T) {
	var record = rdspostgres.Record{
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
	downloadFile := t.TempDir() + "/rdspostgres_test.json"
	log := tools.TestLogger()
	secret, err := rdspostgres.FromCSVRecord(record, &log)
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
