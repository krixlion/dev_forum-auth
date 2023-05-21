package testdata

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"

	ecPb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1/ec"
)

const ECDSAId = "testECDSA"

const (
	ECDSAPem = `-----BEGIN EC PRIVATE KEY-----
	MHcCAQEEIPvaD0k0SRg9JD/grK0adgk0uP4a2ruhJi5qBUBQ95qLoAoGCCqGSM49
	AwEHoUQDQgAEKONQckRXFo/XksZgsl+5ESQ2/If7MJgaAcqfT16h0bo96XaM59qC
	RcjHeoAygjyzwqVdOjqLzIsC7WEtuMl3lw==
	-----END EC PRIVATE KEY-----`
)

// Base64URL encoded Big-Endian value.
const (
	X = `KONQckRXFo_XksZgsl-5ESQ2_If7MJgaAcqfT16h0bo`
	Y = `Pel2jOfagkXIx3qAMoI8s8KlXTo6i8yLAu1hLbjJd5c`
)

const Crv = ecPb.ECType_P256

var PrivateECDSAKey *ecdsa.PrivateKey
var PublicECDSAKey *ecdsa.PublicKey

func initECDSA() {
	block, _ := pem.Decode([]byte(ECDSAPem))

	if block == nil {
		err := errors.New("failed to decode ecdsa pem block")
		panic(err)
	}

	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	PrivateECDSAKey = privateKey
	PublicECDSAKey = privateKey.Public().(*ecdsa.PublicKey)
}
