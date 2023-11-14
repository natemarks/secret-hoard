package main

import (
	"testing"

	"github.com/natemarks/secret-hoard/types"
	"github.com/rs/zerolog"
)

func TestCreateRDSSecrets(t *testing.T) {
	t.Skipf("skipping test -  example only")
	type args struct {
		secrets []types.RDSSecret
		log     *zerolog.Logger
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "sdfdsf",
			args: args{
				secrets: []types.RDSSecret{
					{
						Data: types.RDSSecretData{
							Password:             "password",
							Engine:               "postgres",
							Port:                 5432,
							DbInstanceIdentifier: "dbInstanceIdentifier",
							Host:                 "host",
							Username:             "username",
						},
						Metadata: types.RDSSecretMetadata{
							Environment: "myenv",
							Instance:    "myinstance",
							Database:    "mydb",
							Type:        "mytype",
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
			if err := CreateRDSSecrets(tt.args.secrets, tt.args.log); (err != nil) != tt.wantErr {
				t.Errorf("CreateRDSSecrets() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
