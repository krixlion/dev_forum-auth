package storage

import (
	"context"
	"io"

	"github.com/krixlion/dev_forum-auth/pkg/entity"
	"github.com/krixlion/dev_forum-lib/event"
)

type CQRStorage interface {
	Getter
	Writer
	CatchUp(event.Event)
}

type Storage interface {
	Getter
	Writer
}

type Eventstore interface {
	event.Consumer
	Writer
}

type Getter interface {
	io.Closer
	Get(ctx context.Context, id string) (entity.Token, error)
	GetMultiple(ctx context.Context, offset, limit string) ([]entity.Token, error)
}

type Writer interface {
	io.Closer
	Create(context.Context, entity.Token) error
	Update(context.Context, entity.Token) error
	Delete(ctx context.Context, id string) error
}
