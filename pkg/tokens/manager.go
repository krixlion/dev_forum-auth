// Opaque Tokens are generated from a random string with appended 8 digit
// crc32 hex checksum and encoded in base64 with a prefix depending on their type.
package tokens

import (
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/krixlion/dev_forum-auth/pkg/entity"
)

type OpaqueTokenPrefix int

const (
	// Opaque Refresh tokens are prefixed with "dfr_"
	RefreshToken OpaqueTokenPrefix = iota
	// Opaque Access tokens are prefixed with "dfa_"
	AccessToken OpaqueTokenPrefix = iota
)

func (t OpaqueTokenPrefix) String() string {
	switch t {
	case RefreshToken:
		return "dfr"
	case AccessToken:
		return "dfa"
	default:
		panic("invalid TokenType")
	}
}

type TokenManager struct {
	issuer     string
	signingKey interface{}
	privateKey interface{}
	publicKey  interface{}
	config     Config
}

type Config struct {
	SigningMethod jwt.SigningMethod
}

func MakeTokenManager(issuer string, signingKey, privateKey, publicKey interface{}) TokenManager {
	return TokenManager{
		issuer:     issuer,
		signingKey: signingKey,
		publicKey:  publicKey,
		privateKey: privateKey,
	}
}

type jwtClaims struct {
	Type entity.TokenType `json:"token_type,omitempty"`
	jwt.StandardClaims
}

func (c jwtClaims) Valid() error {
	now := time.Now().Unix()

	if !c.VerifyExpiresAt(now, true) {
		return jwt.NewValidationError("", jwt.ValidationErrorExpired)
	}

	if !c.VerifyIssuedAt(now, true) {
		return jwt.NewValidationError("", jwt.ValidationErrorIssuedAt)
	}

	if !c.VerifyIssuer(entity.Issuer, true) {
		return jwt.NewValidationError("", jwt.ValidationErrorIssuer)
	}

	return nil
}
