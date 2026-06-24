package secretlogic

import (
	"reflect"
	"testing"
)

func TestBuildBaseFilename(t *testing.T) {
	tests := []struct {
		name       string
		secretType string
		metadata   map[string]string
		want       string
		wantErr    bool
	}{
		// JSONDOC
		{
			name:       "jsondoc_basic",
			secretType: "jsondoc",
			metadata:   map[string]string{"environment": "dev", "access": "app"},
			want:       "jsondoc.dev.app",
		},
		{
			name:       "jsondoc_prod",
			secretType: "jsondoc",
			metadata:   map[string]string{"environment": "prod", "access": "idemia"},
			want:       "jsondoc.prod.idemia",
		},

		// TEXT_FILE
		{
			name:       "textfile_basic",
			secretType: "text_file",
			metadata:   map[string]string{"environment": "staging", "access": "config"},
			want:       "text_file.staging.config",
		},

		// SNOWFLAKE
		{
			name:       "snowflake_basic",
			secretType: "snowflake",
			metadata: map[string]string{
				"environment": "prod",
				"warehouse":   "analytics",
				"access":      "developer",
			},
			want: "snowflake.prod.analytics.developer",
		},

		// RDSPOSTGRES
		{
			name:       "rdspostgres_basic",
			secretType: "rdspostgres",
			metadata: map[string]string{
				"environment": "dev",
				"instance":    "db1",
				"database":    "myapp",
				"access":      "readonly",
			},
			want: "rdspostgres.dev.db1.myapp.readonly",
		},

		// SSL_CERTIFICATE
		{
			name:       "sslcert_basic",
			secretType: "ssl_certificate",
			metadata:   map[string]string{"environment": "prod", "commonname": "example.com"},
			want:       "ssl_certificate.prod.example.com",
		},
		{
			name:       "sslcert_wildcard",
			secretType: "ssl_certificate",
			metadata:   map[string]string{"environment": "staging", "commonname": "*.example.com"},
			want:       "ssl_certificate.staging.*.example.com",
		},

		// ERROR cases
		{
			name:       "unknown_type",
			secretType: "invalid",
			metadata:   map[string]string{"environment": "dev"},
			wantErr:    true,
		},
		{
			name:       "missing_environment",
			secretType: "jsondoc",
			metadata:   map[string]string{"access": "app"},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildBaseFilename(tt.secretType, tt.metadata)

			if (err != nil) != tt.wantErr {
				t.Errorf("BuildBaseFilename() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("BuildBaseFilename() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFilesToGenerate(t *testing.T) {
	tests := []struct {
		name       string
		secretType string
		metadata   map[string]string
		want       []string
		wantErr    bool
	}{
		// JSONDOC - split files
		{
			name:       "jsondoc_files",
			secretType: "jsondoc",
			metadata:   map[string]string{"environment": "dev", "access": "app"},
			want: []string{
				"jsondoc.dev.app.metadata.json",
				"jsondoc.dev.app.contents.json",
			},
		},

		// TEXT_FILE - split files
		{
			name:       "textfile_files",
			secretType: "text_file",
			metadata:   map[string]string{"environment": "prod", "access": "config"},
			want: []string{
				"text_file.prod.config.metadata.json",
				"text_file.prod.config.contents.txt",
			},
		},

		// SSL_CERTIFICATE - three files
		{
			name:       "sslcert_files",
			secretType: "ssl_certificate",
			metadata:   map[string]string{"environment": "staging", "commonname": "example.com"},
			want: []string{
				"ssl_certificate.staging.example.com.metadata.json",
				"ssl_certificate.staging.example.com.crt",
				"ssl_certificate.staging.example.com.key",
			},
		},

		// RDSPOSTGRES - single file
		{
			name:       "rdspostgres_single",
			secretType: "rdspostgres",
			metadata: map[string]string{
				"environment": "dev",
				"instance":    "db1",
				"database":    "app",
				"access":      "ro",
			},
			want: []string{
				"rdspostgres.dev.db1.app.ro.json",
			},
		},

		// SNOWFLAKE - single file
		{
			name:       "snowflake_single",
			secretType: "snowflake",
			metadata: map[string]string{
				"environment": "prod",
				"warehouse":   "analytics",
				"access":      "dev",
			},
			want: []string{
				"snowflake.prod.analytics.dev.json",
			},
		},

		// ERROR cases
		{
			name:       "unknown_type",
			secretType: "invalid",
			metadata:   map[string]string{"environment": "dev"},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FilesToGenerate(tt.secretType, tt.metadata)

			if (err != nil) != tt.wantErr {
				t.Errorf("FilesToGenerate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilesToGenerate() = %v, want %v", got, tt.want)
			}
		})
	}
}
