package server

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	pb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1"
	"github.com/krixlion/dev_forum-lib/mocks"
	"github.com/krixlion/dev_forum-lib/nulls"
	"google.golang.org/grpc"
)

func setUpStubServer() AuthServer {
	return AuthServer{
		tracer: nulls.NullTracer{},
		logger: nulls.NullLogger{},
	}
}

func TestAuthServer_validateSignIn(t *testing.T) {
	type args struct {
		req     *pb.SignInRequest
		handler grpc.UnaryHandler
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "Test if fails on invalid email",
			args: args{
				req: &pb.SignInRequest{
					Email:    "sdfsd.aaasad/asdfsg",
					Password: "zaq1@WSX",
				},
				handler: mocks.NewUnaryHandler().GetMock(),
			},
			wantErr: true,
		},
		{
			name: "Test if fails on empty password",
			args: args{
				req: &pb.SignInRequest{
					Password: "",
					Email:    "example@test.com",
				},
				handler: mocks.NewUnaryHandler().GetMock(),
			},
			wantErr: true,
		},
		{
			name: "Test if fails on empty email",
			args: args{
				req: &pb.SignInRequest{
					Password: "zaq1@WSX",
				},
				handler: mocks.NewUnaryHandler().GetMock(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			server := setUpStubServer()

			got, err := server.validateSignIn(ctx, tt.args.req, tt.args.handler)
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthServer.validateSignIn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want) && tt.wantErr {
				t.Errorf("AuthServer.validateSignIn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthServer_validateSignOut(t *testing.T) {
	type args struct {
		req     *pb.SignOutRequest
		handler grpc.UnaryHandler
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "Test if fails on empty refresh token",
			args: args{
				req: &pb.SignOutRequest{
					RefreshToken: "",
				},
				handler: mocks.NewUnaryHandler().GetMock(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			server := setUpStubServer()

			got, err := server.validateSignOut(ctx, tt.args.req, tt.args.handler)
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthServer.validateSignOut() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want) && tt.wantErr {
				t.Errorf("AuthServer.validateSignOut() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthServer_validateGetAccessToken(t *testing.T) {
	type args struct {
		req     *pb.GetAccessTokenRequest
		handler grpc.UnaryHandler
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "Test if fails on empty refresh token",
			args: args{
				req: &pb.GetAccessTokenRequest{
					RefreshToken: "",
				},
				handler: mocks.NewUnaryHandler().GetMock(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			server := setUpStubServer()

			got, err := server.validateGetAccessToken(ctx, tt.args.req, tt.args.handler)
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthServer.validateGetAccessToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want) && tt.wantErr {
				t.Errorf("AuthServer.validateGetAccessToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthServer_validateTranslateAccessToken(t *testing.T) {
	type args struct {
		req     *pb.TranslateAccessTokenRequest
		handler grpc.UnaryHandler
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "Test if fails on empty access token",
			args: args{
				req: &pb.TranslateAccessTokenRequest{
					OpaqueAccessToken: "",
				},
				handler: mocks.NewUnaryHandler().GetMock(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			server := setUpStubServer()

			got, err := server.validateTranslateAccessToken(ctx, tt.args.req, tt.args.handler)
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthServer.validateTranslateAccessToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want) && tt.wantErr {
				t.Errorf("AuthServer.validateTranslateAccessToken() = %v, want %v", got, tt.want)
			}
		})
	}
}
