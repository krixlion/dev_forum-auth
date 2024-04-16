package testdata

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"

	ecpb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1/ec"
)

type ECKeyPairData struct {
	Id      string
	PrivPem string // Private key as a PEM cert.

	PrivKey *ecdsa.PrivateKey
	PubKey  *ecdsa.PublicKey

	Crv ecpb.ECType
	// Base64RawURL encoded Big-Endian values.
	X string
	Y string
}

var ECDSA ECKeyPairData = ECKeyPairData{
	Id:      "testECDSA",
	PrivPem: "-----BEGIN EC PRIVATE KEY-----\nMHcCAQEEIPvaD0k0SRg9JD/grK0adgk0uP4a2ruhJi5qBUBQ95qLoAoGCCqGSM49\nAwEHoUQDQgAEKONQckRXFo/XksZgsl+5ESQ2/If7MJgaAcqfT16h0bo96XaM59qC\nRcjHeoAygjyzwqVdOjqLzIsC7WEtuMl3lw==\n-----END EC PRIVATE KEY-----",

	Crv: ecpb.ECType_P256,
	X:   `KONQckRXFo_XksZgsl-5ESQ2_If7MJgaAcqfT16h0bo`,
	Y:   `Pel2jOfagkXIx3qAMoI8s8KlXTo6i8yLAu1hLbjJd5c`,
}

func initECDSA() {
	block, _ := pem.Decode([]byte(ECDSA.PrivPem))

	if block == nil {
		err := errors.New("failed to decode ecdsa pem block")
		panic(err)
	}

	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	ECDSA.PrivKey = privateKey
	ECDSA.PubKey = privateKey.Public().(*ecdsa.PublicKey)
}
