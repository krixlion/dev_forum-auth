package tokens

import (
	"errors"

	"github.com/krixlion/dev_forum-auth/pkg/entity"
)

var (
	ErrMalformedToken   = errors.New("malformed token")
	ErrInvalidTokenType = errors.New("invalid token type")
	ErrInvalidAlgorithm = errors.New("invalid algorithm")
)

type Manager interface {
	Encode(privateKey entity.Key, token entity.Token) ([]byte, error)
	GenerateOpaque(typ OpaqueTokenPrefix) (opaqueAccessToken string, seed string, err error)
	DecodeOpaque(typ OpaqueTokenPrefix, encodedOpaqueToken string) (string, error)
}

type Validator interface {
	VerifyToken(string) error
}
