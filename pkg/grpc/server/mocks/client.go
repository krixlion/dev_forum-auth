package mocks

import (
	"context"

	pb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AuthClient struct {
	*mock.Mock
}

func NewAuthClient() AuthClient {
	return AuthClient{
		new(mock.Mock),
	}
}

// Upon succesful login user receives a refresh_token.
// When it expires or is revoked user has to login again.
func (m AuthClient) SignIn(ctx context.Context, in *pb.SignInRequest, opts ...grpc.CallOption) (*pb.SignInResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.SignInResponse), args.Error(1)
}

// SignOut revokes user's active refresh_token.
func (m AuthClient) SignOut(ctx context.Context, in *pb.SignOutRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m AuthClient) GetAccessToken(ctx context.Context, in *pb.GetAccessTokenRequest, opts ...grpc.CallOption) (*pb.GetAccessTokenResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.GetAccessTokenResponse), args.Error(1)
}

func (m AuthClient) GetValidationKeySet(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (pb.AuthService_GetValidationKeySetClient, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(pb.AuthService_GetValidationKeySetClient), args.Error(1)
}

func (m AuthClient) TranslateAccessToken(ctx context.Context, opts ...grpc.CallOption) (pb.AuthService_TranslateAccessTokenClient, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(pb.AuthService_TranslateAccessTokenClient), args.Error(1)
}
