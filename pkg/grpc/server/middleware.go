package server

import (
	"context"

	"github.com/krixlion/dev_forum-proto/auth_service/pb"
	"google.golang.org/grpc"
)

func (s AuthServer) ValidateRequestInterceptor() grpc.UnaryServerInterceptor {

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		switch info.FullMethod {
		case "/authService/Create":
			return s.validateCreate(ctx, req.(*pb.CreateauthRequest), handler)
		case "/authService/Update":
			return s.validateUpdate(ctx, req.(*pb.UpdateauthRequest), handler)
		case "/authService/Delete":
			return s.validateDelete(ctx, req.(*pb.DeleteauthRequest), handler)
		default:
			return handler(ctx, req)
		}
	}
}

func (s AuthServer) validateCreate(ctx context.Context, req *pb.CreateauthRequest, handler grpc.UnaryHandler) (interface{}, error) {
	return handler(ctx, req)
}

func (s AuthServer) validateUpdate(ctx context.Context, req *pb.UpdateauthRequest, handler grpc.UnaryHandler) (interface{}, error) {
	return handler(ctx, req)
}

func (s AuthServer) validateDelete(ctx context.Context, req *pb.DeleteauthRequest, handler grpc.UnaryHandler) (interface{}, error) {
	return handler(ctx, req)
}
