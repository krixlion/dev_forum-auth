package server

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
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
	AccessTokenLength       int
	AccessTokenValidityTime time.Duration
	RefreshTokenExpiration  time.Duration
}

type Secrets struct {
	SigningKey             interface{}
	PrivateKey             interface{}
	PublicKey              interface{}
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

func (s AuthServer) SignIn(ctx context.Context, req *pb.SignInRequest) (*pb.SignInResponse, error) {
	password := req.GetPassword()
	email := req.GetEmail()

	resp, err := s.services.User.GetSecret(ctx, &userPb.GetUserSecretRequest{
		Secret: s.secrets.UserServiceAccessToken,
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

	id, err := uuid.NewV4()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	token := entity.Token{
		Id:        id.String(),
		UserId:    user.Id,
		Type:      entity.RefreshToken,
		Issuer:    entity.Issuer,
		ExpiresAt: time.Now().Add(s.config.RefreshTokenExpiration),
	}

	encodedOpaqueRefreshToken, opaqueRefreshToken, err := s.tokenManager.GenerateOpaqueToken(tokens.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := s.storage.Create(ctx, opaqueRefreshToken, token); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SignInResponse{
		RefreshToken: encodedOpaqueRefreshToken,
	}, nil
}

func (s AuthServer) SignOut(ctx context.Context, req *pb.SignOutRequest) (*pb.Empty, error) {
	encodedOpaqueRefreshToken := req.GetRefreshToken()

	refreshOpaqueToken, err := s.tokenManager.DecodeOpaqueToken(tokens.RefreshToken, encodedOpaqueRefreshToken)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	if err := s.storage.Delete(ctx, refreshOpaqueToken); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return nil, nil
}

func (s AuthServer) GetAccessToken(ctx context.Context, req *pb.GetAccessTokenRequest) (*pb.GetAccessTokenResponse, error) {
	encodedOpaqueRefreshToken := req.GetRefreshToken()

	refreshOpaqueToken, err := s.tokenManager.DecodeOpaqueToken(tokens.RefreshToken, encodedOpaqueRefreshToken)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	refreshToken, err := s.storage.Get(ctx, refreshOpaqueToken)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	id, err := uuid.NewV4()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	now := time.Now()

	accessToken := entity.Token{
		Id:        id.String(),
		UserId:    refreshToken.UserId,
		Type:      entity.AccessToken,
		Issuer:    entity.Issuer,
		ExpiresAt: now.Add(s.config.AccessTokenValidityTime),
		IssuedAt:  now,
	}

	encodedOpaqueAccessToken, opaqueAccessToken, err := s.tokenManager.GenerateOpaqueToken(tokens.AccessToken)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if s.storage.Create(ctx, opaqueAccessToken, accessToken); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetAccessTokenResponse{
		AccessToken: encodedOpaqueAccessToken,
	}, nil
}

func (s AuthServer) TranslateAccessToken(ctx context.Context, req *pb.TranslateAccessTokenRequest) (*pb.TranslateAccessTokenResponse, error) {
	encodedOpaqueAccessToken := req.GetOpaqueAccessToken()

	opaqueAccessToken, err := s.tokenManager.DecodeOpaqueToken(tokens.AccessToken, encodedOpaqueAccessToken)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	token, err := s.storage.Get(ctx, opaqueAccessToken)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	tokenEncoded, err := s.tokenManager.Encode(token)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.TranslateAccessTokenResponse{
		AccessToken: tokenEncoded,
	}, nil
}
