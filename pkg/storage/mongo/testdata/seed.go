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

var Token = entity.Token{
	Id:        "test",
	UserId:    "test-user",
	Type:      entity.AccessToken,
	ExpiresAt: time.Now(),
	IssuedAt:  time.Now(),
}

func Seed() error {
	if err := env.Load("app"); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	dbName := os.Getenv("DB_NAME")
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?replicaSet=mongodb&ssl=false", os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), dbName)

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to connect to mongodb: %w", err)
	}

	if err := client.Database(dbName).Collection("tokens").Drop(ctx); err != nil {
		return fmt.Errorf("failed to drop tokens collection: %w", err)
	}

	testData := map[string]interface{}{
		"_id":        Token.Id,
		"user_id":    Token.UserId,
		"type":       Token.Type,
		"expires_at": Token.ExpiresAt,
		"issued_at":  Token.IssuedAt,
	}

	if _, err := client.Database(dbName).Collection("tokens").InsertOne(ctx, testData); err != nil {
		return fmt.Errorf("failed to insert testData: %w", err)
	}

	if err := client.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect: %w", err)
	}

	return nil
}
