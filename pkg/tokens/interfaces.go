package tokens

import (
	"context"
	"errors"

	"github.com/krixlion/dev_forum-auth/pkg/entity"
)

const DefaultIssuer = "http://auth-service"

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
	ValidateToken(string) error
}

type Translator interface {
	TranslateAccessToken(ctx context.Context, opaqueAccessToken string) (string, error)
}
