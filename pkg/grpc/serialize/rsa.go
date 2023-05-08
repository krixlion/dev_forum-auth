package serialize

import (
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"math/big"
	"strconv"

	rsapb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1/rsa"
	"google.golang.org/protobuf/proto"
)

var (
	ErrUnknownMessageType = errors.New("unknown message type")
	ErrKeyNil             = errors.New("key is nil")
)

// Key detects a gRPC format of key and de-serializes it using the corresponding function.
func Key(input proto.Message) (interface{}, error) {
	switch msg := input.(type) {
	case *rsapb.RSA:
		return RSA(msg)
	default:
		return nil, ErrUnknownMessageType
	}
}

func RSA(input *rsapb.RSA) (*rsa.PublicKey, error) {
	if input == nil {
		return nil, ErrKeyNil
	}

	n, err := base64.RawURLEncoding.DecodeString(input.GetN())
	if err != nil {
		return nil, err
	}

	N := new(big.Int).SetBytes(n)

	// The short path.
	// The default exponent is usually 65537, so just compare the base64 for [1,0,1] or [0,1,0,1].
	if input.GetE() == "AQAB" || input.GetE() == "AAEAAQ" {
		return &rsa.PublicKey{
			E: 65537,
			N: N,
		}, nil
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(input.GetE())
	if err != nil {
		return nil, err
	}

	E, err := strconv.Atoi(string(eBytes))
	if err != nil {
		return nil, err
	}

	return &rsa.PublicKey{
		N: N,
		E: E,
	}, nil
}
