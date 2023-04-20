package storage

import (
	"context"
	"io"

	"github.com/krixlion/dev_forum-auth/pkg/entity"
)

type Storage interface {
	Getter
	Writer
	io.Closer
}

type Vault interface {
	// Get(ctx context.Context, id string) (entity.Key, error)
	GetRandom(ctx context.Context) (entity.Key, error)
	GetKeySet(ctx context.Context) ([]entity.Key, error)
}

type Getter interface {
	// Token's id is it's corresponding opaque token.
	Get(ctx context.Context, id string) (entity.Token, error)
}

type Writer interface {
	Create(ctx context.Context, token entity.Token) error
	Delete(ctx context.Context, id string) error
}
