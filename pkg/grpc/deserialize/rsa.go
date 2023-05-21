package deserialize

import (
	"crypto/rsa"
	"encoding/base64"
	"math/big"
	"strconv"

	rsapb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1/rsa"
)

// Unserializes an RSA message.
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
