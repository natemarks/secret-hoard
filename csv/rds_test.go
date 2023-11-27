package csv

import (
	"testing"

	"github.com/natemarks/secret-hoard/types"
	"github.com/rs/zerolog"
)

func TestReadRDSSecrets(t *testing.T) {
	type args struct {
		filename string
		log      *zerolog.Logger
	}
	tests := []struct {
		name    string
		args    args
		want    []types.RDSSecret
		wantErr bool
	}{
		{
			name: "fff",
			args: args{
				filename: "../testdata/csv/ReadRDSSecrets/fff.csv",
				log:      &zerolog.Logger{},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadRDSSecrets(tt.args.filename, tt.args.log)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadRDSSecrets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != 4 {
				t.Errorf("ReadRDSSecrets() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWriteRDSSecrets(t *testing.T) {

	type args struct {
		filename string
		secrets  []types.RDSSecret
		log      *zerolog.Logger
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ggg",
			args: args{
				filename: "../testdata/csv/WriteRDSSecrets/ggg.csv",
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
							ResourceType: "rdspostgres",
							Environment:  "environment",
							Instance:     "instance",
							Database:     "database",
							Access:       "type",
						},
					},
				},
				log: &zerolog.Logger{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := WriteRDSSecrets(tt.args.filename, tt.args.secrets, tt.args.log); (err != nil) != tt.wantErr {
				t.Errorf("WriteRDSSecrets() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
