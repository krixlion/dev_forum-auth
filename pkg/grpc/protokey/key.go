package protokey

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"math/big"
	"strconv"

	ecpb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1/ec"
	rsapb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1/rsa"
	"google.golang.org/protobuf/proto"
)

var (
	ErrUnknownMessageType = errors.New("unknown message type")
	ErrKeyNil             = errors.New("key is nil")
)

// DeserializeKey detects a gRPC format of key and deserializes it using the corresponding function.
func DeserializeKey(input proto.Message) (interface{}, error) {
	switch msg := input.(type) {
	case *rsapb.RSA:
		return DeserializeRSA(msg)
	case *ecpb.EC:
		return DeserializeEC(msg)
	default:
		return nil, ErrUnknownMessageType
	}
}

// DeserializeRSA deserializes an RSA message.
func DeserializeRSA(input *rsapb.RSA) (*rsa.PublicKey, error) {
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

// DeserializeEC deserializes an EC message.
func DeserializeEC(input *ecpb.EC) (*ecdsa.PublicKey, error) {
	if input == nil {
		return nil, ErrKeyNil
	}

	x, err := base64.RawURLEncoding.DecodeString(input.GetX())
	if err != nil {
		return nil, err
	}

	y, err := base64.RawURLEncoding.DecodeString(input.GetY())
	if err != nil {
		return nil, err
	}

	X := new(big.Int).SetBytes(x)
	Y := new(big.Int).SetBytes(y)

	var crv elliptic.Curve
	switch input.GetCrv() {
	case ecpb.ECType_P256:
		crv = elliptic.P256()
	case ecpb.ECType_P384:
		crv = elliptic.P384()
	case ecpb.ECType_P521:
		crv = elliptic.P521()
	case ecpb.ECType_UNDEFINED:
		return nil, errors.New("failed to parse ECDSA key: curve undefined")
	default:
		return nil, errors.New("failed to parse ECDSA key: unexpected curve")
	}

	return &ecdsa.PublicKey{
		Curve: crv,
		X:     X,
		Y:     Y,
	}, nil
}
