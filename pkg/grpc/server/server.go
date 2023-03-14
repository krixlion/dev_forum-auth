package server

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/krixlion/dev_forum-auth/pkg/entity"
	"github.com/krixlion/dev_forum-auth/pkg/storage"
	"github.com/krixlion/dev_forum-auth/pkg/tokens"
	"github.com/krixlion/dev_forum-lib/event/dispatcher"
	"github.com/krixlion/dev_forum-lib/logging"
	"github.com/krixlion/dev_forum-proto/auth_service/pb"
	userPb "github.com/krixlion/dev_forum-proto/user_service/pb"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	services     Services
	secrets      Secrets
	storage      storage.Storage
	tokenManager tokens.TokenManager
	logger       logging.Logger
	dispatcher   *dispatcher.Dispatcher
	config       Config
}

type Dependencies struct {
	Services
	Storage      storage.Storage
	TokenManager tokens.TokenManager
	Dispatcher   *dispatcher.Dispatcher
	Logger       logging.Logger
	Tracer       trace.Tracer
}

type Services struct {
	User userPb.UserServiceClient
}

type Config struct {
	AccessTokenValidityTime  time.Duration
	RefreshTokenValidityTime time.Duration
}

type Secrets struct {
	// PrivateKey             interface{}
	// PublicKey              interface{}
	UserServiceAccessToken string
}

func NewAuthServer(dependencies Dependencies, config Config, secrets Secrets) AuthServer {
	return AuthServer{
		services:   dependencies.Services,
		storage:    dependencies.Storage,
		logger:     dependencies.Logger,
		dispatcher: dependencies.Dispatcher,
	}
}

func (s AuthServer) Close() error {
	var errMsg string

	err := s.storage.Close()
	if err != nil {
		errMsg = fmt.Sprintf("%s, failed to close storage: %s", errMsg, err)
	}

	if errMsg != "" {
		return errors.New(errMsg)
	}

	return nil
}

func (server AuthServer) SignIn(ctx context.Context, req *pb.SignInRequest) (*pb.SignInResponse, error) {
	password := req.GetPassword()
	email := req.GetEmail()

	resp, err := server.services.User.GetSecret(ctx, &userPb.GetUserSecretRequest{
		Secret: server.secrets.UserServiceAccessToken,
		Query: &userPb.GetUserSecretRequest_Email{
			Email: email,
		},
	})
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	user := resp.GetUser()

	if user.GetPassword() != password {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	encodedOpaqueRefreshToken, tokenId, err := server.tokenManager.GenerateOpaqueToken(tokens.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	token := entity.Token{
		Id:        tokenId,
		UserId:    user.Id,
		Type:      entity.RefreshToken,
		ExpiresAt: time.Now().Add(server.config.RefreshTokenValidityTime),
	}

	if err := server.storage.Create(ctx, token); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SignInResponse{
		RefreshToken: encodedOpaqueRefreshToken,
	}, nil
}

func (server AuthServer) SignOut(ctx context.Context, req *pb.SignOutRequest) (*pb.Empty, error) {
	encodedOpaqueRefreshToken := req.GetRefreshToken()

	opaqueRefreshToken, err := server.tokenManager.DecodeOpaqueToken(tokens.RefreshToken, encodedOpaqueRefreshToken)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	if err := server.storage.Delete(ctx, opaqueRefreshToken); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return nil, nil
}

func (server AuthServer) GetAccessToken(ctx context.Context, req *pb.GetAccessTokenRequest) (*pb.GetAccessTokenResponse, error) {
	opaqueRefreshToken := req.GetRefreshToken()

	refreshTokenId, err := server.tokenManager.DecodeOpaqueToken(tokens.RefreshToken, opaqueRefreshToken)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	refreshToken, err := server.storage.Get(ctx, refreshTokenId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	now := time.Now()

	opaqueAccessToken, accessTokenId, err := server.tokenManager.GenerateOpaqueToken(tokens.AccessToken)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	accessToken := entity.Token{
		Id:        accessTokenId,
		UserId:    refreshToken.UserId,
		Type:      entity.AccessToken,
		ExpiresAt: now.Add(server.config.AccessTokenValidityTime),
		IssuedAt:  now,
	}

	if server.storage.Create(ctx, accessToken); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetAccessTokenResponse{
		AccessToken: opaqueAccessToken,
	}, nil
}

func (server AuthServer) TranslateAccessToken(ctx context.Context, req *pb.TranslateAccessTokenRequest) (*pb.TranslateAccessTokenResponse, error) {
	encodedOpaqueAccessToken := req.GetOpaqueAccessToken()

	opaqueAccessToken, err := server.tokenManager.DecodeOpaqueToken(tokens.AccessToken, encodedOpaqueAccessToken)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	token, err := server.storage.Get(ctx, opaqueAccessToken)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	response, err := server.services.User.GetSecret(ctx, &userPb.GetUserSecretRequest{
		Secret: server.secrets.UserServiceAccessToken,
		Query: &userPb.GetUserSecretRequest_Id{
			Id: token.UserId,
		},
	})
	if err != nil {
		return nil, err
	}

	user := response.GetUser()

	tokenEncoded, err := server.tokenManager.Encode(user.GetPassword(), token)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.TranslateAccessTokenResponse{
		AccessToken: tokenEncoded,
	}, nil
}
