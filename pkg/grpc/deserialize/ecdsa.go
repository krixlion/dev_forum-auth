package deserialize

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"errors"
	"math/big"

	ecPb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1/ec"
)

func ECDSA(input *ecPb.EC) (*ecdsa.PublicKey, error) {
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
	case ecPb.ECType_P256:
		crv = elliptic.P256()
	case ecPb.ECType_P384:
		crv = elliptic.P384()
	case ecPb.ECType_P521:
		crv = elliptic.P521()
	case ecPb.ECType_UNDEFINED:
		return nil, errors.New("failed to parse ECDSA key: curve undefined")
	}

	return &ecdsa.PublicKey{
		Curve: crv,
		X:     X,
		Y:     Y,
	}, nil
}
