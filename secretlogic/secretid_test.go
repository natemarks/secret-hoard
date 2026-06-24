package secretlogic

import "testing"

func TestBuildSecretID(t *testing.T) {
	tests := []struct {
		name       string
		secretType string
		metadata   map[string]string
		want       string
		wantErr    bool
	}{
		// JSONDOC cases
		{
			name:       "jsondoc_basic",
			secretType: "jsondoc",
			metadata:   map[string]string{"environment": "dev", "access": "app"},
			want:       "jsondoc/dev/app",
		},
		{
			name:       "jsondoc_prod",
			secretType: "jsondoc",
			metadata:   map[string]string{"environment": "prod", "access": "config"},
			want:       "jsondoc/prod/config",
		},
		{
			name:       "jsondoc_integration",
			secretType: "jsondoc",
			metadata:   map[string]string{"environment": "integration", "access": "idemia"},
			want:       "jsondoc/integration/idemia",
		},

		// TEXT_FILE cases
		{
			name:       "textfile_basic",
			secretType: "text_file",
			metadata:   map[string]string{"environment": "staging", "access": "keys"},
			want:       "text_file/staging/keys",
		},
		{
			name:       "textfile_prod",
			secretType: "text_file",
			metadata:   map[string]string{"environment": "prod", "access": "config-file"},
			want:       "text_file/prod/config-file",
		},

		// SNOWFLAKE cases (3 metadata fields)
		{
			name:       "snowflake_basic",
			secretType: "snowflake",
			metadata: map[string]string{
				"environment": "prod",
				"warehouse":   "analytics",
				"access":      "readonly",
			},
			want: "snowflake/prod/analytics/readonly",
		},
		{
			name:       "snowflake_dev",
			secretType: "snowflake",
			metadata: map[string]string{
				"environment": "dev",
				"warehouse":   "compute",
				"access":      "developer",
			},
			want: "snowflake/dev/compute/developer",
		},

		// RDSPOSTGRES cases (4 metadata fields)
		{
			name:       "rdspostgres_basic",
			secretType: "rdspostgres",
			metadata: map[string]string{
				"environment": "dev",
				"instance":    "db1",
				"database":    "myapp",
				"access":      "readwrite",
			},
			want: "rdspostgres/dev/db1/myapp/readwrite",
		},
		{
			name:       "rdspostgres_readonly",
			secretType: "rdspostgres",
			metadata: map[string]string{
				"environment": "prod",
				"instance":    "primary",
				"database":    "orders",
				"access":      "readonly",
			},
			want: "rdspostgres/prod/primary/orders/readonly",
		},
		{
			name:       "rdspostgres_master",
			secretType: "rdspostgres",
			metadata: map[string]string{
				"environment": "test",
				"instance":    "test-db",
				"database":    "app",
				"access":      "master",
			},
			want: "rdspostgres/test/test-db/app/master",
		},

		// SSL_CERTIFICATE cases
		{
			name:       "sslcert_basic",
			secretType: "ssl_certificate",
			metadata:   map[string]string{"environment": "prod", "commonname": "api.example.com"},
			want:       "ssl_certificate/prod/api.example.com",
		},
		{
			name:       "sslcert_wildcard",
			secretType: "ssl_certificate",
			metadata:   map[string]string{"environment": "staging", "commonname": "*.example.com"},
			want:       "ssl_certificate/staging/*.example.com",
		},
		{
			name:       "sslcert_subdomain",
			secretType: "ssl_certificate",
			metadata:   map[string]string{"environment": "dev", "commonname": "test.example.com"},
			want:       "ssl_certificate/dev/test.example.com",
		},

		// ERROR cases - unknown type
		{
			name:       "unknown_type",
			secretType: "invalid",
			metadata:   map[string]string{"environment": "dev"},
			wantErr:    true,
		},

		// ERROR cases - missing environment (all types need this)
		{
			name:       "jsondoc_missing_env",
			secretType: "jsondoc",
			metadata:   map[string]string{"access": "app"},
			wantErr:    true,
		},

		// ERROR cases - missing required fields
		{
			name:       "jsondoc_missing_access",
			secretType: "jsondoc",
			metadata:   map[string]string{"environment": "dev"},
			wantErr:    true,
		},
		{
			name:       "textfile_missing_access",
			secretType: "text_file",
			metadata:   map[string]string{"environment": "dev"},
			wantErr:    true,
		},
		{
			name:       "sslcert_missing_commonname",
			secretType: "ssl_certificate",
			metadata:   map[string]string{"environment": "prod"},
			wantErr:    true,
		},
		{
			name:       "snowflake_missing_warehouse",
			secretType: "snowflake",
			metadata:   map[string]string{"environment": "prod", "access": "dev"},
			wantErr:    true,
		},
		{
			name:       "snowflake_missing_access",
			secretType: "snowflake",
			metadata:   map[string]string{"environment": "prod", "warehouse": "analytics"},
			wantErr:    true,
		},
		{
			name:       "rdspostgres_missing_instance",
			secretType: "rdspostgres",
			metadata:   map[string]string{"environment": "dev", "database": "app", "access": "ro"},
			wantErr:    true,
		},
		{
			name:       "rdspostgres_missing_database",
			secretType: "rdspostgres",
			metadata:   map[string]string{"environment": "dev", "instance": "db1", "access": "ro"},
			wantErr:    true,
		},
		{
			name:       "rdspostgres_missing_access",
			secretType: "rdspostgres",
			metadata:   map[string]string{"environment": "dev", "instance": "db1", "database": "app"},
			wantErr:    true,
		},

		// ERROR cases - empty values
		{
			name:       "empty_environment",
			secretType: "jsondoc",
			metadata:   map[string]string{"environment": "", "access": "app"},
			wantErr:    true,
		},
		{
			name:       "empty_access",
			secretType: "jsondoc",
			metadata:   map[string]string{"environment": "dev", "access": ""},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildSecretID(tt.secretType, tt.metadata)

			if (err != nil) != tt.wantErr {
				t.Errorf("BuildSecretID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("BuildSecretID() = %q, want %q", got, tt.want)
			}
		})
	}
}
