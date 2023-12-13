package aws

import (
	"testing"

	"github.com/natemarks/secret-hoard/store"

	"github.com/natemarks/secret-hoard/types"
	"github.com/rs/zerolog"
)

func TestCreateOrUpdateSnowflakeSecret(t *testing.T) {
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
		secret    types.SnowflakeSecret
		overwrite bool
		log       *zerolog.Logger
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "first - succeed",
			args: args{
				secret: types.SnowflakeSecret{
					Data: types.SnowflakeSecretData{
						Password:    "password",
						AccountName: "accountname",
						Warehouse:   "warehouse",
						Username:    "username",
					},
					Metadata: types.SnowflakeSecretMetadata{
						ResourceType: "snowflake",
						Environment:  "testenv",
						Warehouse:    "warehouse",
						Access:       "read",
					},
				},
				overwrite: false,
				log:       store.GetTestLogger(),
			},
		},
		{
			name: "second - succeed",
			args: args{
				secret: types.SnowflakeSecret{
					Data: types.SnowflakeSecretData{
						Password:    "password",
						AccountName: "accountname",
						Warehouse:   "warehouse",
						Username:    "username",
					},
					Metadata: types.SnowflakeSecretMetadata{
						ResourceType: "snowflake",
						Environment:  "testenv",
						Warehouse:    "warehouse",
						Access:       "read",
					},
				},
				overwrite: true,
				log:       store.GetTestLogger(),
			},
		},
		{
			name: "third - fail",
			args: args{
				secret: types.SnowflakeSecret{
					Data: types.SnowflakeSecretData{
						Password:    "password",
						AccountName: "accountname",
						Warehouse:   "warehouse",
						Username:    "username",
					},
					Metadata: types.SnowflakeSecretMetadata{
						ResourceType: "snowflake",
						Environment:  "testenv",
						Warehouse:    "warehouse",
						Access:       "read",
					},
				},
				overwrite: false,
				log:       store.GetTestLogger(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CreateOrUpdateSnowflakeSecret(tt.args.secret, tt.args.overwrite, tt.args.log)
		})
	}
}
