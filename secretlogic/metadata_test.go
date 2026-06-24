package secretlogic

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "update golden files")

func TestBuildMetadataMap(t *testing.T) {
	tests := []struct {
		name       string
		secretType string
		metadata   map[string]string
		goldenFile string
		wantErr    bool
	}{
		// JSONDOC
		{
			name:       "jsondoc_dev_app",
			secretType: "jsondoc",
			metadata:   map[string]string{"environment": "dev", "access": "app"},
			goldenFile: "testdata/metadata/jsondoc_dev_app.golden.json",
		},
		{
			name:       "jsondoc_prod_config",
			secretType: "jsondoc",
			metadata:   map[string]string{"environment": "prod", "access": "config"},
			goldenFile: "testdata/metadata/jsondoc_prod_config.golden.json",
		},

		// TEXT_FILE
		{
			name:       "textfile_staging_keys",
			secretType: "text_file",
			metadata:   map[string]string{"environment": "staging", "access": "keys"},
			goldenFile: "testdata/metadata/textfile_staging_keys.golden.json",
		},

		// SSL_CERTIFICATE
		{
			name:       "sslcert_prod_example",
			secretType: "ssl_certificate",
			metadata:   map[string]string{"environment": "prod", "commonname": "api.example.com"},
			goldenFile: "testdata/metadata/sslcert_prod_example.golden.json",
		},
		{
			name:       "sslcert_wildcard",
			secretType: "ssl_certificate",
			metadata:   map[string]string{"environment": "staging", "commonname": "*.example.com"},
			goldenFile: "testdata/metadata/sslcert_wildcard.golden.json",
		},

		// SNOWFLAKE
		{
			name:       "snowflake_prod_analytics",
			secretType: "snowflake",
			metadata: map[string]string{
				"environment": "prod",
				"warehouse":   "analytics",
				"access":      "readonly",
			},
			goldenFile: "testdata/metadata/snowflake_prod_analytics.golden.json",
		},

		// RDSPOSTGRES
		{
			name:       "rdspostgres_dev_db1",
			secretType: "rdspostgres",
			metadata: map[string]string{
				"environment": "dev",
				"instance":    "db1",
				"database":    "myapp",
				"access":      "readonly",
			},
			goldenFile: "testdata/metadata/rdspostgres_dev_db1.golden.json",
		},

		// ERROR cases
		{
			name:       "missing_environment",
			secretType: "jsondoc",
			metadata:   map[string]string{"access": "app"},
			wantErr:    true,
		},
		{
			name:       "unknown_type",
			secretType: "invalid",
			metadata:   map[string]string{"environment": "dev"},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildMetadataMap(tt.secretType, tt.metadata)

			if (err != nil) != tt.wantErr {
				t.Errorf("BuildMetadataMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Convert to formatted JSON for comparison
			gotJSON, err := json.MarshalIndent(got, "", "  ")
			if err != nil {
				t.Fatalf("json.MarshalIndent() error = %v", err)
			}

			// Read golden file
			wantJSON, err := os.ReadFile(tt.goldenFile)
			if err != nil {
				// If golden file doesn't exist, create it if -update flag is set
				if os.IsNotExist(err) {
					if *update {
						t.Logf("Creating golden file: %s", tt.goldenFile)
						if err := os.MkdirAll(filepath.Dir(tt.goldenFile), 0755); err != nil {
							t.Fatalf("Failed to create directory: %v", err)
						}
						if err := os.WriteFile(tt.goldenFile, gotJSON, 0644); err != nil {
							t.Fatalf("Failed to create golden file: %v", err)
						}
						return
					}
					t.Fatalf("Golden file does not exist: %s (run with -update to create)", tt.goldenFile)
				}
				t.Fatalf("Failed to read golden file: %v", err)
			}

			// Compare
			if string(gotJSON) != string(wantJSON) {
				t.Errorf("BuildMetadataMap() mismatch\nGot:\n%s\n\nWant:\n%s\n", gotJSON, wantJSON)

				// Optionally update golden file if -update flag is set
				if *update {
					t.Logf("Updating golden file: %s", tt.goldenFile)
					if err := os.WriteFile(tt.goldenFile, gotJSON, 0644); err != nil {
						t.Fatalf("Failed to update golden file: %v", err)
					}
				}
			}
		})
	}
}

func TestRequiredMetadataFields(t *testing.T) {
	tests := []struct {
		name       string
		secretType string
		want       []string
		wantErr    bool
	}{
		{
			name:       "jsondoc",
			secretType: "jsondoc",
			want:       []string{"environment", "access"},
		},
		{
			name:       "text_file",
			secretType: "text_file",
			want:       []string{"environment", "access"},
		},
		{
			name:       "ssl_certificate",
			secretType: "ssl_certificate",
			want:       []string{"environment", "commonname"},
		},
		{
			name:       "snowflake",
			secretType: "snowflake",
			want:       []string{"environment", "warehouse", "access"},
		},
		{
			name:       "rdspostgres",
			secretType: "rdspostgres",
			want:       []string{"environment", "instance", "database", "access"},
		},
		{
			name:       "unknown_type",
			secretType: "invalid",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RequiredMetadataFields(tt.secretType)

			if (err != nil) != tt.wantErr {
				t.Errorf("RequiredMetadataFields() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("RequiredMetadataFields() length = %d, want %d", len(got), len(tt.want))
				return
			}

			for i, field := range tt.want {
				if got[i] != field {
					t.Errorf("RequiredMetadataFields()[%d] = %s, want %s", i, got[i], field)
				}
			}
		})
	}
}
