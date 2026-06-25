package secretlogic

import "testing"

func TestNormalizeSecretID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "textfile to text_file",
			input: "textfile/dev/config",
			want:  "text_file/dev/config",
		},
		{
			name:  "sslcert to ssl_certificate",
			input: "sslcert/prod/example.com",
			want:  "ssl_certificate/prod/example.com",
		},
		{
			name:  "jsondoc unchanged",
			input: "jsondoc/dev/app-config",
			want:  "jsondoc/dev/app-config",
		},
		{
			name:  "already normalized text_file",
			input: "text_file/dev/config",
			want:  "text_file/dev/config",
		},
		{
			name:  "already normalized ssl_certificate",
			input: "ssl_certificate/prod/example.com",
			want:  "ssl_certificate/prod/example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeSecretID(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeSecretID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseJSONDocSecretID(t *testing.T) {
	tests := []struct {
		name       string
		secretID   string
		wantEnv    string
		wantAccess string
		wantErr    bool
	}{
		{
			name:       "valid jsondoc",
			secretID:   "jsondoc/dev/app-config",
			wantEnv:    "dev",
			wantAccess: "app-config",
			wantErr:    false,
		},
		{
			name:       "valid prod",
			secretID:   "jsondoc/prod/database-config",
			wantEnv:    "prod",
			wantAccess: "database-config",
			wantErr:    false,
		},
		{
			name:     "missing access",
			secretID: "jsondoc/dev",
			wantErr:  true,
		},
		{
			name:     "too many parts",
			secretID: "jsondoc/dev/app/extra",
			wantErr:  true,
		},
		{
			name:     "empty string",
			secretID: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEnv, gotAccess, err := ParseJSONDocSecretID(tt.secretID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseJSONDocSecretID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotEnv != tt.wantEnv {
					t.Errorf("ParseJSONDocSecretID() env = %v, want %v", gotEnv, tt.wantEnv)
				}
				if gotAccess != tt.wantAccess {
					t.Errorf("ParseJSONDocSecretID() access = %v, want %v", gotAccess, tt.wantAccess)
				}
			}
		})
	}
}

func TestParseTextFileSecretID(t *testing.T) {
	tests := []struct {
		name       string
		secretID   string
		wantEnv    string
		wantAccess string
		wantErr    bool
	}{
		{
			name:       "valid textfile",
			secretID:   "textfile/dev/api-key",
			wantEnv:    "dev",
			wantAccess: "api-key",
			wantErr:    false,
		},
		{
			name:       "valid text_file normalized",
			secretID:   "text_file/prod/secret-token",
			wantEnv:    "prod",
			wantAccess: "secret-token",
			wantErr:    false,
		},
		{
			name:     "missing access",
			secretID: "textfile/dev",
			wantErr:  true,
		},
		{
			name:     "empty string",
			secretID: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEnv, gotAccess, err := ParseTextFileSecretID(tt.secretID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTextFileSecretID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotEnv != tt.wantEnv {
					t.Errorf("ParseTextFileSecretID() env = %v, want %v", gotEnv, tt.wantEnv)
				}
				if gotAccess != tt.wantAccess {
					t.Errorf("ParseTextFileSecretID() access = %v, want %v", gotAccess, tt.wantAccess)
				}
			}
		})
	}
}

func TestParseSSLCertSecretID(t *testing.T) {
	tests := []struct {
		name           string
		secretID       string
		wantEnv        string
		wantCommonName string
		wantErr        bool
	}{
		{
			name:           "valid sslcert",
			secretID:       "sslcert/prod/example.com",
			wantEnv:        "prod",
			wantCommonName: "example.com",
			wantErr:        false,
		},
		{
			name:           "valid ssl_certificate normalized",
			secretID:       "ssl_certificate/dev/test.example.com",
			wantEnv:        "dev",
			wantCommonName: "test.example.com",
			wantErr:        false,
		},
		{
			name:           "wildcard cert",
			secretID:       "sslcert/prod/*.example.com",
			wantEnv:        "prod",
			wantCommonName: "*.example.com",
			wantErr:        false,
		},
		{
			name:     "missing commonName",
			secretID: "sslcert/prod",
			wantErr:  true,
		},
		{
			name:     "empty string",
			secretID: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEnv, gotCommonName, err := ParseSSLCertSecretID(tt.secretID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSSLCertSecretID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotEnv != tt.wantEnv {
					t.Errorf("ParseSSLCertSecretID() env = %v, want %v", gotEnv, tt.wantEnv)
				}
				if gotCommonName != tt.wantCommonName {
					t.Errorf("ParseSSLCertSecretID() commonName = %v, want %v", gotCommonName, tt.wantCommonName)
				}
			}
		})
	}
}

func TestBuildContentsFilename(t *testing.T) {
	tests := []struct {
		name       string
		secretType string
		env        string
		access     string
		commonName string
		want       string
	}{
		{
			name:       "jsondoc",
			secretType: "jsondoc",
			env:        "dev",
			access:     "app-config",
			want:       "jsondoc.dev.app-config.contents.json",
		},
		{
			name:       "textfile",
			secretType: "textfile",
			env:        "prod",
			access:     "api-key",
			want:       "textfile.prod.api-key.contents.txt",
		},
		{
			name:       "text_file normalized",
			secretType: "text_file",
			env:        "dev",
			access:     "token",
			want:       "textfile.dev.token.contents.txt",
		},
		{
			name:       "sslcert base name",
			secretType: "sslcert",
			env:        "prod",
			commonName: "example.com",
			want:       "sslcert.prod.example.com",
		},
		{
			name:       "ssl_certificate normalized",
			secretType: "ssl_certificate",
			env:        "dev",
			commonName: "test.com",
			want:       "sslcert.dev.test.com",
		},
		{
			name:       "unknown type",
			secretType: "unknown",
			env:        "dev",
			access:     "test",
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildContentsFilename(tt.secretType, tt.env, tt.access, tt.commonName)
			if got != tt.want {
				t.Errorf("BuildContentsFilename() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildSSLCertFilenames(t *testing.T) {
	tests := []struct {
		name       string
		env        string
		commonName string
		wantCert   string
		wantKey    string
	}{
		{
			name:       "simple domain",
			env:        "prod",
			commonName: "example.com",
			wantCert:   "sslcert.prod.example.com.crt",
			wantKey:    "sslcert.prod.example.com.key",
		},
		{
			name:       "subdomain",
			env:        "dev",
			commonName: "api.example.com",
			wantCert:   "sslcert.dev.api.example.com.crt",
			wantKey:    "sslcert.dev.api.example.com.key",
		},
		{
			name:       "wildcard",
			env:        "prod",
			commonName: "*.example.com",
			wantCert:   "sslcert.prod.*.example.com.crt",
			wantKey:    "sslcert.prod.*.example.com.key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCert, gotKey := BuildSSLCertFilenames(tt.env, tt.commonName)
			if gotCert != tt.wantCert {
				t.Errorf("BuildSSLCertFilenames() cert = %q, want %q", gotCert, tt.wantCert)
			}
			if gotKey != tt.wantKey {
				t.Errorf("BuildSSLCertFilenames() key = %q, want %q", gotKey, tt.wantKey)
			}
		})
	}
}
