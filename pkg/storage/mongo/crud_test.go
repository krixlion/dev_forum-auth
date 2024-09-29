package mongo_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/krixlion/dev_forum-auth/internal/gentest"
	"github.com/krixlion/dev_forum-auth/pkg/entity"
	"github.com/krixlion/dev_forum-auth/pkg/storage/mongo/mongotest"
	"github.com/krixlion/dev_forum-auth/pkg/storage/mongo/testdata"
	"github.com/krixlion/dev_forum-lib/filter"
	"go.mongodb.org/mongo-driver/bson"
)

func TestDB_Create(t *testing.T) {
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
				token: func() entity.Token {
					test := testdata.Token
					test.Id = gentest.RandomString(50)
					return test
				}(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()

			db, err := mongotest.NewMongo(ctx)
			if err != nil {
				t.Errorf("mongotest.NewMongo() error = %v", err)
				return
			}

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

func TestDB_Get(t *testing.T) {
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
				id: testdata.Token.Id,
			},
			want: testdata.Token,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()

			db, err := mongotest.NewMongo(ctx)
			if err != nil {
				t.Errorf("mongotest.NewMongo() error = %v", err)
				return
			}

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

func TestDB_GetMultiple(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping db.GetMultiple integration test...")
	}

	type args struct {
		filter filter.Filter
	}
	tests := []struct {
		name    string
		args    args
		want    []entity.Token
		wantErr bool
	}{
		{
			name: "Test if token is retrieved correctly",
			args: args{
				filter: filter.Filter{{
					Attribute: "user_id",
					Operator:  filter.Equal,
					Value:     testdata.Token.UserId,
				}},
			},
			want: []entity.Token{testdata.Token},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()

			db, err := mongotest.NewMongo(ctx)
			if err != nil {
				t.Errorf("mongotest.NewMongo() error = %v", err)
				return
			}

			got, err := db.GetMultiple(ctx, tt.args.filter)
			if (err != nil) != tt.wantErr {
				t.Errorf("DB.GetMultiple() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}

			if !cmp.Equal(got, tt.want, cmpopts.EquateApproxTime(time.Second*5)) {
				t.Errorf("DB.GetMultiple():\n got = %v\n want = %v\n %v", got, tt.want, cmp.Diff(got, tt.want))
			}
		})
	}
}

func TestDB_Delete(t *testing.T) {
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
				id: testdata.Token.Id,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()

			db, err := mongotest.NewMongo(ctx)
			if err != nil {
				t.Errorf("mongotest.NewMongo() error = %v", err)
				return
			}

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
