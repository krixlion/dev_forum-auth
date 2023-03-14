package server

import (
	"context"

	"github.com/krixlion/dev_forum-proto/auth_service/pb"
	"google.golang.org/grpc"
)

func (s AuthServer) ValidateRequestInterceptor() grpc.UnaryServerInterceptor {

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		switch info.FullMethod {
		case "/AuthService/SignIn":
			return s.validateSignIn(ctx, req.(*pb.SignInRequest), handler)
		case "/AuthService/SignOut":
			return s.validateSignOut(ctx, req.(*pb.SignOutRequest), handler)
		case "/AuthService/GetAccessToken":
			return s.validateGetAccessToken(ctx, req.(*pb.GetAccessTokenRequest), handler)
		case "/AuthService/TranslateAccessToken":
			return s.validateTranslateAccessToken(ctx, req.(*pb.TranslateAccessTokenRequest), handler)
		default:
			return handler(ctx, req)
		}
	}
}

func (s AuthServer) validateSignIn(ctx context.Context, req *pb.SignInRequest, handler grpc.UnaryHandler) (interface{}, error) {
	return handler(ctx, req)
}

func (s AuthServer) validateSignOut(ctx context.Context, req *pb.SignOutRequest, handler grpc.UnaryHandler) (interface{}, error) {
	return handler(ctx, req)
}

func (s AuthServer) validateGetAccessToken(ctx context.Context, req *pb.GetAccessTokenRequest, handler grpc.UnaryHandler) (interface{}, error) {
	return handler(ctx, req)
}

func (s AuthServer) validateTranslateAccessToken(ctx context.Context, req *pb.TranslateAccessTokenRequest, handler grpc.UnaryHandler) (interface{}, error) {
	return handler(ctx, req)
}
