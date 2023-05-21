package entity

import (
	"crypto"
	"errors"

	"google.golang.org/protobuf/proto"
)

type Key struct {
	Id         string
	Type       KeyType
	Algorithm  Algorithm
	Raw        crypto.PrivateKey
	EncodeFunc KeyEncodeFunc
}

type KeyEncodeFunc func(crypto.PrivateKey) (proto.Message, error)

type KeyType string

const (
	RSA   KeyType = "RSA"
	ECDSA KeyType = "ECDSA"
	HMAC  KeyType = "HMAC"
)

type Algorithm string

const (
	RS256 Algorithm = "RS256"
	ES256 Algorithm = "ES256"
	HS256 Algorithm = "HS256"
)

// Encode is a helper method used to invoke EncodeFunc safely.
func (key Key) Encode() (proto.Message, error) {
	if key.EncodeFunc == nil {
		// Return generic zero value and an error.
		return nil, errors.New("encodeFunc is nil")
	}

	return key.EncodeFunc(key.Raw)
}
