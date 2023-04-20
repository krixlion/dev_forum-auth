package tokens

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"hash/crc32"
	"strconv"
	"strings"

	"github.com/krixlion/dev_forum-auth/pkg/entity"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/lestrrat-go/jwx/jwt"
)

var ErrInvalidToken error = errors.New("invalid token")
var ErrInvalidTokenType error = errors.New("invalid token type")

func (m TokenManager) Parse(publicKey interface{}, token []byte) (entity.Token, error) {
	jwToken, err := jwt.Parse(token, jwt.WithVerify(m.config.SignatureAlgorithm, publicKey))
	if err != nil {
		return entity.Token{}, err
	}

	if err := jwt.Validate(jwToken, jwt.WithIssuer(m.issuer)); err != nil {
		return entity.Token{}, err
	}

	tokenType, err := validateTokenType(jwToken)
	if err != nil {
		return entity.Token{}, err
	}

	return entity.Token{
		Id:        jwToken.JwtID(),
		UserId:    jwToken.Subject(),
		Type:      tokenType,
		ExpiresAt: jwToken.Expiration(),
		IssuedAt:  jwToken.IssuedAt(),
	}, nil
}

// GenerateAccessToken returns an access token following this package's
// specification or a non-nil error on validation failure.
func (m TokenManager) Encode(privateKey entity.Key, token entity.Token) ([]byte, error) {
	b := jwt.NewBuilder()
	b.Expiration(token.IssuedAt)
	b.IssuedAt(token.IssuedAt)
	b.Issuer(m.issuer)
	b.Subject(token.UserId)
	b.JwtID(token.Id)

	jwtoken, err := b.Build()
	if err != nil {
		return nil, err
	}

	headers := jws.NewHeaders()
	if err := headers.Set("kid", privateKey.Id); err != nil {
		return nil, err
	}

	signedJWT, err := jwt.Sign(jwtoken, m.config.SignatureAlgorithm, privateKey.Raw, jwt.WithHeaders(headers))
	if err != nil {
		return nil, err
	}

	return signedJWT, nil
}

// GenerateOpaqueToken generates an opaque token. It returns an
// encoded token, a random string used as a token's base and an err.
func (TokenManager) GenerateOpaqueToken(typ OpaqueTokenPrefix) (string, string, error) {
	randomString, err := randomAlphaString(16)
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

// DecodeOpaqueToken decodes a opaque AccessToken and returns a non-nil error if it's invalid.
func (m TokenManager) DecodeOpaqueToken(typ OpaqueTokenPrefix, encodedOpaqueToken string) (string, error) {
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
	return m.decodeAndValidateOpaque(encodedOpaqueToken[4:])
}

func validateTokenType(jwToken jwt.Token) (entity.TokenType, error) {
	typ, ok := jwToken.Get("type")
	if !ok {
		return "", ErrInvalidTokenType
	}

	tokenType, ok := typ.(entity.TokenType)
	if !ok {
		return "", ErrInvalidTokenType
	}

	if tokenType != entity.RefreshToken && tokenType != entity.AccessToken {
		return "", ErrInvalidTokenType
	}
	return tokenType, nil
}

// decodeAndValidateOpaque takes a base64 encoded token without it's prefix, decodes it and returns it.
// Returns ErrInvalidToken on invalid checksum or length or any errors during decoding.
func (m TokenManager) decodeAndValidateOpaque(encodedToken string) (string, error) {
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

func randomAlphaString(length int) (string, error) {
	const (
		letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ" // 52 possibilities
		letterIdxBits = 6                                                      // 6 bits to represent 64 possibilities / indexes
		letterIdxMask = 1<<letterIdxBits - 1                                   // All 1-bits, as many as letterIdxBits
	)
	result := make([]byte, length)
	bufferSize := int(float64(length) * 1.3)

	for i, j, randomBytes := 0, 0, []byte{}; i < length; j++ {
		if j%bufferSize == 0 {
			randomBytes = make([]byte, bufferSize)
			if _, err := rand.Read(randomBytes); err != nil {
				return "", err
			}
		}
		if idx := int(randomBytes[j%length] & letterIdxMask); idx < len(letterBytes) {
			result[i] = letterBytes[idx]
			i++
		}
	}

	return string(result), nil
}
