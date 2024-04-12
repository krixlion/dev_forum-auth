package mocks

import (
	"context"

	pb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/metadata"
)

type AuthStreamClient struct {
	*mock.Mock
}

func NewAuthStreamClient() AuthStreamClient {
	return AuthStreamClient{
		new(mock.Mock),
	}
}

func (m AuthStreamClient) Send(r *pb.TranslateAccessTokenRequest) error {
	returnVals := m.Called(r)
	return returnVals.Error(0)
}

func (m AuthStreamClient) Recv() (*pb.TranslateAccessTokenResponse, error) {
	returnVals := m.Called()
	return returnVals.Get(0).(*pb.TranslateAccessTokenResponse), returnVals.Error(1)
}

func (m AuthStreamClient) CloseSend() error {
	returnVals := m.Called()
	return returnVals.Error(0)
}

func (m AuthStreamClient) Header() (metadata.MD, error) {
	returnVals := m.Called()
	return returnVals.Get(0).(metadata.MD), returnVals.Error(1)
}

func (m AuthStreamClient) Trailer() metadata.MD {
	returnVals := m.Called()
	return returnVals.Get(0).(metadata.MD)
}

func (m AuthStreamClient) Context() context.Context {
	returnVals := m.Called()
	return returnVals.Get(0).(context.Context)
}

func (m AuthStreamClient) SendMsg(msg any) error {
	returnVals := m.Called(msg)
	return returnVals.Error(0)
}

func (m AuthStreamClient) RecvMsg(msg any) error {
	returnVals := m.Called(msg)
	return returnVals.Error(0)
}
