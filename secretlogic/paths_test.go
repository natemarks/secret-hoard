package secretlogic

import "testing"

func TestBuildJSONDocPaths(t *testing.T) {
	metadata, contents := BuildJSONDocPaths("/tmp", "dev", "myaccess")

	expectedMetadata := "/tmp/jsondoc.dev.myaccess.metadata.json"
	expectedContents := "/tmp/jsondoc.dev.myaccess.contents.json"

	if metadata != expectedMetadata {
		t.Errorf("Expected metadata path %s, got %s", expectedMetadata, metadata)
	}
	if contents != expectedContents {
		t.Errorf("Expected contents path %s, got %s", expectedContents, contents)
	}
}

func TestBuildTextFilePaths(t *testing.T) {
	metadata, contents := BuildTextFilePaths("/home/user", "prod", "apikey")

	expectedMetadata := "/home/user/textfile.prod.apikey.metadata.json"
	expectedContents := "/home/user/textfile.prod.apikey.contents.txt"

	if metadata != expectedMetadata {
		t.Errorf("Expected metadata path %s, got %s", expectedMetadata, metadata)
	}
	if contents != expectedContents {
		t.Errorf("Expected contents path %s, got %s", expectedContents, contents)
	}
}

func TestBuildSSLCertPaths(t *testing.T) {
	metadata, cert, key := BuildSSLCertPaths("/var/secrets", "staging", "example.com")

	expectedMetadata := "/var/secrets/sslcert.staging.example.com.metadata.json"
	expectedCert := "/var/secrets/sslcert.staging.example.com.crt"
	expectedKey := "/var/secrets/sslcert.staging.example.com.key"

	if metadata != expectedMetadata {
		t.Errorf("Expected metadata path %s, got %s", expectedMetadata, metadata)
	}
	if cert != expectedCert {
		t.Errorf("Expected cert path %s, got %s", expectedCert, cert)
	}
	if key != expectedKey {
		t.Errorf("Expected key path %s, got %s", expectedKey, key)
	}
}

func TestBuildRDSPostgresPath(t *testing.T) {
	path := BuildRDSPostgresPath("/opt/data", "dev", "mydb", "appdb", "readonly")
	expected := "/opt/data/rdspostgres.dev.mydb.appdb.readonly.json"

	if path != expected {
		t.Errorf("Expected path %s, got %s", expected, path)
	}
}

func TestBuildSnowflakePath(t *testing.T) {
	path := BuildSnowflakePath("/tmp/secrets", "prod", "analytics", "admin")
	expected := "/tmp/secrets/snowflake.prod.analytics.admin.json"

	if path != expected {
		t.Errorf("Expected path %s, got %s", expected, path)
	}
}

func TestGetBaseName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "/home/user/jsondoc.dev.myaccess.metadata.json",
			expected: "/home/user/jsondoc.dev.myaccess",
		},
		{
			input:    "textfile.prod.apikey.metadata.json",
			expected: "textfile.prod.apikey",
		},
		{
			input:    "/var/secrets/sslcert.staging.example.com.metadata.json",
			expected: "/var/secrets/sslcert.staging.example.com",
		},
	}

	for _, tt := range tests {
		result := GetBaseName(tt.input)
		if result != tt.expected {
			t.Errorf("GetBaseName(%s) = %s, expected %s", tt.input, result, tt.expected)
		}
	}
}
