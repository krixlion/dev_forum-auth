package db

import (
	"context"
	"fmt"
	"time"

	"github.com/krixlion/dev_forum-lib/event"
	"github.com/krixlion/dev_forum-lib/logging"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
	"go.opentelemetry.io/otel/trace"
)

const databaseName = "auth-service"
const collectionName = "tokens"

type DB struct {
	client *mongo.Client
	logger logging.Logger
	tracer trace.Tracer
}

func MakeDB(user, pass, host, port string, logger logging.Logger, tracer trace.Tracer) (DB, error) {
	uri := fmt.Sprintf("mongodb+srv://%s:%s@%s:%s/?retryWrites=true&w=majority", user, pass, host, port)

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)
	// Add tracing and metrics.
	opts.Monitor = otelmongo.NewMonitor()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return DB{}, err
	}

	return DB{
		client: client,
		logger: logger,
		tracer: tracer,
	}, nil
}

func (db DB) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	if err := db.client.Disconnect(ctx); err != nil {
		return err
	}

	return nil
}

func (db DB) EventHandlers() map[event.EventType][]event.Handler {
	return nil
}
