package server

import (
	"github.com/krixlion/dev_forum-Entity/pkg/entity"
	"github.com/krixlion/dev_forum-proto/Entity_service/pb"
)

func entityFromPB(v *pb.Entity) entity.Entity {
	return entity.Entity{
		Id:        v.GetId(),
		UserId:    v.GetUserId(),
		Title:     v.GetTitle(),
		Body:      v.GetBody(),
		CreatedAt: v.GetCreatedAt().AsTime(),
		UpdatedAt: v.GetUpdatedAt().AsTime(),
	}
}
