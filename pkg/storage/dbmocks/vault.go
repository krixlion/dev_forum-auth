package dbmocks

import (
	"context"

	"github.com/krixlion/dev_forum-auth/pkg/entity"
	"github.com/stretchr/testify/mock"
)

type Vault struct {
	*mock.Mock
}

func NewVault() Vault {
	return Vault{
		Mock: new(mock.Mock),
	}
}

// Get(ctx context.Context, id string) (entity.Key, error)
func (m Vault) GetRandom(ctx context.Context) (entity.Key, error) {
	args := m.Called(ctx)
	return args.Get(0).(entity.Key), args.Error(1)
}

func (m Vault) GetKeySet(ctx context.Context) ([]entity.Key, error) {
	args := m.Called(ctx)
	return args.Get(0).([]entity.Key), args.Error(1)
}
