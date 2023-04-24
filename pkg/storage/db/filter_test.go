package db

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/krixlion/dev_forum-auth/pkg/helpers/gentest"
	"github.com/krixlion/dev_forum-lib/filter"
	"go.mongodb.org/mongo-driver/bson"
)

func Test_filterToBSON(t *testing.T) {
	type args struct {
		params []filter.Parameter
	}
	tests := []struct {
		name    string
		args    args
		want    bson.D
		wantErr bool
	}{
		{
			name: "Test if does not return unexpected errors on valid filter",
			args: args{
				params: []filter.Parameter{
					{
						Attribute: "user_id",
						Operator:  filter.Equal,
						Value:     "test-user",
					},
				},
			},
			want: bson.D{{"user_id", bson.D{{"$eq", "test-user"}}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := filterToBSON(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("filterToBSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !cmp.Equal(got, tt.want) {
				t.Errorf("filterToBSON():\n got = %+v\n want %+v\n %v", got, tt.want, cmp.Diff(got, tt.want))
			}
		})
	}
}

func Test_toMongoOperator(t *testing.T) {
	type args struct {
		op filter.Operator
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Test if recognizes equal operator",
			args: args{
				op: filter.Equal,
			},
			want: "$eq",
		},
		{
			name: "Test if returns an error on invalid operator",
			args: args{
				op: filter.Operator(gentest.RandomString(10)),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toMongoOperator(tt.args.op)
			if (err != nil) != tt.wantErr {
				t.Errorf("toMongoOperator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("toMongoOperator():\n got = %+v\n want %+v\n %v", got, tt.want, cmp.Diff(got, tt.want))
			}
		})
	}
}
