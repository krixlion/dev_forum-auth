package protokey

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"strconv"

	ecpb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1/ec"
	rsapb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1/rsa"
	"google.golang.org/protobuf/proto"
)

var (
	ErrUnknownMessageType = errors.New("unknown message type")
	ErrKeyNil             = errors.New("key is nil")
)

// DeserializeKey detects a gRPC format of key and deserializes it using the corresponding function.
func DeserializeKey(input proto.Message) (interface{}, error) {
	switch msg := input.(type) {
	case *rsapb.RSA:
		return DeserializeRSA(msg)
	case *ecpb.EC:
		return DeserializeEC(msg)
	default:
		return nil, ErrUnknownMessageType
	}
}

// DeserializeRSA deserializes an RSA message.
func DeserializeRSA(input *rsapb.RSA) (*rsa.PublicKey, error) {
	if input == nil {
		return nil, ErrKeyNil
	}

	n, err := base64.RawURLEncoding.DecodeString(input.GetN())
	if err != nil {
		return nil, err
	}

	N := new(big.Int).SetBytes(n)

	// The short path.
	// The default exponent is usually 65537, so just compare the base64 for [1,0,1] or [0,1,0,1].
	if input.GetE() == "AQAB" || input.GetE() == "AAEAAQ" {
		return &rsa.PublicKey{
			E: 65537,
			N: N,
		}, nil
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(input.GetE())
	if err != nil {
		return nil, err
	}

	E, err := strconv.Atoi(string(eBytes))
	if err != nil {
		return nil, err
	}

	return &rsa.PublicKey{
		N: N,
		E: E,
	}, nil
}

// DeserializeEC deserializes an EC message.
func DeserializeEC(input *ecpb.EC) (*ecdsa.PublicKey, error) {
	if input == nil {
		return nil, ErrKeyNil
	}

	x, err := base64.RawURLEncoding.DecodeString(input.GetX())
	if err != nil {
		return nil, err
	}

	y, err := base64.RawURLEncoding.DecodeString(input.GetY())
	if err != nil {
		return nil, err
	}

	X := new(big.Int).SetBytes(x)
	Y := new(big.Int).SetBytes(y)

	var crv elliptic.Curve
	switch input.GetCrv() {
	case ecpb.ECType_P256:
		crv = elliptic.P256()
	case ecpb.ECType_P384:
		crv = elliptic.P384()
	case ecpb.ECType_P521:
		crv = elliptic.P521()
	case ecpb.ECType_UNDEFINED:
		return nil, errors.New("failed to parse ECDSA key: curve undefined")
	default:
		return nil, errors.New("failed to parse ECDSA key: unexpected curve")
	}

	return &ecdsa.PublicKey{
		Curve: crv,
		X:     X,
		Y:     Y,
	}, nil
}

// SerializeRSA encodes given RSA PublicKey into a supported gRPC message format.
// Returns an error if given key is not of type rsa.PublicKey or a pointer to it.
func SerializeRSA(key crypto.PublicKey) (proto.Message, error) {
	var pubKey *rsa.PublicKey

	switch k := key.(type) {
	case *rsa.PublicKey:
		pubKey = k
	case rsa.PublicKey:
		pubKey = &k
	default:
		return nil, fmt.Errorf("received invalid key type, expected *rsa.PublicKey, received %T", key)
	}

	e := make([]byte, 4)
	binary.BigEndian.PutUint32(e, uint32(pubKey.E))

	n := pubKey.N.Bytes()

	message := &rsapb.RSA{
		N: base64.RawURLEncoding.EncodeToString(n),
		E: base64.RawURLEncoding.EncodeToString(e),
	}

	return message, nil
}

// SerializeECDSA encodes given EC PublicKey into a supported gRPC message format.
// Returns an error if given key is not of type ecdsa.PublicKey or a pointer to it.
func SerializeECDSA(key crypto.PublicKey) (proto.Message, error) {
	if key == nil {
		return nil, errors.New("received nil key")
	}

	var pubKey *ecdsa.PublicKey
	switch k := key.(type) {
	case *ecdsa.PublicKey:
		pubKey = k
	case ecdsa.PublicKey:
		pubKey = &k
	default:
		return nil, fmt.Errorf("received invalid key type, expected *ecdsa.PublicKey, received %T", key)
	}

	x := pubKey.X.Bytes()
	y := pubKey.Y.Bytes()

	var crv ecpb.ECType
	switch pubKey.Curve {
	case elliptic.P256():
		crv = ecpb.ECType_P256
	case elliptic.P384():
		crv = ecpb.ECType_P384
	case elliptic.P521():
		crv = ecpb.ECType_P521
	}

	message := &ecpb.EC{
		Crv: crv,
		X:   base64.RawURLEncoding.EncodeToString(x),
		Y:   base64.RawURLEncoding.EncodeToString(y),
	}

	return message, nil
}
