package db

import (
	"reflect"
	"testing"

	"github.com/krixlion/dev_forum-auth/pkg/entity"
)

func Test_makeDocumentFromToken(t *testing.T) {
	type args struct {
		token entity.Token
	}
	tests := []struct {
		name string
		args args
		want tokenDocument
	}{
		{
			name: "Test if correctly parses a test token",
			args: args{
				token: testToken,
			},
			want: testTokenDoc,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeDocumentFromToken(tt.args.token); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeTokenDocument() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_makeTokenFromDocument(t *testing.T) {
	type args struct {
		v tokenDocument
	}
	tests := []struct {
		name string
		args args
		want entity.Token
	}{
		{
			name: "Test if correctly makes a token",
			args: args{
				v: testTokenDoc,
			},
			want: testToken,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeTokenFromDocument(tt.args.v); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeTokenFromDocument() = %v, want %v", got, tt.want)
			}
		})
	}
}
