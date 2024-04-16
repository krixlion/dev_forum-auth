package testdata

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

type RSAKeyPairData struct {
	Id      string
	PrivPem string // Private key as a PEM cert.

	PrivKey *rsa.PrivateKey
	PubKey  *rsa.PublicKey

	// Base64RawURL encoded Big-Endian values.
	N string
	E string
}

var RSA RSAKeyPairData = RSAKeyPairData{
	Id:      "testRSA",
	PrivPem: "-----BEGIN RSA PRIVATE KEY-----\nMIICWwIBAAKBgQCpeBSMpYWGgvJRbcqM6JBLs0OJ7ThwXWMcE4anqW098HPHtH9D\ntL9ZjQ6Z2HzOq7JuP78UC8JuSUZX0bHRhTMVP1Em/qAYL1Q9LlokXaeP7cnCQmHv\n2FNrp8kbyDK0PhTv1VA/j/nVajjqCQ/bDCf4oR9hQYXChhCTXQHFgNZGFQIDAQAB\nAoGAVtc4uIXNMYuCfqWjKKe34YK/9jrANBw2wFllJB9W4mmH+usMV/aUI2B7/ewI\nsKMdMQ+ra6tG+9rCmBfVZgc6kC7kRXh/yrpnviGD/qSqVoY5Cg/7Gef+Ek5OSlvc\nukzdb6hZnq3Uf1KdqAuryh8wX2EFZz9+FDlGJHtzc7Fv6GECQQDpVUZV5wk/UC5m\nuKX5BpURW8WjQwz51VK3g5/idbpc91O1unINKsXSWUiImdTHJIdsMz7F85qhYVXZ\nihUTE+9ZAkEAue6UmXeq1vahNMo02ZJrMg2tRkwdH2TmxPQauiJvisAK1w8Jh8E3\nb/PbZTACey2TyfG4YVpvP6VXCcLANYFRHQJAZMr3ZSg2MGlcgfcFizsyrZrtFwdh\n1ZI29xset96PMJWOTZRKrDFr3t++m3OIHLZE4ZKJbU074LaBNUWWsPUNkQJAY8Dt\nttyuKsCNQr5N1oEow+T0lueVJFfFO9vfTwfUojNgXXty2IPAU28YwVQdsKqGRO1L\nx+d2EkaJyPHUn6AuvQJAd0uaFhf0H7GbSDNAVv2QBjEhjVOld63ssbx1Xjk4akRZ\ng9oFHdubikYTy8VTA2zmwZcw6GRK7wjqtlVqjABlJA==\n-----END RSA PRIVATE KEY-----",

	N: `qXgUjKWFhoLyUW3KjOiQS7NDie04cF1jHBOGp6ltPfBzx7R_Q7S_WY0Omdh8zquybj-_FAvCbklGV9Gx0YUzFT9RJv6gGC9UPS5aJF2nj-3JwkJh79hTa6fJG8gytD4U79VQP4_51Wo46gkP2wwn-KEfYUGFwoYQk10BxYDWRhU`,
	E: `AAEAAQ`,
}

func initRSA() {
	block, _ := pem.Decode([]byte(RSA.PrivPem))

	if block == nil {
		err := errors.New("failed to decode rsa pem block")
		panic(err)
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	RSA.PrivKey = privateKey
	RSA.PubKey = privateKey.Public().(*rsa.PublicKey)
}
