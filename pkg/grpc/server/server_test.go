package server_test

import (
	"context"
	"errors"
	"log"
	"net"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/krixlion/dev_forum-auth/pkg/entity"
	"github.com/krixlion/dev_forum-auth/pkg/grpc/server"
	"github.com/krixlion/dev_forum-auth/pkg/helpers/gentest"
	"github.com/krixlion/dev_forum-auth/pkg/storage"
	"github.com/krixlion/dev_forum-lib/mocks"
	"github.com/krixlion/dev_forum-proto/auth_service/pb"
	"github.com/stretchr/testify/mock"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

// setUpServer initializes and runs in the background a gRPC
// server allowing only for local calls for testing.
// Returns a client to interact with the server.
// The server is shutdown when ctx.Done() receives.
func setUpServer(ctx context.Context, mock storage.CQRStorage) pb.AuthServiceClient {
	// bufconn allows the server to call itself
	// great for testing across whole infrastructure
	lis := bufconn.Listen(1024 * 1024)
	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	s := grpc.NewServer()
	server := server.AuthServer{
		Storage: mock,
	}
	pb.RegisterauthServiceServer(s, server)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	go func() {
		<-ctx.Done()
		s.Stop()
	}()

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial bufnet: %v", err)
	}

	client := pb.NewauthServiceClient(conn)
	return client
}

func Test_Get(t *testing.T) {
	v := gentest.Randomauth(2, 5)
	auth := &pb.Auth{
		Id:     v.Id,
		UserId: v.UserId,
		Title:  v.Title,
		Body:   v.Body,
	}

	testCases := []struct {
		desc    string
		arg     *pb.GetauthRequest
		want    *pb.GetauthResponse
		wantErr bool
		storage storage.CQRStorage
	}{
		{
			desc: "Test if response is returned properly on simple request",
			arg: &pb.GetauthRequest{
				authId: auth.Id,
			},
			want: &pb.GetauthResponse{
				auth: auth,
			},
			storage: func() (m mocks.Storage[entity.Token]) {
				m.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(v, nil).Times(1)
				return
			}(),
		},
		{
			desc: "Test if error is returned properly on storage error",
			arg: &pb.GetauthRequest{
				authId: "",
			},
			want:    nil,
			wantErr: true,
			storage: func() (m mocks.Storage[entity.Token]) {
				m.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(entity.Token{}, errors.New("test err")).Times(1)
				return
			}(),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctx, shutdown := context.WithCancel(context.Background())
			defer shutdown()

			client := setUpServer(ctx, tC.storage)

			ctx, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()

			getResponse, err := client.Get(ctx, tC.arg)
			if (err != nil) != tC.wantErr {
				t.Errorf("Failed to Get auth, err: %v", err)
				return
			}

			// Equals false if both are nil or they point to the same memory address
			// so be sure to use seperate structs when providing args in order to prevent SEGV.
			if getResponse != tC.want {
				if !cmp.Equal(getResponse.auth, tC.want.auth, cmpopts.IgnoreUnexported(pb.Auth{})) {
					t.Errorf("auths are not equal:\n Got = %+v\n, want = %+v\n", getResponse.auth, tC.want.auth)
					return
				}
			}
		})
	}
}

func Test_Create(t *testing.T) {
	v := gentest.Randomauth(2, 5)
	auth := &pb.Auth{
		Id:     v.Id,
		UserId: v.UserId,
		Title:  v.Title,
		Body:   v.Body,
	}

	testCases := []struct {
		desc     string
		arg      *pb.CreateauthRequest
		dontWant *pb.CreateauthResponse
		wantErr  bool
		storage  storage.CQRStorage
	}{
		{
			desc: "Test if response is returned properly on simple request",
			arg: &pb.CreateauthRequest{
				auth: auth,
			},
			dontWant: &pb.CreateauthResponse{
				Id: auth.Id,
			},
			storage: func() (m mocks.Storage[entity.Token]) {
				m.On("Create", mock.Anything, mock.AnythingOfType("entity.Auth")).Return(nil).Times(1)
				return
			}(),
		},
		{
			desc: "Test if error is returned properly on storage error",
			arg: &pb.CreateauthRequest{
				auth: auth,
			},
			dontWant: nil,
			wantErr:  true,
			storage: func() (m mocks.Storage[entity.Token]) {
				m.On("Create", mock.Anything, mock.AnythingOfType("entity.Auth")).Return(errors.New("test err")).Times(1)
				return
			}(),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctx, shutdown := context.WithCancel(context.Background())
			defer shutdown()
			client := setUpServer(ctx, tC.storage)

			createResponse, err := client.Create(ctx, tC.arg)
			if (err != nil) != tC.wantErr {
				t.Errorf("Failed to Get auth, err: %v", err)
				return
			}

			// Equals false if both are nil or they point to the same memory address
			// so be sure to use seperate structs when providing args in order to prevent SEGV.
			if createResponse != tC.dontWant {
				if cmp.Equal(createResponse.Id, tC.dontWant.Id) {
					t.Errorf("auth IDs was not reassigned:\n Got = %+v\n want = %+v\n", createResponse.Id, tC.dontWant.Id)
					return
				}
				if _, err := uuid.FromString(createResponse.Id); err != nil {
					t.Errorf("auth ID is not correct UUID:\n ID = %+v\n err = %+v", createResponse.Id, err)
					return
				}
			}
		})
	}
}

func Test_Update(t *testing.T) {
	v := gentest.Randomauth(2, 5)
	auth := &pb.Auth{
		Id:     v.Id,
		UserId: v.UserId,
		Title:  v.Title,
		Body:   v.Body,
	}

	testCases := []struct {
		desc    string
		arg     *pb.UpdateauthRequest
		want    *pb.UpdateauthResponse
		wantErr bool
		storage storage.CQRStorage
	}{
		{
			desc: "Test if response is returned properly on simple request",
			arg: &pb.UpdateauthRequest{
				auth: auth,
			},
			want: &pb.UpdateauthResponse{},
			storage: func() (m mocks.Storage[entity.Token]) {
				m.On("Update", mock.Anything, mock.AnythingOfType("entity.Auth")).Return(nil).Times(1)
				return
			}(),
		},
		{
			desc: "Test if error is returned properly on storage error",
			arg: &pb.UpdateauthRequest{
				auth: auth,
			},
			want:    nil,
			wantErr: true,
			storage: func() (m mocks.Storage[entity.Token]) {
				m.On("Update", mock.Anything, mock.AnythingOfType("entity.Auth")).Return(errors.New("test err")).Times(1)
				return
			}(),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctx, shutdown := context.WithCancel(context.Background())
			defer shutdown()
			client := setUpServer(ctx, tC.storage)

			got, err := client.Update(ctx, tC.arg)
			if (err != nil) != tC.wantErr {
				t.Errorf("Failed to Update auth, err: %v", err)
				return
			}

			// Equals false if both are nil or they point to the same memory address
			// so be sure to use seperate structs when providing args in order to prevent SEGV.
			if got != tC.want {
				if !cmp.Equal(got, tC.want, cmpopts.IgnoreUnexported(pb.UpdateauthResponse{})) {
					t.Errorf("Wrong response:\n got = %+v\n want = %+v\n", got, tC.want)
					return
				}
			}
		})
	}
}

func Test_Delete(t *testing.T) {
	v := gentest.Randomauth(2, 5)
	auth := &pb.Auth{
		Id:     v.Id,
		UserId: v.UserId,
		Title:  v.Title,
		Body:   v.Body,
	}

	testCases := []struct {
		desc    string
		arg     *pb.DeleteauthRequest
		want    *pb.DeleteauthResponse
		wantErr bool
		storage storage.CQRStorage
	}{
		{
			desc: "Test if response is returned properly on simple request",
			arg: &pb.DeleteauthRequest{
				authId: auth.Id,
			},
			want: &pb.DeleteauthResponse{},
			storage: func() (m mocks.Storage[entity.Token]) {
				m.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(nil).Times(1)
				return
			}(),
		},
		{
			desc: "Test if error is returned properly on storage error",
			arg: &pb.DeleteauthRequest{
				authId: auth.Id,
			},
			want:    nil,
			wantErr: true,
			storage: func() (m mocks.Storage[entity.Token]) {
				m.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(errors.New("test err")).Times(1)
				return
			}(),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctx, shutdown := context.WithCancel(context.Background())
			defer shutdown()
			client := setUpServer(ctx, tC.storage)

			got, err := client.Delete(ctx, tC.arg)
			if (err != nil) != tC.wantErr {
				t.Errorf("Failed to Delete auth, err: %v", err)
				return
			}

			// Equals false if both are nil or they point to the same memory address
			// so be sure to use seperate structs when providing args in order to prevent SEGV.
			if got != tC.want {
				if !cmp.Equal(got, tC.want, cmpopts.IgnoreUnexported(pb.DeleteauthResponse{})) {
					t.Errorf("Wrong response:\n got = %+v\n want = %+v\n", got, tC.want)
					return
				}
			}
		})
	}
}

func Test_GetStream(t *testing.T) {
	var auths []entity.Token
	for i := 0; i < 5; i++ {
		auth := gentest.Randomauth(2, 5)
		auths = append(auths, auth)
	}

	var pbauths []*pb.Auth
	for _, v := range auths {
		pbauth := &pb.Auth{
			Id:     v.Id,
			UserId: v.UserId,
			Title:  v.Title,
			Body:   v.Body,
		}
		pbauths = append(pbauths, pbauth)
	}

	testCases := []struct {
		desc    string
		arg     *pb.GetauthsRequest
		want    []*pb.Auth
		wantErr bool
		storage storage.CQRStorage
	}{
		{
			desc: "Test if response is returned properly on simple request",
			arg: &pb.GetauthsRequest{
				Offset: "0",
				Limit:  "5",
			},
			want: pbauths,
			storage: func() (m mocks.Storage[entity.Token]) {
				m.On("GetMultiple", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(auths, nil).Times(1)
				return
			}(),
		},
		{
			desc:    "Test if error is returned properly on storage error",
			arg:     &pb.GetauthsRequest{},
			want:    nil,
			wantErr: true,
			storage: func() (m mocks.Storage[entity.Token]) {
				m.On("GetMultiple", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return([]entity.Token{}, errors.New("test err")).Times(1)
				return
			}(),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctx, shutdown := context.WithCancel(context.Background())
			defer shutdown()
			client := setUpServer(ctx, tC.storage)

			stream, err := client.GetStream(ctx, tC.arg)
			if err != nil {
				t.Errorf("Failed to Get stream, err: %v", err)
				return
			}

			var got []*pb.Auth
			for i := 0; i < len(tC.want); i++ {
				auth, err := stream.Recv()
				if (err != nil) != tC.wantErr {
					t.Errorf("Failed to receive auth from stream, err: %v", err)
					return
				}
				got = append(got, auth)
			}

			if !cmp.Equal(got, tC.want, cmpopts.IgnoreUnexported(pb.Auth{})) {
				t.Errorf("auths are not equal:\n Got = %+v\n want = %+v\n", got, tC.want)
				return
			}
		})
	}
}
