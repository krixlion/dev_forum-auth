package vault

import (
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/krixlion/dev_forum-auth/pkg/entity"
	rsapb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1/rsa"
	"google.golang.org/protobuf/proto"
)

// Decode decodes provided key with specified algorithm and returns it along with a callback
// that should be used to encode this key to proto message format.
// If decode func for specified algorithm is not found it returns an ErrAlgorithmNotSupported.
// If the algorithm is not recognized it returns an ErrInvalidAlgorithm.
func Decode(algorithm entity.Algorithm, encodedKey string) (crypto.PrivateKey, entity.KeyEncodeFunc, error) {
	switch algorithm {
	case entity.RS256:
		v, err := DecodeRSA(encodedKey)
		if err != nil {
			return nil, nil, err
		}

		return v, EncodeRSA, nil
	case entity.ES256:
		return nil, nil, ErrAlgorithmNotSupported
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

func EncodeRSA(key crypto.PrivateKey) (proto.Message, error) {
	var privateKey *rsa.PrivateKey

	switch k := key.(type) {
	case *rsa.PrivateKey:
		privateKey = k
	case rsa.PrivateKey:
		privateKey = &k
	default:
		return nil, fmt.Errorf("received invalid key type, expected *rsa.PrivateKey, received %T", key)
	}

	e := make([]byte, 4)
	binary.BigEndian.PutUint32(e, uint32(privateKey.PublicKey.E))

	n := privateKey.PublicKey.N.Bytes()

	message := &rsapb.RSA{
		N: base64.RawURLEncoding.EncodeToString(n),
		E: base64.RawURLEncoding.EncodeToString(e),
	}

	return message, nil
}
