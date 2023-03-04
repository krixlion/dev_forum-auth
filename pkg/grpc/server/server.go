package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/krixlion/dev_forum-auth/pkg/storage"
	"github.com/krixlion/dev_forum-lib/event/dispatcher"
	"github.com/krixlion/dev_forum-lib/logging"
	"github.com/krixlion/dev_forum-proto/auth_service/pb"
	userPb "github.com/krixlion/dev_forum-proto/user_service/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	storage    storage.Storage
	services   Services
	logger     logging.Logger
	dispatcher *dispatcher.Dispatcher
}

type Dependencies struct {
	Services
	Storage    storage.Storage
	Logger     logging.Logger
	Dispatcher *dispatcher.Dispatcher
}

type Services struct {
	User userPb.UserServiceClient
}

func NewAuthServer(d Dependencies) AuthServer {
	return AuthServer{
		services:   d.Services,
		storage:    d.Storage,
		logger:     d.Logger,
		dispatcher: d.Dispatcher,
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

	user, err := s.services.User.GetSecret(ctx, &userPb.GetUserSecretRequest{
		Email: email,
	})
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	if user.Password != password {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	accessToken := s.generateAccessToken()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	refreshToken := s.generateRefreshToken()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SignInResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s AuthServer) SignOut(ctx context.Context, req *pb.SignOutRequest) (*pb.Empty, error) {
	accessToken := req.GetAccessToken()
	if err := s.storage.Delete(ctx, accessToken); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return nil, nil
}

func (s AuthServer) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	refreshToken := req.GetRefreshToken()
	if isExpired(refreshToken) {
		return nil, status.Error(codes.PermissionDenied, "Refresh token expired")
	}

	accessToken, err := s.generateAccessToken()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.RefreshTokenResponse{
		AccessToken: accessToken,
	}, nil
}
