package server

import (
	"context"
	"time"

	empty "github.com/golang/protobuf/ptypes/empty"
	"github.com/krixlion/dev_forum-auth/pkg/entity"
	pb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1"
	"github.com/krixlion/dev_forum-auth/pkg/storage"
	"github.com/krixlion/dev_forum-auth/pkg/tokens"
	"github.com/krixlion/dev_forum-lib/event/dispatcher"
	"github.com/krixlion/dev_forum-lib/logging"
	"github.com/krixlion/dev_forum-lib/tracing"
	userPb "github.com/krixlion/dev_forum-user/pkg/grpc/v1"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	services Services
	// secrets      Secrets
	storage      storage.Storage
	tokenManager tokens.TokenManager
	dispatcher   *dispatcher.Dispatcher
	logger       logging.Logger
	tracer       trace.Tracer
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
		dispatcher: dependencies.Dispatcher,
		logger:     dependencies.Logger,
		tracer:     dependencies.Tracer,
	}
}

func (s AuthServer) Close() error {
	// // A way to wrap multiple err messages from different sources into one.
	// var errMsg string

	// if err := s.storage.Close(); err != nil {
	// 	errMsg = fmt.Sprintf("%s, failed to close storage: %s", errMsg, err)
	// }

	// if errMsg != "" {
	// 	return errors.New(errMsg)
	// }

	// return nil

	return s.storage.Close()
}

func (server AuthServer) SignIn(ctx context.Context, req *pb.SignInRequest) (*pb.SignInResponse, error) {
	ctx, span := server.tracer.Start(ctx, "server.SignIn")
	defer span.End()

	password := req.GetPassword()
	email := req.GetEmail()

	resp, err := server.services.User.GetSecret(ctx, &userPb.GetUserSecretRequest{
		Query: &userPb.GetUserSecretRequest_Email{
			Email: email,
		},
	})
	if err != nil {
		tracing.SetSpanErr(span, err)
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	user := resp.GetUser()

	if err := bcrypt.CompareHashAndPassword([]byte(user.GetPassword()), []byte(password)); err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	encodedOpaqueRefreshToken, tokenId, err := server.tokenManager.GenerateOpaqueToken(tokens.RefreshToken)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	now := time.Now()
	token := entity.Token{
		Id:        tokenId,
		UserId:    user.Id,
		Type:      entity.RefreshToken,
		ExpiresAt: now.Add(server.config.RefreshTokenValidityTime),
		IssuedAt:  now,
	}

	if err := server.storage.Create(ctx, token); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SignInResponse{
		RefreshToken: encodedOpaqueRefreshToken,
	}, nil
}

func (server AuthServer) SignOut(ctx context.Context, req *pb.SignOutRequest) (*empty.Empty, error) {
	ctx, span := server.tracer.Start(ctx, "server.SignOut")
	defer span.End()

	encodedOpaqueRefreshToken := req.GetRefreshToken()

	opaqueRefreshToken, err := server.tokenManager.DecodeOpaqueToken(tokens.RefreshToken, encodedOpaqueRefreshToken)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	if err := server.storage.Delete(ctx, opaqueRefreshToken); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return nil, nil
}

func (server AuthServer) GetAccessToken(ctx context.Context, req *pb.GetAccessTokenRequest) (*pb.GetAccessTokenResponse, error) {
	ctx, span := server.tracer.Start(ctx, "server.GetAccessToken")
	defer span.End()

	opaqueRefreshToken := req.GetRefreshToken()

	refreshTokenId, err := server.tokenManager.DecodeOpaqueToken(tokens.RefreshToken, opaqueRefreshToken)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	refreshToken, err := server.storage.Get(ctx, refreshTokenId)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	now := time.Now()

	opaqueAccessToken, accessTokenId, err := server.tokenManager.GenerateOpaqueToken(tokens.AccessToken)
	if err != nil {
		tracing.SetSpanErr(span, err)
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
	ctx, span := server.tracer.Start(ctx, "server.TranslateAccessToken")
	defer span.End()

	encodedOpaqueAccessToken := req.GetOpaqueAccessToken()

	opaqueAccessToken, err := server.tokenManager.DecodeOpaqueToken(tokens.AccessToken, encodedOpaqueAccessToken)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	token, err := server.storage.Get(ctx, opaqueAccessToken)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	response, err := server.services.User.GetSecret(ctx, &userPb.GetUserSecretRequest{
		Query: &userPb.GetUserSecretRequest_Id{
			Id: token.UserId,
		},
	})
	if err != nil {
		tracing.SetSpanErr(span, err)
		return nil, err
	}

	user := response.GetUser()

	tokenEncoded, err := server.tokenManager.Encode(user.GetPassword(), token)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.TranslateAccessTokenResponse{
		AccessToken: tokenEncoded,
	}, nil
}
