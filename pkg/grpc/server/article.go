package server

import (
	"github.com/krixlion/dev_forum-auth/pkg/entity"
	"github.com/krixlion/dev_forum-proto/auth_service/pb"
)

func authFromPB(v *pb.Auth) entity.Token {
	return entity.Token{
		Id:        v.GetId(),
		UserId:    v.GetUserId(),
		Title:     v.GetTitle(),
		Body:      v.GetBody(),
		CreatedAt: v.GetCreatedAt().AsTime(),
		UpdatedAt: v.GetUpdatedAt().AsTime(),
	}
}
