package tokens

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"hash/crc32"
	"strconv"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/golang-jwt/jwt"
	"github.com/krixlion/dev_forum-auth/pkg/entity"
)

var ErrInvalidToken error = errors.New("invalid token")

func (m TokenManager) Parse(token string) (entity.Token, error) {
	jwToken, err := jwt.Parse(token, m.isValid)
	if err != nil {
		return entity.Token{}, err
	}

	claims, ok := jwToken.Claims.(jwtClaims)
	if !ok {
		return entity.Token{}, ErrInvalidToken
	}

	return entity.Token{
		Id:        claims.Id,
		UserId:    claims.Subject,
		Type:      claims.Type,
		Issuer:    claims.Issuer,
		ExpiresAt: time.Unix(claims.ExpiresAt, 0),
		IssuedAt:  time.Unix(claims.IssuedAt, 0),
	}, nil
}

// GenerateAccessToken returns an access token following this package's
// specification or a non-nil error on validation failure.
func (m TokenManager) Encode(token entity.Token) (string, error) {
	jwtoken := jwt.NewWithClaims(m.config.SigningMethod, jwtClaims{
		Type: token.Type,
		StandardClaims: jwt.StandardClaims{
			Id:        token.Id,
			Issuer:    token.Issuer,
			Subject:   token.UserId,
			IssuedAt:  token.IssuedAt.Unix(),
			ExpiresAt: token.ExpiresAt.Unix(),
		},
	})

	signedJWT, err := jwtoken.SignedString(m.signingKey)
	if err != nil {
		return "", err
	}

	return signedJWT, nil
}

func (TokenManager) GenerateOpaqueToken(prefixType OpaqueTokenPrefix) (string, string, error) {
	randomString, err := randomAlphaString(16)
	if err != nil {
		return "", "", err
	}

	checksum := crc32.ChecksumIEEE([]byte(randomString))
	suffix := strconv.FormatUint(uint64(checksum), 16)

	encoded := base64.StdEncoding.EncodeToString([]byte(randomString + suffix))

	token := prefixType.String() + "_" + encoded

	return token, randomString, nil
}

// DecodeOpaqueToken decodes a opaque AccessToken and returns a non-nil error if it's invalid.
func (m TokenManager) DecodeOpaqueToken(typ OpaqueTokenPrefix, encodedOpaqueToken string) (string, error) {
	// Token is shorter than prefix.
	if len(encodedOpaqueToken) < 4 {
		return "", ErrInvalidToken
	}

	if prefix := encodedOpaqueToken[:4]; prefix != typ.String()+"_" {
		// Invalid prefix.
		return "", ErrInvalidToken
	}

	// Decode token part without it's prefix.
	return m.decodeAndValidateOpaque(encodedOpaqueToken[4:])
}

func (m TokenManager) isValid(t *jwt.Token) (interface{}, error) {
	if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok || t.Method != jwt.SigningMethodHS256 {
		return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
	}

	if err := t.Claims.Valid(); err != nil {
		return nil, err
	}

	return t, nil
}

func (m TokenManager) EncryptAsJWE(data []byte) (string, error) {
	encrypter, err := jose.NewEncrypter(jose.A128GCM, jose.Recipient{
		Algorithm:  jose.RSA_OAEP,
		Key:        m.privateKey,
		PBES2Count: 2048,
	}, nil)
	if err != nil {
		return "", err
	}

	jwe, err := encrypter.Encrypt(data)
	if err != nil {
		return "", err
	}

	return jwe.FullSerialize(), nil
}

func (m TokenManager) DecryptJWE(input string) ([]byte, error) {
	jwe, err := jose.ParseEncrypted(input)
	if err != nil {
		return nil, err
	}

	data, err := jwe.Decrypt(m.publicKey)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// decodeAndValidateOpaque takes a base64 encoded token without it's prefix, decodes it and returns it.
// Returns ErrInvalidToken on invalid checksum or length or any errors during decoding.
func (m TokenManager) decodeAndValidateOpaque(rawToken string) (string, error) {
	decodedToken, err := base64.StdEncoding.DecodeString(rawToken)
	if err != nil {
		return "", err
	}

	offset := len(decodedToken) - 8
	if offset < 0 {
		// Token is too short and does not have a valid checksum.
		return "", ErrInvalidToken
	}

	// Extract token's value and checksum.
	extractedChecksum := decodedToken[offset:]
	token := decodedToken[:offset]

	// Convert to HEX for comparison with extracted checksum.
	// Use strconv.FormatUint instead of fmt.Sprintf("%x") for better performance.
	receivedChecksum := strconv.FormatUint(uint64(crc32.ChecksumIEEE(token)), 16)

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
