package storage

import (
	"context"
	"io"

	"github.com/krixlion/dev_forum-auth/pkg/entity"
)

type Storage interface {
	Getter
	Writer
}

type Getter interface {
	io.Closer
	Get(ctx context.Context, opaqueToken string) (entity.Token, error)
	GetMultiple(ctx context.Context, offset, limit string) ([]entity.Token, error)
}

type Writer interface {
	io.Closer
	Create(ctx context.Context, opaqueToken string, token entity.Token) error
	Delete(ctx context.Context, id string) error
}
