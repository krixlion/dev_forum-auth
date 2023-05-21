package deserialize

import (
	"errors"

	ecPb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1/ec"
	rsaPb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1/rsa"
	"google.golang.org/protobuf/proto"
)

var (
	ErrUnknownMessageType = errors.New("unknown message type")
	ErrKeyNil             = errors.New("key is nil")
)

// Key detects a gRPC format of key and deserializes it using the corresponding function.
func Key(input proto.Message) (interface{}, error) {
	switch msg := input.(type) {
	case *rsaPb.RSA:
		return RSA(msg)
	case *ecPb.EC:
		return ECDSA(msg)
	default:
		return nil, ErrUnknownMessageType
	}
}
