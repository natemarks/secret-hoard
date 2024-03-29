package get

import (
	"testing"
	"time"

	"github.com/natemarks/secret-hoard/jsondoc"
	"github.com/natemarks/secret-hoard/tools"
)

func TestJSONDocSecret(t *testing.T) {
	var record = jsondoc.Record{
		ResourceType: "jsondoc",
		Environment:  "testenv",
		Access:       "some_json_access_type",
		JSONFilePath: "../examples/jsondoc_example.json",
	}
	downloadFile := t.TempDir() + "/jsondoc_test.json"
	log := tools.TestLogger()
	secret, err := jsondoc.FromCSVRecord(record, &log)
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
