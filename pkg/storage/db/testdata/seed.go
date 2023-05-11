package testdata

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/krixlion/dev_forum-auth/pkg/entity"
	"github.com/krixlion/dev_forum-lib/env"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var TestToken = entity.Token{
	Id:        "test",
	UserId:    "test-user",
	Type:      entity.AccessToken,
	ExpiresAt: time.Now(),
	IssuedAt:  time.Now(),
}

func Seed() error {
	env.Load("app")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	port := os.Getenv("DB_PORT")
	host := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")
	client, err := connect(ctx, host, port)
	if err != nil {
		return fmt.Errorf("failed to connect to mongodb: %w", err)
	}

	if err := client.Database(dbName).Collection("tokens").Drop(ctx); err != nil {
		return fmt.Errorf("failed to drop tokens collection: %w", err)
	}

	testData := map[string]interface{}{
		"_id":        TestToken.Id,
		"user_id":    TestToken.UserId,
		"type":       TestToken.Type,
		"expires_at": TestToken.ExpiresAt,
		"issued_at":  TestToken.IssuedAt,
	}

	if _, err := client.Database(dbName).Collection("tokens").InsertOne(ctx, testData); err != nil {
		return fmt.Errorf("failed to insert testData: %w", err)
	}

	if err := client.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect: %w", err)
	}

	return nil
}

func connect(ctx context.Context, host, port string) (*mongo.Client, error) {
	// dbUser := os.Getenv("DB_USER")
	// pass := os.Getenv("DB_PASS")
	// uri := fmt.Sprintf("mongodb://%s:%s@%s:%s/?retryWrites=true&w=majority&tls=false", dbUser, pass, host, port)

	uri := fmt.Sprintf("mongodb://%s:%s/?retryWrites=true&w=majority&tls=false", host, port)

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	return client, nil
}
