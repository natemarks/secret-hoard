package aws

import (
	"testing"

	"github.com/natemarks/secret-hoard/types"
	"github.com/rs/zerolog"
)

func TestCreateSnowflakeSecrets(t *testing.T) {
	// skipping to avoid the slow WS interaction while working on other tests
	t.Skip()
	if err := CredsOK(); err != nil {
		t.Fatalf("skipping test -  AWS credentials not valid")
	}
	t.Logf("AWS credentials are valid")

	err := setup(t)
	if err != nil {
		t.Fatalf("error setting up test: %s", err)
	}
	t.Logf("setup complete")

	type args struct {
		secrets []types.SnowflakeSecret
		log     *zerolog.Logger
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ggg",
			args: args{
				secrets: []types.SnowflakeSecret{
					{
						Metadata: types.SnowflakeSecretMetadata{
							ResourceType: "snowflake",
							Environment:  "testenv",
							Warehouse:    "warehouse",
							Access:       "read",
						},
						Data: types.SnowflakeSecretData{
							Password:    "mypassword",
							AccountName: "myaccountname",
							Warehouse:   "warehouse",
							Username:    "myusername",
						},
					},
				},
				log: &zerolog.Logger{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CreateSnowflakeSecrets(tt.args.secrets, tt.args.log); (err != nil) != tt.wantErr {
				t.Errorf("CreateSnowflakeSecrets() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
