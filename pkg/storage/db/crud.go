package db

import (
	"context"

	"github.com/krixlion/dev_forum-auth/pkg/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (db DB) Get(ctx context.Context, opaqueToken string) (entity.Token, error) {
	ctx, span := db.tracer.Start(ctx, "db.Get")
	defer span.End()

	tokens := db.client.Database(databaseName).Collection(collectionName)

	filter := bson.D{{Key: "_id", Value: bson.D{{Key: "$eq", Value: opaqueToken}}}}
	opts := options.FindOne().SetHint(bson.D{{Key: "_id", Value: 1}})

	result := tokens.FindOne(ctx, filter, opts)
	tokenDoc := tokenDocument{}
	if err := result.Decode(&tokenDoc); err != nil {
		return entity.Token{}, err
	}

	token := makeTokenFromDoc(tokenDoc)

	return token, nil
}

// func (db DB) GetMultiple(ctx context.Context, offset string, limit string) ([]entity.Token, error) {
// 	ctx, span := db.tracer.Start(ctx, "db.GetMultiple")
// 	defer span.End()
// }

func (db DB) Create(ctx context.Context, token entity.Token) error {
	ctx, span := db.tracer.Start(ctx, "db.Create")
	defer span.End()

	tokens := db.client.Database(databaseName).Collection(collectionName)
	tokenDoc := makeTokenDocument(token)

	if _, err := tokens.InsertOne(ctx, tokenDoc); err != nil {
		return err
	}

	return nil
}

func (db DB) Delete(ctx context.Context, id string) error {
	ctx, span := db.tracer.Start(ctx, "db.Delete")
	defer span.End()

	tokens := db.client.Database(databaseName).Collection(collectionName)

	filter := bson.D{{Key: "_id", Value: bson.D{{Key: "$eq", Value: id}}}}
	opts := options.Delete().SetHint(bson.D{{Key: "_id", Value: 1}})

	if _, err := tokens.DeleteOne(ctx, filter, opts); err != nil {
		return err
	}

	return nil
}
