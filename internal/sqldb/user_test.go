package sqldb

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_FindOneUserByEmail(t *testing.T) {
	db := NewTestDB(t)
	defer func() {
		db.ResetTable(t, "users")
	}()
	u := User{
		Email:       "test@test.com",
		Password:    "test",
		Description: "description",
		LastName:    "lastname",
		FirstName:   "firstname",
	}

	ctx := context.Background()
	err := db.InsertUser(ctx, &u)
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
			got, err := db.FindOneUserByEmail(ctx, tt.args.email)
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
	db := NewTestDB(t)
	defer func() {
		db.ResetTable(t, "users")
	}()
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
			wantErr: errors.New("user with email test@test.com does already exist"),
		},
		{
			name: "Test missing password",
			user: &User{
				Email:       "test2@test.com",
				Description: "description",
				LastName:    "lastname",
				FirstName:   "firstname",
			},
			wantErr: errors.New("password is required"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := db.InsertUser(ctx, tt.user)
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
