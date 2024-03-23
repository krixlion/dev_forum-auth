package storage

import (
	"context"
	"io"

	"github.com/krixlion/dev_forum-auth/pkg/entity"
	"github.com/krixlion/dev_forum-lib/filter"
)

type Storage interface {
	Getter
	Writer
	io.Closer
}

type Vault interface {
	GetRandom(ctx context.Context) (entity.Key, error)
	GetKeySet(ctx context.Context) ([]entity.Key, error)
}

type Getter interface {
	// Token's id is its corresponding opaque token.
	Get(ctx context.Context, id string) (entity.Token, error)
	// Filter string according to the dev_forum-lib filter pkg.
	GetMultiple(ctx context.Context, filter filter.Filter) ([]entity.Token, error)
}

type Writer interface {
	Create(ctx context.Context, token entity.Token) error
	Delete(ctx context.Context, id string) error
}
