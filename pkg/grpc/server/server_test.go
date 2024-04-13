package server_test

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/krixlion/dev_forum-auth/pkg/entity"
	"github.com/krixlion/dev_forum-auth/pkg/grpc/server/servertest"
	pb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1"
	"github.com/krixlion/dev_forum-auth/pkg/storage"
	"github.com/krixlion/dev_forum-auth/pkg/storage/storagemocks"
	"github.com/krixlion/dev_forum-auth/pkg/tokens"
	"github.com/krixlion/dev_forum-auth/pkg/tokens/tokensmocks"
	"github.com/krixlion/dev_forum-lib/filter"
	usermocks "github.com/krixlion/dev_forum-user/pkg/grpc/mocks"
	userPb "github.com/krixlion/dev_forum-user/pkg/grpc/v1"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/anypb"
	empty "google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestAuthServer_SignIn(t *testing.T) {
	type args struct {
		req *pb.SignInRequest
	}
	tests := []struct {
		name    string
		deps    servertest.Deps
		args    args
		want    *pb.SignInResponse
		wantErr bool
	}{
		{
			name: "Test no unexpected errors are returned on valid flow",
			deps: servertest.Deps{
				VerifyClientCert: false,
				Now:              func() time.Time { return time.Unix(0, 0) },
				UserClient: func() usermocks.UserClient {
					m := usermocks.NewUserClient()
					r := &userPb.GetUserSecretRequest{Query: &userPb.GetUserSecretRequest_Email{Email: "test-email"}}
					resp := &userPb.GetUserSecretResponse{
						User: &userPb.User{
							Id:        "test-id",
							Name:      "test-name",
							Email:     "test-email",
							Password:  "$2a$10$QD5AMz7x8T6xvI8QLb7rpuwKTOni6VGInPSxYLm3BEkXbWTjkaw/W", // "test-pass" - hashed with bcrypt, cost 10.
							CreatedAt: timestamppb.New(time.Unix(0, 0)),
							UpdatedAt: timestamppb.New(time.Unix(0, 0)),
						},
					}

					m.On("GetSecret", mock.Anything, r, mock.Anything).Return(resp, nil).Once()
					return m
				}(),
				Storage: func() storagemocks.Storage {
					m := storagemocks.NewStorage()
					tk := entity.Token{
						Id:        "seed",
						UserId:    "test-id",
						Type:      entity.RefreshToken,
						ExpiresAt: time.Unix(0, 0).Add(time.Minute),
						IssuedAt:  time.Unix(0, 0),
					}
					m.On("Create", mock.Anything, tk).Return(nil).Once()
					return m
				}(),
				Vault: storagemocks.NewVault(),
				TokenManager: func() tokensmocks.TokenManager {
					m := tokensmocks.NewTokenManager()
					m.On("GenerateOpaque", tokens.RefreshToken).Return("opaque-refresh-token", "seed", nil).Once()
					return m
				}(),
			},
			args: args{
				req: &pb.SignInRequest{
					Password: "test-pass",
					Email:    "test-email",
				},
			},
			want: &pb.SignInResponse{
				RefreshToken: "opaque-refresh-token",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			client := servertest.NewServer(ctx, tt.deps)

			got, err := client.SignIn(ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthServer.SignIn() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want, cmpopts.IgnoreUnexported(pb.SignInResponse{}, pb.SignInResponse{})) {
				t.Errorf("AuthServer.SignIn():\n got = %v\n want = %v", got, tt.want)
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
		deps    servertest.Deps
		args    args
		wantErr bool
	}{
		{
			name: "Test if no unexpected errors are returned on valid flow",
			deps: servertest.Deps{
				TokenManager: func() tokens.Manager {
					manager := tokensmocks.NewTokenManager()
					manager.On("DecodeOpaque", tokens.RefreshToken, "test-opaque").Return("test-opaque-seed", nil).Once()
					return manager
				}(),
				Storage: func() storage.Storage {
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

			client := servertest.NewServer(ctx, tt.deps)

			_, err := client.SignOut(ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthServer.SignOut() error = %v, wantErr = %v", err, tt.wantErr)
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
		deps    servertest.Deps
		args    args
		want    *pb.GetAccessTokenResponse
		wantErr bool
	}{
		{
			name: "Test if no unexpected errors are returned on valid flow",
			deps: servertest.Deps{
				TokenManager: func() tokens.Manager {
					manager := tokensmocks.NewTokenManager()
					manager.On("DecodeOpaque", tokens.RefreshToken, "test-opaque").Return("test-opaque-decoded", nil).Once()
					manager.On("GenerateOpaque", tokens.AccessToken).Return("test-opaque-generated", "test-opaque-seed", nil).Once()
					return manager
				}(),
				Storage: func() storage.Storage {
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

			client := servertest.NewServer(ctx, tt.deps)

			got, err := client.GetAccessToken(ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthServer.GetAccessToken() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}

			// Equals false if both are nil or they point to the same memory address
			// so be sure to use separate variables when providing args in order to prevent SEGV.
			if got != tt.want {
				if !cmp.Equal(got, tt.want, cmpopts.IgnoreUnexported(pb.GetAccessTokenRequest{}, pb.GetAccessTokenResponse{})) {
					t.Errorf("AuthServer.GetAccessToken():\n got = %v\n want = %v", got, tt.want)
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
		deps    servertest.Deps
		want    *pb.TranslateAccessTokenResponse
		wantErr bool
	}{
		{
			name: "Test if no unexpected errors are returned on valid flow",
			deps: servertest.Deps{
				TokenManager: func() tokens.Manager {
					manager := tokensmocks.NewTokenManager()
					manager.On("DecodeOpaque", tokens.AccessToken, "test-opaque").Return("test-opaque-decoded", nil).Once()
					manager.On("Encode", mock.AnythingOfType("entity.Key"), mock.AnythingOfType("entity.Token")).Return([]byte("test-jwt-encoded"), nil).Once()
					return manager
				}(),
				Storage: func() storage.Storage {
					storage := storagemocks.NewStorage()
					testToken := entity.Token{
						Id:     "test",
						UserId: "test",
						Type:   entity.AccessToken,
					}
					storage.On("Get", mock.Anything, "test-opaque-decoded").Return(testToken, nil).Once()
					return storage
				}(),
				Vault: func() storage.Vault {
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
		{
			name: "Test if returns an error on missing client cert",
			deps: servertest.Deps{
				VerifyClientCert: true,
			},
			args: args{
				req: &pb.TranslateAccessTokenRequest{
					OpaqueAccessToken: "test-opaque",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			client := servertest.NewServer(ctx, tt.deps)

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
			// so be sure to use separate variables when providing args in order to prevent SEGV.
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
		deps    servertest.Deps
		want    []entity.Key
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			client := servertest.NewServer(ctx, tt.deps)
			stream, err := client.GetValidationKeySet(ctx, &empty.Empty{})
			if err != nil {
				t.Errorf("AuthServer.GetValidationKeySet() error = %v", err)
				return
			}

			// TODO: Make this into an actual test.
			jwk, err := stream.Recv()
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthServer.GetValidationKeySet() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}

			anypb.UnmarshalNew(jwk.GetKey(), proto.UnmarshalOptions{})
		})
	}
}
