package mongo_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/krixlion/dev_forum-auth/pkg/entity"
	"github.com/krixlion/dev_forum-auth/pkg/storage/mongo/mongotest"
	"github.com/krixlion/dev_forum-lib/event"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestMongo_SignOutUsersOnDeletion(t *testing.T) {
	t.Run("Test all tokens assigned to a user are deleted when the user is deleted", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		db, err := mongotest.NewMongo(ctx)
		if err != nil {
			t.Errorf("mongotest.NewMongo() error = %v", err)
			return
		}

		token := entity.Token{
			Id:        "test-id",
			UserId:    "test-user-id",
			Type:      entity.RefreshToken,
			ExpiresAt: time.Now().Add(time.Hour),
			IssuedAt:  time.Now(),
		}

		if err := db.Create(ctx, token); err != nil {
			t.Errorf("DB.Create() error = %v", err)
			return
		}

		e, err := event.MakeEvent(event.UserAggregate, event.UserDeleted, token.UserId, nil)
		if err != nil {
			t.Errorf("event.MakeEvent() error = %v", err)
			return
		}

		db.SignOutUsersOnDeletion().Handle(e)

		if _, err := db.Get(ctx, token.Id); !errors.Is(err, mongo.ErrNoDocuments) {
			t.Errorf("db.Get() error = %v", err)
			return
		}
	})
}
