package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/metadata"
	"github.com/krixlion/dev_forum-auth/pkg/entity"
	"github.com/krixlion/dev_forum-auth/pkg/entity/testdata"
	"github.com/krixlion/dev_forum-auth/pkg/tokens"
	"github.com/krixlion/dev_forum-auth/pkg/tokens/tokensmocks"
	"github.com/krixlion/dev_forum-lib/nulls"
)

func TestNewAuthFunc(t *testing.T) {
	type args struct {
		tokenParser tokens.Parser
		ctx         context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    entity.Token
		wantErr bool
	}{
		{
			name: "Test no error is returned on valid token",
			args: args{
				tokenParser: func() tokensmocks.TokenParser {
					m := tokensmocks.NewTokenParser()
					m.On("ParseToken", "test-token").Return(testdata.AccessToken, nil).Once()
					return m
				}(),
				ctx: metadata.MD{}.Add("authorization", "Bearer test-token").ToIncoming(context.Background()),
			},
			want:    testdata.AccessToken,
			wantErr: false,
		},
		{
			name: "Test error is returned if the context does not contain a Bearer token",
			args: args{
				tokenParser: tokensmocks.NewTokenParser(),
				ctx:         context.Background(),
			},
			wantErr: true,
		},
		{
			name: "Test error is returned if the parser fails to validate the token",
			args: args{
				tokenParser: func() tokensmocks.TokenParser {
					m := tokensmocks.NewTokenParser()
					m.On("ParseToken", "test-token").Return(entity.Token{}, errors.New("test-err")).Once()
					return m
				}(),
				ctx: metadata.MD{}.Add("authorization", "Bearer test-token").ToIncoming(context.Background()),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := NewAuthFunc(tt.args.tokenParser, nulls.NullTracer{})
			gotCtx, err := fn(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewAuthFunc():\n error = %v\n wantErr = %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			got := gotCtx.Value(ctxTokenKey{}).(entity.Token)

			if !cmp.Equal(got, tt.want) {
				t.Fatalf("NewAuthFunc():\n got = %v\n want = %v", got, tt.want)
			}
		})
	}
}

func TestGetTokenFromCtx(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    entity.Token
		wantErr bool
	}{
		{
			name: "Test returns no error when the token is found",
			args: args{
				ctx: context.WithValue(context.Background(), ctxTokenKey{}, testdata.AccessToken),
			},
			want:    testdata.AccessToken,
			wantErr: false,
		},
		{
			name: "Test returns an error when the token is of an unexpected type",
			args: args{
				ctx: context.WithValue(context.Background(), ctxTokenKey{}, "test-data"),
			},
			wantErr: true,
		},
		{
			name: "Test returns an error when the token is missing",
			args: args{
				context.Background(),
			},
			wantErr: true,
		},
		{
			name: "Test returns an error when the token is zero value",
			args: args{
				ctx: context.WithValue(context.Background(), ctxTokenKey{}, entity.Token{}),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetTokenFromCtx(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetTokenFromCtx() error = %v wantErr = %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if !cmp.Equal(got, tt.want) {
				t.Fatalf("GetTokenFromCtx():\n = %v\n want = %v", got, tt.want)
			}
		})
	}
}
