package mock

import (
	"context"

	"github.com/efimovalex/replaceme/adapters/mongodb"
	"github.com/stretchr/testify/mock"
)

type ClientMock struct {
	mock.Mock
}

func (m *ClientMock) FindOneUserByEmail(ctx context.Context, email string) (*mongodb.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*mongodb.User), args.Error(1)
}

func (m *ClientMock) InsertUser(ctx context.Context, u *mongodb.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *ClientMock) Ping() error {
	args := m.Called()
	return args.Error(0)
}
