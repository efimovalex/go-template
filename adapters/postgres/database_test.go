package postgres

import (
	"errors"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	type args struct {
		host     string
		port     string
		user     string
		password string
		database string
		ssl      string
	}
	tests := []struct {
		name    string
		args    args
		want    *Client
		wantErr error
	}{
		{
			name: "Connects to local database",
			args: args{
				host:     "localhost",
				port:     "5433",
				user:     "replaceme",
				password: "replaceme",
				database: "replaceme_test",
				ssl:      "disable",
			},
		},
		{
			name: "Connects to local database, wrong database",
			args: args{
				host:     "localhost",
				port:     "5433",
				user:     "replaceme",
				password: "replaceme",
				database: "replaceme_test_wrong",
				ssl:      "disable",
			},
			wantErr: errors.New(`pq: database "replaceme_test_wrong" does not exist`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.host, tt.args.port, tt.args.user, tt.args.password, tt.args.database, tt.args.ssl)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
			}
		})
	}
}

func TestNewTestDB(t *testing.T) {
	type args struct {
		t *testing.T
	}
	tests := []struct {
		name    string
		args    args
		wantNil bool
	}{
		{
			name: "Connects to local database",
			args: args{
				t: t,
			},
		},
		{
			name: "Connects to local database",
			args: args{
				t: nil,
			},
			wantNil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewTestDB(tt.args.t)
			if tt.wantNil {
				assert.Nil(t, got)
			} else {
				assert.NotNil(t, got)
			}
		})
	}
}
