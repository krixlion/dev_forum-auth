package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/krixlion/dev_forum-lib/env"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	env.Load("app")

	dbPort := os.Getenv("DB_PORT")
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")
	// dbUser := os.Getenv("DB_USER")
	// dbPass := os.Getenv("DB_PASS")
	// uri := fmt.Sprintf("mongodb://%s:%s@%s:%s/?retryWrites=true&w=majority&tls=false", dbUser, dbPass, dbHost, dbPort)
	uri := fmt.Sprintf("mongodb://%s:%s/?retryWrites=true&w=majority&tls=false", dbHost, dbPort)

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Database(dbName).Collection("tokens").Drop(ctx); err != nil {
		log.Fatal(err)
	}
}
