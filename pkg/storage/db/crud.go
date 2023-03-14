package db

import (
	"context"

	"github.com/krixlion/dev_forum-auth/pkg/entity"
)

func (db DB) Get(ctx context.Context, opaqueToken string) (entity.Token, error) {

	ctx, span := db.tracer.Start(ctx, "db.Get")
	defer span.End()

	panic("not implemented") // TODO: Implement
}

func (db DB) GetMultiple(ctx context.Context, offset string, limit string) ([]entity.Token, error) {
	ctx, span := db.tracer.Start(ctx, "db.GetMultiple")
	defer span.End()

	panic("not implemented") // TODO: Implement
}

func (db DB) Create(ctx context.Context, opaqueToken string, token entity.Token) error {
	ctx, span := db.tracer.Start(ctx, "db.Create")
	defer span.End()

	panic("not implemented") // TODO: Implement
}

func (db DB) Delete(ctx context.Context, id string) error {
	ctx, span := db.tracer.Start(ctx, "db.Delete")
	defer span.End()

	panic("not implemented") // TODO: Implement
}
