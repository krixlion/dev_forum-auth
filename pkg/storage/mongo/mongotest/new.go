package mongotest

import (
	"context"
	"os"

	"github.com/krixlion/dev_forum-auth/pkg/storage/mongo"
	"github.com/krixlion/dev_forum-auth/pkg/storage/mongo/testdata"
	"github.com/krixlion/dev_forum-lib/env"
	"github.com/krixlion/dev_forum-lib/nulls"
)

func NewMongo(ctx context.Context) (mongo.Mongo, error) {
	env.Load("app")

	port := os.Getenv("DB_PORT")
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")
	storage, err := mongo.Make(user, pass, host, port, dbName, nulls.NullLogger{}, nulls.NullTracer{})
	if err != nil {
		return mongo.Mongo{}, err
	}

	// Prepare the database for each test.
	if err := testdata.Seed(); err != nil {
		return mongo.Mongo{}, err
	}

	go func() {
		<-ctx.Done()
		storage.Close()
	}()

	return storage, nil
}
