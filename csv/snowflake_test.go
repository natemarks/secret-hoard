package csv

import (
	"testing"

	"github.com/natemarks/secret-hoard/types"
	"github.com/rs/zerolog"
)

func TestReadSnowflakeSecrets(t *testing.T) {
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
				filename: "../testdata/csv/ReadSnowflakeSecrets/fff.csv",
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
