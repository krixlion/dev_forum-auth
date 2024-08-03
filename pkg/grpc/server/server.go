package server

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/krixlion/dev_forum-auth/pkg/entity"
	pb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1"
	"github.com/krixlion/dev_forum-auth/pkg/storage"
	"github.com/krixlion/dev_forum-auth/pkg/tokens"
	"github.com/krixlion/dev_forum-lib/cert"
	"github.com/krixlion/dev_forum-lib/filter"
	"github.com/krixlion/dev_forum-lib/logging"
	"github.com/krixlion/dev_forum-lib/tracing"
	userPb "github.com/krixlion/dev_forum-user/pkg/grpc/v1"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	empty "google.golang.org/protobuf/types/known/emptypb"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	services     Services
	vault        storage.Vault
	storage      storage.Storage
	tokenManager tokens.Manager
	logger       logging.Logger
	tracer       trace.Tracer
	config       Config
}

type Dependencies struct {
	Services
	Storage      storage.Storage
	Vault        storage.Vault
	TokenManager tokens.Manager
	Logger       logging.Logger
	Tracer       trace.Tracer
}

type Services struct {
	User userPb.UserServiceClient
}

type Config struct {
	VerifyClientCert         bool
	AccessTokenValidityTime  time.Duration
	RefreshTokenValidityTime time.Duration

	// Allows to override time.Now for testing purposes.
	Now func() time.Time
}

func MakeAuthServer(dependencies Dependencies, config Config) AuthServer {
	s := AuthServer{
		config:       config,
		services:     dependencies.Services,
		storage:      dependencies.Storage,
		vault:        dependencies.Vault,
		tokenManager: dependencies.TokenManager,
		logger:       dependencies.Logger,
		tracer:       dependencies.Tracer,
	}

	if s.config.Now == nil {
		s.config.Now = time.Now
	}

	return s
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

	encodedOpaqueRefreshToken, tokenId, err := server.tokenManager.GenerateOpaque(tokens.RefreshToken)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	now := server.config.Now()
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

	opaqueRefreshToken, err := server.tokenManager.DecodeOpaque(tokens.RefreshToken, encodedOpaqueRefreshToken)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	token, err := server.storage.Get(ctx, opaqueRefreshToken)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	tokens, err := server.storage.GetMultiple(ctx, filter.Filter{{
		Attribute: "user_id",
		Operator:  filter.Equal,
		Value:     token.UserId,
	}})
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	for _, token := range tokens {
		if err := server.storage.Delete(ctx, token.Id); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &empty.Empty{}, nil
}

func (server AuthServer) GetAccessToken(ctx context.Context, req *pb.GetAccessTokenRequest) (*pb.GetAccessTokenResponse, error) {
	ctx, span := server.tracer.Start(ctx, "server.GetAccessToken")
	defer span.End()

	opaqueRefreshToken := req.GetRefreshToken()

	refreshTokenId, err := server.tokenManager.DecodeOpaque(tokens.RefreshToken, opaqueRefreshToken)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	refreshToken, err := server.storage.Get(ctx, refreshTokenId)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	opaqueAccessToken, accessTokenId, err := server.tokenManager.GenerateOpaque(tokens.AccessToken)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	now := server.config.Now()
	accessToken := entity.Token{
		Id:        accessTokenId,
		UserId:    refreshToken.UserId,
		Type:      entity.AccessToken,
		ExpiresAt: now.Add(server.config.AccessTokenValidityTime),
		IssuedAt:  now,
	}

	if err := server.storage.Create(ctx, accessToken); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetAccessTokenResponse{
		AccessToken: opaqueAccessToken,
	}, nil
}

func (server AuthServer) TranslateAccessToken(stream pb.AuthService_TranslateAccessTokenServer) error {
	ctx := stream.Context()

	if server.config.VerifyClientCert {
		if err := cert.VerifyClientTLS(ctx, "gateway"); err != nil {
			return fmt.Errorf("failed to verify client cert: %w", err)
		}
	}

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		response, err := server.translateAccessToken(ctx, req)
		if err != nil {
			return err
		}

		if err := stream.Send(response); err != nil {
			return err
		}
	}
}

func (server AuthServer) translateAccessToken(ctx context.Context, req *pb.TranslateAccessTokenRequest) (*pb.TranslateAccessTokenResponse, error) {
	ctx, span := server.tracer.Start(tracing.InjectMetadataIntoContext(ctx, req.Metadata), "server.TranslateAccessToken")
	defer span.End()

	encodedOpaqueAccessToken := req.GetOpaqueAccessToken()

	opaqueAccessToken, err := server.tokenManager.DecodeOpaque(tokens.AccessToken, encodedOpaqueAccessToken)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	token, err := server.storage.Get(ctx, opaqueAccessToken)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	privateKey, err := server.vault.GetRandom(ctx)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	tokenEncoded, err := server.tokenManager.Encode(privateKey, token)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.TranslateAccessTokenResponse{
		AccessToken: string(tokenEncoded),
		Metadata:    tracing.ExtractMetadataFromContext(ctx),
	}, nil
}

func (server AuthServer) GetValidationKeySet(_ *empty.Empty, stream pb.AuthService_GetValidationKeySetServer) error {
	ctx := stream.Context()

	ctx, span := server.tracer.Start(ctx, "server.GetValidationKeySet")
	defer span.End()

	keys, err := server.vault.GetKeySet(ctx)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return err
	}

	for _, key := range keys {
		if err := ctx.Err(); err != nil {
			tracing.SetSpanErr(span, err)
			return err
		}

		encoded, err := key.Encode()
		if err != nil {
			tracing.SetSpanErr(span, err)
			return err
		}

		marshaledKey, err := anypb.New(encoded)
		if err != nil {
			tracing.SetSpanErr(span, err)
			return err
		}

		jwk := &pb.Jwk{
			Kid: key.Id,
			Kty: string(key.Type),
			Alg: string(key.Algorithm),
			Key: marshaledKey,
		}

		if err := stream.Send(jwk); err != nil {
			tracing.SetSpanErr(span, err)
			return err
		}
	}

	return nil
}
