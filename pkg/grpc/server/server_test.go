package server

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/krixlion/dev_forum-auth/pkg/entity"
	pb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1"
	"github.com/krixlion/dev_forum-auth/pkg/storage"
	"github.com/krixlion/dev_forum-auth/pkg/storage/storagemocks"
	"github.com/krixlion/dev_forum-auth/pkg/tokens"
	"github.com/krixlion/dev_forum-auth/pkg/tokens/tokensmocks"
	"github.com/krixlion/dev_forum-lib/event/dispatcher"
	"github.com/krixlion/dev_forum-lib/filter"
	"github.com/krixlion/dev_forum-lib/nulls"
	userPb "github.com/krixlion/dev_forum-user/pkg/grpc/v1"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	empty "google.golang.org/protobuf/types/known/emptypb"
)

// Struct for server mock dependencies.
type deps struct {
	storage      storage.Storage
	vault        storage.Vault
	userClient   userPb.UserServiceClient
	tokenManager tokens.Manager
}

// setUpServer initializes and runs in the background a gRPC
// server allowing only for local calls for testing.
// Returns a client to interact with the server.
// The server is shutdown when provided context is cancelled.
// No interceptor is registered.
func setUpServer(ctx context.Context, d deps) pb.AuthServiceClient {
	// bufconn allows the server to call itself
	// great for testing across whole infrastructure
	lis := bufconn.Listen(1024 * 1024)
	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	config := Config{
		AccessTokenValidityTime:  time.Minute,
		RefreshTokenValidityTime: time.Minute,
	}

	deps := Dependencies{
		Services: Services{
			User: d.userClient,
		},
		Vault:        d.vault,
		TokenManager: d.tokenManager,
		Storage:      d.storage,
		Dispatcher:   dispatcher.NewDispatcher(0),
		Logger:       nulls.NullLogger{},
		Tracer:       nulls.NullTracer{},
	}

	s := grpc.NewServer()
	server := MakeAuthServer(deps, config)
	pb.RegisterAuthServiceServer(s, server)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with an error: %v", err)
		}
	}()

	go func() {
		<-ctx.Done()
		s.Stop()
	}()

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial bufnet: %v", err)
	}

	client := pb.NewAuthServiceClient(conn)
	return client
}

func TestAuthServer_SignIn(t *testing.T) {
	type args struct {
		req *pb.SignInRequest
	}
	tests := []struct {
		name    string
		deps    deps
		args    args
		want    *pb.SignInResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			client := setUpServer(ctx, tt.deps)

			got, err := client.SignIn(ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthServer.SignIn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want, cmpopts.IgnoreUnexported(pb.SignInResponse{}, pb.SignInResponse{})) {
				t.Errorf("AuthServer.SignIn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthServer_SignOut(t *testing.T) {
	type args struct {
		req *pb.SignOutRequest
	}
	tests := []struct {
		name    string
		deps    deps
		args    args
		wantErr bool
	}{
		{
			name: "Test if no unexpected errors are returned on valid flow",
			deps: deps{
				tokenManager: func() tokens.Manager {
					manager := tokensmocks.NewTokenManager()
					manager.On("DecodeOpaque", tokens.RefreshToken, "test-opaque").Return("test-opaque-seed", nil).Once()
					return manager
				}(),
				storage: func() storage.Storage {
					storage := storagemocks.NewStorage()
					testToken := entity.Token{
						Id:     "test-opaque-seed",
						UserId: "test",
						Type:   entity.RefreshToken,
					}
					testToken2 := entity.Token{
						Id:     "test-opaque-seeded",
						UserId: testToken.UserId,
						Type:   entity.AccessToken,
					}
					testTokens := []entity.Token{testToken, testToken2}
					query := filter.Filter{{
						Attribute: "user_id",
						Operator:  filter.Equal,
						Value:     testToken.UserId,
					}}

					storage.On("Get", mock.Anything, "test-opaque-seed").Return(testToken, nil).Once()
					storage.On("GetMultiple", mock.Anything, query).Return(testTokens, nil).Once()
					storage.On("Delete", mock.Anything, "test-opaque-seed").Return(nil).Once()
					storage.On("Delete", mock.Anything, "test-opaque-seeded").Return(nil).Once()
					return storage
				}(),
			},
			args: args{
				req: &pb.SignOutRequest{
					RefreshToken: "test-opaque",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			client := setUpServer(ctx, tt.deps)

			_, err := client.SignOut(ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthServer.SignOut() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestAuthServer_GetAccessToken(t *testing.T) {
	type args struct {
		req *pb.GetAccessTokenRequest
	}
	tests := []struct {
		name    string
		deps    deps
		args    args
		want    *pb.GetAccessTokenResponse
		wantErr bool
	}{
		{
			name: "Test if no unexpected errors are returned on valid flow",
			deps: deps{
				tokenManager: func() tokens.Manager {
					manager := tokensmocks.NewTokenManager()
					manager.On("DecodeOpaque", tokens.RefreshToken, "test-opaque").Return("test-opaque-decoded", nil).Once()
					manager.On("GenerateOpaque", tokens.AccessToken).Return("test-opaque-generated", "test-opaque-seed", nil).Once()
					return manager
				}(),
				storage: func() storage.Storage {
					storage := storagemocks.NewStorage()
					testToken := entity.Token{
						Id:     "test-opaque-seed",
						UserId: "test",
						Type:   entity.RefreshToken,
					}
					storage.On("Get", mock.Anything, "test-opaque-decoded").Return(testToken, nil).Once()
					storage.On("Create", mock.Anything, mock.AnythingOfType("entity.Token")).Return(nil).Once()
					return storage
				}(),
			},
			args: args{
				req: &pb.GetAccessTokenRequest{
					RefreshToken: "test-opaque",
				},
			},
			want: &pb.GetAccessTokenResponse{
				AccessToken: "test-opaque-generated",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			client := setUpServer(ctx, tt.deps)

			got, err := client.GetAccessToken(ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthServer.GetAccessToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Equals false if both are nil or they point to the same memory address
			// so be sure to use seperate variables when providing args in order to prevent SEGV.
			if got != tt.want {
				if !cmp.Equal(got, tt.want, cmpopts.IgnoreUnexported(pb.GetAccessTokenRequest{}, pb.GetAccessTokenResponse{})) {
					t.Errorf("AuthServer.GetAccessToken() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestAuthServer_TranslateAccessToken(t *testing.T) {
	type args struct {
		req *pb.TranslateAccessTokenRequest
	}
	tests := []struct {
		name    string
		args    args
		deps    deps
		want    *pb.TranslateAccessTokenResponse
		wantErr bool
	}{
		{
			name: "Test if no unexpected errors are returned on valid flow",
			deps: deps{
				tokenManager: func() tokens.Manager {
					manager := tokensmocks.NewTokenManager()
					manager.On("DecodeOpaque", tokens.AccessToken, "test-opaque").Return("test-opaque-decoded", nil).Once()
					manager.On("Encode", mock.AnythingOfType("entity.Key"), mock.AnythingOfType("entity.Token")).Return([]byte("test-jwt-encoded"), nil).Once()
					return manager
				}(),
				storage: func() storage.Storage {
					storage := storagemocks.NewStorage()
					testToken := entity.Token{
						Id:     "test",
						UserId: "test",
						Type:   entity.AccessToken,
					}
					storage.On("Get", mock.Anything, "test-opaque-decoded").Return(testToken, nil).Once()
					return storage
				}(),
				vault: func() storage.Vault {
					vault := storagemocks.NewVault()
					testKey := entity.Key{
						Id:        "test",
						Type:      "test",
						Algorithm: "test",
					}
					vault.On("GetRandom", mock.Anything).Return(testKey, nil).Once()
					return vault
				}(),
			},
			args: args{
				req: &pb.TranslateAccessTokenRequest{
					OpaqueAccessToken: "test-opaque",
				},
			},
			want: &pb.TranslateAccessTokenResponse{
				AccessToken: "test-jwt-encoded",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			client := setUpServer(ctx, tt.deps)

			stream, err := client.TranslateAccessToken(ctx)
			if err != nil {
				t.Errorf("AuthServer.TranslateAccessToken() error = %v", err)
				return
			}

			if err := stream.Send(tt.args.req); err != nil {
				t.Errorf("AuthServer.TranslateAccessToken() error = %v", err)
				return
			}

			got, err := stream.Recv()
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthServer.TranslateAccessToken() error = %v", err)
				return
			}

			// Equals false if both are nil or they point to the same memory address
			// so be sure to use seperate variables when providing args in order to prevent SEGV.
			if got != tt.want {
				if !cmp.Equal(got, tt.want, cmpopts.IgnoreUnexported(pb.TranslateAccessTokenRequest{}, pb.TranslateAccessTokenResponse{})) {
					t.Errorf("AuthServer.TranslateAccessToken():\n got = %v\n want = %v\n %v", got, tt.want, cmp.Diff(got, tt.want))
				}
			}
		})
	}
}

func TestAuthServer_GetValidationKeySet(t *testing.T) {
	tests := []struct {
		name    string
		deps    deps
		want    []entity.Key
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			client := setUpServer(ctx, tt.deps)
			stream, err := client.GetValidationKeySet(ctx, &empty.Empty{})
			if err != nil {
				t.Errorf("AuthServer.GetValidationKeySet() error = %v", err)
				return
			}

			// TODO: Make this into an actual test.
			jwk, err := stream.Recv()
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthServer.GetValidationKeySet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			anypb.UnmarshalNew(jwk.GetKey(), proto.UnmarshalOptions{})
		})
	}
}
