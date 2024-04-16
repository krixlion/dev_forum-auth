package vault

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"

	"github.com/krixlion/dev_forum-auth/pkg/entity"
	"github.com/krixlion/dev_forum-auth/pkg/grpc/protokey"
)

// DecodeKey decodes provided key with specified algorithm and returns it along with a callback
// that should be used to encode the key to proto message format.
// If decode func for specified algorithm is not found it returns an ErrAlgorithmNotSupported.
// If the algorithm is not recognized it returns an ErrInvalidAlgorithm.
func DecodeKey(algorithm entity.Algorithm, encodedKey string) (crypto.PrivateKey, entity.KeyEncodeFunc, error) {
	switch algorithm {
	case entity.RS256:
		v, err := DecodeRSA(encodedKey)
		if err != nil {
			return nil, nil, err
		}
		return v, protokey.SerializeRSA, nil
	case entity.ES256:
		v, err := DecodeECDSA(encodedKey)
		if err != nil {
			return nil, nil, err
		}
		return v, protokey.SerializeECDSA, nil
	case entity.HS256:
		return nil, nil, ErrAlgorithmNotSupported
	default:
		return nil, nil, ErrInvalidAlgorithm
	}
}

// DecodeRSA decodes RSA PEM block and returns a non-nil err on failure.
func DecodeRSA(rsaPem string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(rsaPem))

	if block == nil {
		return nil, errors.New("failed to decode rsa pem block")
	}

	if block.Type != "RSA PRIVATE KEY" {
		return nil, ErrInvalidKeyType
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// DecodeECDSA decodes EC PEM block and returns a non-nil err on failure.
func DecodeECDSA(ecPem string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(ecPem))

	if block == nil {
		return nil, errors.New("failed to decode ecdsa pem block")
	}

	if block.Type != "EC PRIVATE KEY" {
		return nil, ErrInvalidKeyType
	}

	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}
