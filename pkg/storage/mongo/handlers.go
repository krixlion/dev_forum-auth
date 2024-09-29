package mongo

import (
	"context"
	"encoding/json"
	"time"

	"github.com/krixlion/dev_forum-lib/event"
	"github.com/krixlion/dev_forum-lib/tracing"
	"go.mongodb.org/mongo-driver/bson"
)

// SignOutUsersOnDeletion deletes all, both access and refresh tokens created for the deleted user.
func (db Mongo) SignOutUsersOnDeletion() event.Handler {
	return event.HandlerFunc(func(e event.Event) {
		ctx, span := db.tracer.Start(tracing.InjectMetadataIntoContext(context.Background(), e.Metadata), "SignOutUsersOnDeletion")
		defer span.End()

		ctx, cancel := context.WithTimeout(ctx, time.Second*2)
		defer cancel()

		var id string
		if err := json.Unmarshal(e.Body, &id); err != nil {
			tracing.SetSpanErr(span, err)
			db.logger.Log(ctx, "failed to parse event body", "err", err)
			return
		}

		filter := bson.M{"user_id": bson.M{"$eq": id}} //nolint:govet // Unkeyed for convienience.

		if _, err := db.tokens.DeleteMany(ctx, filter); err != nil {
			tracing.SetSpanErr(span, err)
			db.logger.Log(ctx, "failed to delete tokens", "err", err)
		}
	})
}
