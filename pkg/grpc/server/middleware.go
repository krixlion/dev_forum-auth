package server

import (
	"context"

	"github.com/krixlion/dev_forum-proto/Entity_service/pb"
	"google.golang.org/grpc"
)

func (s EntityServer) ValidateRequestInterceptor() grpc.UnaryServerInterceptor {

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		switch info.FullMethod {
		case "/EntityService/Create":
			return s.validateCreate(ctx, req.(*pb.CreateEntityRequest), handler)
		case "/EntityService/Update":
			return s.validateUpdate(ctx, req.(*pb.UpdateEntityRequest), handler)
		case "/EntityService/Delete":
			return s.validateDelete(ctx, req.(*pb.DeleteEntityRequest), handler)
		default:
			return handler(ctx, req)
		}
	}
}

func (s EntityServer) validateCreate(ctx context.Context, req *pb.CreateEntityRequest, handler grpc.UnaryHandler) (interface{}, error) {
	return handler(ctx, req)
}

func (s EntityServer) validateUpdate(ctx context.Context, req *pb.UpdateEntityRequest, handler grpc.UnaryHandler) (interface{}, error) {
	return handler(ctx, req)
}

func (s EntityServer) validateDelete(ctx context.Context, req *pb.DeleteEntityRequest, handler grpc.UnaryHandler) (interface{}, error) {
	return handler(ctx, req)
}
