package store

import (
	"reflect"
	"testing"

	"github.com/natemarks/secret-hoard/types"
	"github.com/rs/zerolog"
)

func TestRDSSecretsToCSVString(t *testing.T) {
	type args struct {
		secrets []types.RDSSecret
		log     *zerolog.Logger
	}
	tests := []struct {
		name        string
		args        args
		wantCsvData string
		wantErr     bool
	}{
		{
			name: "hhh",
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
			wantCsvData: "ResourceType,Environment,Instance,Database,Access,Password,Engine,Port,DbInstanceIdentifier,Host,Username\nrdspostgres,environment,instance,database,type,password,postgres,5432,dbInstanceIdentifier,host,username\n",
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCsvData, err := RDSSecretsToCSVString(tt.args.secrets, tt.args.log)
			if (err != nil) != tt.wantErr {
				t.Errorf("RDSSecretsToCSVString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotCsvData != tt.wantCsvData {
				t.Errorf("RDSSecretsToCSVString() gotCsvData = %v, want %v", gotCsvData, tt.wantCsvData)
			}
		})
	}
}

func TestRDSSecretsFromCSVString(t *testing.T) {
	type args struct {
		csvData string
		log     *zerolog.Logger
	}
	tests := []struct {
		name    string
		args    args
		want    []types.RDSSecret
		wantErr bool
	}{
		{
			name: "hhh",
			args: args{
				csvData: "ResourceType,Environment,Instance,Database,Access,Password,Engine,Port,DbInstanceIdentifier,Host,Username\nrdspostgres,environment,instance,database,type,password,postgres,5432,dbInstanceIdentifier,host,username\n",
				log:     GetTestLogger(),
			},
			want: []types.RDSSecret{
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
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RDSSecretsFromCSVString(tt.args.csvData, tt.args.log)
			if (err != nil) != tt.wantErr {
				t.Errorf("RDSSecretsFromCSVString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RDSSecretsFromCSVString() got = %v, want %v", got, tt.want)
			}
		})
	}
}
