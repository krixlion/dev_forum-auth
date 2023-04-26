package tokens

import (
	"encoding/base64"
	"errors"
	"hash/crc32"
	"strconv"
	"strings"

	"github.com/krixlion/dev_forum-auth/pkg/entity"
	"github.com/krixlion/dev_forum-lib/str"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/lestrrat-go/jwx/jwt"
)

var (
	ErrInvalidAlgorithm = errors.New("invalid algorithm")
	ErrInvalidToken     = errors.New("invalid token")
	ErrInvalidTokenType = errors.New("invalid token type")
)

// GenerateAccessToken returns an access token following this package's
// specification or a non-nil error on validation failure.
func (m StdTokenManager) Encode(privateKey entity.Key, token entity.Token) ([]byte, error) {
	b := jwt.NewBuilder()
	b.Expiration(token.ExpiresAt)
	b.IssuedAt(token.IssuedAt)
	b.Issuer(m.config.Issuer)
	b.Subject(token.UserId)
	b.JwtID(token.Id)
	b.Claim("type", token.Type)

	jwtoken, err := b.Build()
	if err != nil {
		return nil, err
	}

	headers := jws.NewHeaders()
	if err := headers.Set("kid", privateKey.Id); err != nil {
		return nil, err
	}

	algo, err := verifyAlgorithm(privateKey.Algorithm)
	if err != nil {
		return nil, err
	}

	signedJWT, err := jwt.Sign(jwtoken, algo, privateKey.Raw, jwt.WithHeaders(headers))
	if err != nil {
		return nil, err
	}

	return signedJWT, nil
}

// GenerateOpaque generates an opaque token. It returns an
// encoded token, a random string used as a token's base and an err.
func (StdTokenManager) GenerateOpaque(typ OpaqueTokenPrefix) (string, string, error) {
	randomString, err := str.RandomAlphaString(16)
	if err != nil {
		return "", "", err
	}

	checksum := crc32.ChecksumIEEE([]byte(randomString))
	suffix := strconv.FormatUint(uint64(checksum), 16)

	encoded := base64.URLEncoding.EncodeToString([]byte(randomString + "_" + suffix))

	prefixType, err := typ.String()
	if err != nil {
		return "", "", err
	}

	token := prefixType + "_" + encoded

	return token, randomString, nil
}

// DecodeOpaque decodes a opaque AccessToken and returns a non-nil error if it's invalid.
func (StdTokenManager) DecodeOpaque(typ OpaqueTokenPrefix, encodedOpaqueToken string) (string, error) {
	if len(encodedOpaqueToken) < 4 {
		// Token is shorter than prefix.
		return "", ErrInvalidToken
	}

	typePrefix, err := typ.String()
	if err != nil {
		return "", err
	}

	if prefix := encodedOpaqueToken[:4]; prefix != typePrefix+"_" {
		// Invalid prefix.
		return "", ErrInvalidToken
	}

	// Decode token part without it's prefix.
	return decodeAndValidateOpaque(encodedOpaqueToken[4:])
}

// func verifyTokenType(tokenType string) (entity.TokenType, error) {
// 	if tokenType != string(entity.RefreshToken) && tokenType != string(entity.AccessToken) {
// 		return "", ErrInvalidTokenType
// 	}
// 	return entity.TokenType(tokenType), nil
// }

func verifyAlgorithm(algo entity.Algorithm) (jwa.SignatureAlgorithm, error) {
	switch algo {
	case entity.RS256:
		return jwa.RS256, nil
	case entity.HS256:
		return jwa.HS256, nil
	case entity.ES256:
		return jwa.ES256, nil
	default:
		return "", ErrInvalidAlgorithm
	}
}

// decodeAndValidateOpaque takes a base64 encoded token without it's prefix, decodes and returns it.
// Returns ErrInvalidToken on invalid checksum or length or any errors during decoding.
func decodeAndValidateOpaque(encodedToken string) (string, error) {
	decodedToken, err := base64.URLEncoding.DecodeString(encodedToken)
	if err != nil {
		return "", err
	}

	// Extract token's value and checksum.
	token, extractedChecksum, ok := strings.Cut(string(decodedToken), "_")
	if !ok {
		return "", ErrInvalidToken
	}

	// Convert to HEX for comparison with extracted checksum.
	// Use strconv.FormatUint instead of fmt.Sprintf("%x") for better performance.
	receivedChecksum := strconv.FormatUint(uint64(crc32.ChecksumIEEE([]byte(token))), 16)

	if receivedChecksum != string(extractedChecksum) {
		return "", ErrInvalidToken
	}

	return string(token), nil
}
