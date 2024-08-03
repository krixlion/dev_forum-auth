package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/krixlion/dev_forum-lib/event"
	"github.com/krixlion/dev_forum-lib/event/dispatcher"
	"github.com/krixlion/dev_forum-lib/logging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
	"go.opentelemetry.io/otel/trace"
)

const collectionName = "tokens"

var _ dispatcher.Listener = (*Mongo)(nil)

type Mongo struct {
	client *mongo.Client
	tokens *mongo.Collection
	logger logging.Logger
	tracer trace.Tracer
}

func Make(user, pass, host, port, dbName string, logger logging.Logger, tracer trace.Tracer) (Mongo, error) {
	// uri := fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?retryWrites=true&w=majority&tls=false&authSource=admin", user, pass, host, port, dbName)
	uri := fmt.Sprintf("mongodb://%s:%s/%s?retryWrites=true&w=majority&tls=false", host, port, dbName)
	reg := bson.NewRegistryBuilder().Build()

	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(options.ServerAPI(options.ServerAPIVersion1)).SetRegistry(reg).SetMonitor(otelmongo.NewMonitor())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return Mongo{}, err
	}

	tokens := client.Database(dbName).Collection(collectionName)

	return Mongo{
		client: client,
		tokens: tokens,
		logger: logger,
		tracer: tracer,
	}, nil
}

func (db Mongo) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	return db.client.Disconnect(ctx)
}

func (db Mongo) EventHandlers() map[event.EventType][]event.Handler {
	// TODO: add handler for removing deleted users tokens.
	return nil
}
