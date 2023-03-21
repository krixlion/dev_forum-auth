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

type Getter interface {
	Get(ctx context.Context, opaqueToken string) (entity.Token, error)
	// GetMultiple(ctx context.Context, offset, limit string) ([]entity.Token, error)
}

type Writer interface {
	Create(ctx context.Context, token entity.Token) error
	Delete(ctx context.Context, id string) error
}
