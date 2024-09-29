package mongo

import (
	"context"

	"github.com/krixlion/dev_forum-auth/pkg/entity"
	"github.com/krixlion/dev_forum-lib/filter"
	"go.mongodb.org/mongo-driver/bson"
)

func (db Mongo) Get(ctx context.Context, opaqueToken string) (entity.Token, error) {
	ctx, span := db.tracer.Start(ctx, "db.Get")
	defer span.End()

	filter := bson.M{"_id": bson.M{"$eq": opaqueToken}}

	tokenDoc := tokenDocument{}
	if err := db.tokens.FindOne(ctx, filter).Decode(&tokenDoc); err != nil {
		return entity.Token{}, err
	}

	return makeTokenFromDocument(tokenDoc), nil
}

func (db Mongo) GetMultiple(ctx context.Context, query filter.Filter) ([]entity.Token, error) {
	ctx, span := db.tracer.Start(ctx, "db.GetMultiple")
	defer span.End()

	filterDoc, err := filterToBSON(query)
	if err != nil {
		return nil, err
	}

	result, err := db.tokens.Find(ctx, filterDoc)
	if err != nil {
		return nil, err
	}

	tokenDocs := []tokenDocument{}
	if err := result.All(ctx, &tokenDocs); err != nil {
		return nil, err
	}

	tokens := make([]entity.Token, 0, len(tokenDocs))

	for _, tokenDoc := range tokenDocs {
		token := makeTokenFromDocument(tokenDoc)
		tokens = append(tokens, token)
	}

	return tokens, nil
}

func (db Mongo) Create(ctx context.Context, token entity.Token) error {
	ctx, span := db.tracer.Start(ctx, "db.Create")
	defer span.End()

	tokenDoc := makeDocumentFromToken(token)

	_, err := db.tokens.InsertOne(ctx, tokenDoc)
	return err
}

func (db Mongo) Delete(ctx context.Context, id string) error {
	ctx, span := db.tracer.Start(ctx, "db.Delete")
	defer span.End()

	filter := bson.M{"_id": bson.M{"$eq": id}}

	_, err := db.tokens.DeleteOne(ctx, filter)
	return err
}
