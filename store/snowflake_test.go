package store

import (
	"reflect"
	"testing"

	"github.com/natemarks/secret-hoard/types"
	"github.com/rs/zerolog"
)

func TestReadSnowflakeSecrets(t *testing.T) {
	t.Skip()
	type args struct {
		filename string
		log      *zerolog.Logger
	}
	tests := []struct {
		name    string
		args    args
		want    []types.SnowflakeSecret
		wantErr bool
	}{
		{
			name: "fff",
			args: args{
				filename: "../testdata/store/ReadSnowflakeSecrets/fff.store",
				log:      &zerolog.Logger{},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadSnowflakeSecrets(tt.args.filename, tt.args.log)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadSnowflakeSecrets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != 2 {
				t.Errorf("ReadSnowflakeSecrets() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSnowflakeSecretsFromCSVString(t *testing.T) {
	type args struct {
		csvData string
		log     *zerolog.Logger
	}
	tests := []struct {
		name    string
		args    args
		want    []types.SnowflakeSecret
		wantErr bool
	}{
		{
			name: "fff",
			args: args{
				csvData: "ResourceType,Environment,Warehouse,Access,Password,AccountName,Username\n" +
					"Snowflake,dev,warehouse1,read,pass1,account1,user1\n" +
					"Snowflake,dev,warehouse2,write,pass2,account2,user2\n",
				log: &zerolog.Logger{},
			},
			want: []types.SnowflakeSecret{
				{
					Metadata: types.SnowflakeSecretMetadata{
						ResourceType: "Snowflake",
						Environment:  "dev",
						Warehouse:    "warehouse1",
						Access:       "read",
					},
					Data: types.SnowflakeSecretData{
						Password:    "pass1",
						AccountName: "account1",
						Warehouse:   "warehouse1",
						Username:    "user1",
					},
				},
				{
					Metadata: types.SnowflakeSecretMetadata{
						ResourceType: "Snowflake",
						Environment:  "dev",
						Warehouse:    "warehouse2",
						Access:       "write",
					},
					Data: types.SnowflakeSecretData{
						Password:    "pass2",
						AccountName: "account2",
						Warehouse:   "warehouse2",
						Username:    "user2",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SnowflakeSecretsFromCSVString(tt.args.csvData, tt.args.log)
			if (err != nil) != tt.wantErr {
				t.Errorf("SnowflakeSecretsFromCSVString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SnowflakeSecretsFromCSVString() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSnowflakeSecretsToCSvString(t *testing.T) {
	type args struct {
		secrets []types.SnowflakeSecret
		log     *zerolog.Logger
	}
	tests := []struct {
		name        string
		args        args
		wantCsvData string
		wantErr     bool
	}{
		{
			name: "fff",
			args: args{
				secrets: []types.SnowflakeSecret{
					{
						Metadata: types.SnowflakeSecretMetadata{
							ResourceType: "Snowflake",
							Environment:  "dev",
							Warehouse:    "warehouse1",
							Access:       "read",
						},
						Data: types.SnowflakeSecretData{
							Password:    "pass1",
							AccountName: "account1",
							Warehouse:   "warehouse1",
							Username:    "user1",
						},
					},
					{
						Metadata: types.SnowflakeSecretMetadata{
							ResourceType: "Snowflake",
							Environment:  "dev",
							Warehouse:    "warehouse2",
							Access:       "write",
						},
						Data: types.SnowflakeSecretData{
							Password:    "pass2",
							AccountName: "account2",
							Warehouse:   "warehouse2",
							Username:    "user2",
						},
					},
				},
				log: &zerolog.Logger{},
			},
			wantCsvData: "ResourceType,Environment,Warehouse,Access,Password,AccountName,Username\n" +
				"Snowflake,dev,warehouse1,read,pass1,account1,user1\n" +
				"Snowflake,dev,warehouse2,write,pass2,account2,user2\n",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCsvData, err := SnowflakeSecretsToCSvString(tt.args.secrets, tt.args.log)
			if (err != nil) != tt.wantErr {
				t.Errorf("SnowflakeSecretsToCSvString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotCsvData != tt.wantCsvData {
				t.Errorf("SnowflakeSecretsToCSvString() gotCsvData = %v, want %v", gotCsvData, tt.wantCsvData)
			}
		})
	}
}
