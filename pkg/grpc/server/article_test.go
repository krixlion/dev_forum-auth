package server

// import (
// 	"testing"
// 	"time"

// 	"github.com/google/go-cmp/cmp"
// 	"github.com/google/go-cmp/cmp/cmpopts"
// 	"github.com/krixlion/dev_forum-auth/pkg/entity"
// 	"github.com/krixlion/dev_forum-auth/pkg/helpers/gentest"
// 	"google.golang.org/protobuf/types/known/timestamppb"
// )

// func Test_authFromPB(t *testing.T) {
// 	id := gentest.RandomString(3)
// 	userId := gentest.RandomString(3)
// 	body := gentest.RandomString(3)
// 	title := gentest.RandomString(3)

// 	tests := []struct {
// 		desc string
// 		arg  *pb.auth
// 		want entity.Token
// 	}{
// 		{
// 			desc: "Test if works on simple random data",
// 			arg: &pb.auth{
// 				Id:        id,
// 				UserId:    userId,
// 				Title:     title,
// 				Body:      body,
// 				CreatedAt: timestamppb.New(time.Time{}),
// 				UpdatedAt: timestamppb.New(time.Time{}),
// 			},
// 			want: entity.Token{
// 				Id:        id,
// 				UserId:    userId,
// 				Title:     title,
// 				Body:      body,
// 				CreatedAt: time.Time{},
// 				UpdatedAt: time.Time{},
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.desc, func(t *testing.T) {
// 			got := tokenFromPB(tt.arg)

// 			if !cmp.Equal(got, tt.want, cmpopts.IgnoreUnexported(pb.auth{})) {
// 				t.Errorf("auths are not equal:\n got = %+v\n want = %+v\n", got, tt.want)
// 				return
// 			}
// 		})
// 	}
// }
