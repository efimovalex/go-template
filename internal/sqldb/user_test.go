package sqldb

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestClient_FindOneUserByEmail(t *testing.T) {
	db, err := New("localhost", "5432", "replaceme", "replaceme", "replaceme_test", "disable", zap.NewNop().Sugar())
	defer func() {
		db.Exec("TRUNCATE TABLE users")
	}()
	assert.NoError(t, err)
	u := User{
		Email:       "test@test.com",
		Password:    "test",
		Description: "description",
		LastName:    "lastname",
		FirstName:   "firstname",
	}

	err = db.InsertUser(&u)
	assert.NoError(t, err)

	type args struct {
		email string
	}
	tests := []struct {
		name    string
		args    args
		want    *User
		wantErr error
	}{
		{
			name: "Test find user by email",
			args: args{
				email: u.Email,
			},
			want: &u,
		},
		{
			name: "Test user not found",
			args: args{
				email: "not@found.com",
			},
			wantErr: nil,
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.FindOneUserByEmail(tt.args.email)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestClient_InsertUser(t *testing.T) {
	db, err := New("localhost", "5432", "replaceme", "replaceme", "replaceme_test", "disable", zap.NewNop().Sugar())
	defer func() {
		db.Exec("TRUNCATE TABLE users")
	}()
	assert.NoError(t, err)
	u := User{
		Email:       "test@test.com",
		Password:    "test",
		Description: "description",
		LastName:    "lastname",
		FirstName:   "firstname",
	}
	tests := []struct {
		name    string
		user    *User
		wantErr error
	}{
		{
			name:    "Test insert user",
			user:    &u,
			wantErr: nil,
		},
		{
			name:    "Test reinsert user",
			user:    &u,
			wantErr: errors.New("user.go:69: conflict: user with email test@test.com does already exist"),
		},
		{
			name: "Test missing password",
			user: &User{
				Email:       "test2@test.com",
				Description: "description",
				LastName:    "lastname",
				FirstName:   "firstname",
			},
			wantErr: errors.New("user.go:53: bad request: password is required"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := db.InsertUser(tt.user)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.wantErr.Error())

				return
			} else {
				assert.NoError(t, err)
			}

			assert.NotEmpty(t, tt.user.ID)
		})
	}
}
