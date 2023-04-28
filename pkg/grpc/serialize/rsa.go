package serialize

import (
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"math/big"
	"strconv"

	rsapb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1/rsa"
)

var (
	ErrKeyNil = errors.New("key is nil")
)

func RSA(input *rsapb.RSA) (rsa.PublicKey, error) {
	if input == nil {
		return rsa.PublicKey{}, ErrKeyNil
	}

	// decode the base64 bytes for n
	n, err := base64.RawURLEncoding.DecodeString(input.GetN())
	if err != nil {
		return rsa.PublicKey{}, err
	}

	N := new(big.Int).SetBytes(n)

	// The default exponent is usually 65537, so just compare the
	// base64 for [1,0,1] or [0,1,0,1].
	if input.GetE() == "AQAB" || input.GetE() == "AAEAAQ" {
		return rsa.PublicKey{
			E: 65537,
			N: N,
		}, nil
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(input.GetE())
	if err != nil {
		return rsa.PublicKey{}, err
	}

	E, err := strconv.Atoi(string(eBytes))
	if err != nil {
		return rsa.PublicKey{}, err
	}

	return rsa.PublicKey{
		N: N,
		E: E,
	}, nil
}
