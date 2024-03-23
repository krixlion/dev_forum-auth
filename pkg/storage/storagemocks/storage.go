package storagemocks

import (
	"context"

	"github.com/krixlion/dev_forum-auth/pkg/entity"
	"github.com/krixlion/dev_forum-lib/filter"
	"github.com/stretchr/testify/mock"
)

type Storage struct {
	*mock.Mock
}

func NewStorage() Storage {
	return Storage{
		Mock: new(mock.Mock),
	}
}

// Token's id is its corresponding opaque token.
func (m Storage) Get(ctx context.Context, id string) (entity.Token, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(entity.Token), args.Error(1)
}

func (m Storage) GetMultiple(ctx context.Context, query filter.Filter) ([]entity.Token, error) {
	args := m.Called(ctx, query)
	return args.Get(0).([]entity.Token), args.Error(1)
}

func (m Storage) Create(ctx context.Context, token entity.Token) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m Storage) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m Storage) Close() error {
	return nil
}
