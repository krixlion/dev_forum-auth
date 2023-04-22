// Opaque Tokens are generated from a random string with appended 8 digit
// crc32 hex checksum and encoded in base64 with a prefix depending on their type.
package tokens

import (
	"github.com/krixlion/dev_forum-auth/pkg/entity"
)

type TokenManager interface {
	Encode(privateKey entity.Key, token entity.Token) ([]byte, error)
	GenerateOpaque(typ OpaqueTokenPrefix) (opaqueAccessToken string, seed string, err error)
	DecodeOpaque(typ OpaqueTokenPrefix, encodedOpaqueToken string) (string, error)
}

type OpaqueTokenPrefix int

const (
	// Opaque Refresh tokens are prefixed with "dfr_"
	RefreshToken OpaqueTokenPrefix = iota
	// Opaque Access tokens are prefixed with "dfa_"
	AccessToken
)

func (t OpaqueTokenPrefix) String() (string, error) {
	switch t {
	case RefreshToken:
		return "dfr", nil
	case AccessToken:
		return "dfa", nil
	default:
		return "", ErrInvalidTokenType
	}
}
