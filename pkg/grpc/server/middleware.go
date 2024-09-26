package server

import (
	"context"
	"net/mail"

	pb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1"
	"github.com/krixlion/dev_forum-lib/tracing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server AuthServer) ValidateRequestInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		switch info.FullMethod {
		case "/auth.AuthService/SignIn":
			return server.validateSignIn(ctx, req.(*pb.SignInRequest), handler)
		case "/auth.AuthService/SignOut":
			return server.validateSignOut(ctx, req.(*pb.SignOutRequest), handler)
		case "/auth.AuthService/GetAccessToken":
			return server.validateGetAccessToken(ctx, req.(*pb.GetAccessTokenRequest), handler)
		case "/auth.AuthService/TranslateAccessToken":
			return server.validateTranslateAccessToken(ctx, req.(*pb.TranslateAccessTokenRequest), handler)
		default:
			return handler(ctx, req)
		}
	}
}

func (server AuthServer) validateSignIn(ctx context.Context, req *pb.SignInRequest, handler grpc.UnaryHandler) (_ interface{}, err error) {
	ctx, span := server.tracer.Start(ctx, "server.validateSignIn")
	defer span.End()
	defer tracing.SetSpanErr(span, err)

	if _, err := mail.ParseAddress(req.GetEmail()); err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	if req.GetPassword() == "" {
		return nil, status.Error(codes.FailedPrecondition, "invalid password")
	}

	return handler(ctx, req)
}

func (server AuthServer) validateSignOut(ctx context.Context, req *pb.SignOutRequest, handler grpc.UnaryHandler) (_ interface{}, err error) {
	ctx, span := server.tracer.Start(ctx, "server.validateSignOut")
	defer span.End()
	defer tracing.SetSpanErr(span, err)

	if req.GetRefreshToken() == "" {
		return nil, status.Error(codes.FailedPrecondition, "invalid refresh token")
	}

	return handler(ctx, req)
}

func (server AuthServer) validateGetAccessToken(ctx context.Context, req *pb.GetAccessTokenRequest, handler grpc.UnaryHandler) (_ interface{}, err error) {
	ctx, span := server.tracer.Start(ctx, "server.validateGetAccessToken")
	defer span.End()
	defer tracing.SetSpanErr(span, err)

	if req.GetRefreshToken() == "" {
		return nil, status.Error(codes.FailedPrecondition, "invalid refresh token")
	}

	return handler(ctx, req)
}

func (server AuthServer) validateTranslateAccessToken(ctx context.Context, req *pb.TranslateAccessTokenRequest, handler grpc.UnaryHandler) (_ interface{}, err error) {
	ctx, span := server.tracer.Start(ctx, "server.validateTranslateAccessToken")
	defer span.End()
	defer tracing.SetSpanErr(span, err)

	if req.GetOpaqueAccessToken() == "" {
		return nil, status.Error(codes.FailedPrecondition, "invalid access token")
	}

	return handler(ctx, req)
}
