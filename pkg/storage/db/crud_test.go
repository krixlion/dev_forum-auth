package db

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/krixlion/dev_forum-auth/pkg/entity"
	"github.com/krixlion/dev_forum-lib/env"
	"github.com/krixlion/dev_forum-lib/nulls"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	testToken = entity.Token{
		Id:        "test",
		UserId:    "test-user",
		Type:      entity.AccessToken,
		ExpiresAt: time.Now(),
		IssuedAt:  time.Now(),
	}

	testTokenDoc = tokenDocument{
		Id:        testToken.Id,
		UserId:    testToken.UserId,
		Type:      string(testToken.Type),
		ExpiresAt: testToken.ExpiresAt,
		IssuedAt:  testToken.IssuedAt,
	}
)

func setUpDB() DB {
	env.Load("app")

	port := os.Getenv("DB_PORT")
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")
	storage, err := Make(user, pass, host, port, dbName, nulls.NullLogger{}, nulls.NullTracer{})
	if err != nil {
		panic(err)
	}

	return storage
}

func Test_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping db.Create integration test...")
	}

	type args struct {
		token entity.Token
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test if random token is created correctly",
			args: args{
				token: testToken,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()

			db := setUpDB()

			if err := db.Create(ctx, tt.args.token); (err != nil) != tt.wantErr {
				t.Errorf("DB.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			var doc tokenDocument
			if err := db.tokens.FindOne(ctx, bson.M{"_id": tt.args.token.Id}).Decode(&doc); err != nil {
				t.Errorf("DB.tokens.FindOne() error = %v", err)
				return
			}

			got := makeTokenFromDocument(doc)

			if !cmp.Equal(tt.args.token, got, cmpopts.EquateApproxTime(time.Second*2)) {
				t.Errorf("DB.tokens.FindOne():\n want = %v\n got %v\n %v", tt.args.token, got, cmp.Diff(tt.args.token, got))
			}

		})
	}
}

func Test_Get(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping db.Get integration test...")
	}

	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    entity.Token
		wantErr bool
	}{
		{
			name: "Test if token is retrieved correctly",
			args: args{
				id: testToken.Id,
			},
			want: testToken,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()

			db := setUpDB()

			got, err := db.Get(ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("DB.Get() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}

			if !cmp.Equal(got, tt.want, cmpopts.EquateApproxTime(time.Second*5)) {
				t.Errorf("DB.Get():\n got = %v\n want = %v\n %v", got, tt.want, cmp.Diff(got, tt.want))
			}
		})
	}
}

func Test_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping db.Delete integration test...")
	}

	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test if token is deleted correctly.",
			args: args{
				id: testToken.Id,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()

			db := setUpDB()

			if err := db.Delete(ctx, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("DB.Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err := db.tokens.FindOne(ctx, bson.D{{Key: "_id", Value: tt.args.id}}).Err(); err == nil {
				t.Errorf("DB.tokens.FindOne() error not nil")
			}
		})
	}
}
