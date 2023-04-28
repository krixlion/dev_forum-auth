package testdata

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

const Id = "test"

const (
	N = `qXgUjKWFhoLyUW3KjOiQS7NDie04cF1jHBOGp6ltPfBzx7R_Q7S_WY0Omdh8zquybj-_FAvCbklGV9Gx0YUzFT9RJv6gGC9UPS5aJF2nj-3JwkJh79hTa6fJG8gytD4U79VQP4_51Wo46gkP2wwn-KEfYUGFwoYQk10BxYDWRhU`
	E = `AAEAAQ`
)

const RSAPem = `-----BEGIN RSA PRIVATE KEY-----
MIICWwIBAAKBgQCpeBSMpYWGgvJRbcqM6JBLs0OJ7ThwXWMcE4anqW098HPHtH9D
tL9ZjQ6Z2HzOq7JuP78UC8JuSUZX0bHRhTMVP1Em/qAYL1Q9LlokXaeP7cnCQmHv
2FNrp8kbyDK0PhTv1VA/j/nVajjqCQ/bDCf4oR9hQYXChhCTXQHFgNZGFQIDAQAB
AoGAVtc4uIXNMYuCfqWjKKe34YK/9jrANBw2wFllJB9W4mmH+usMV/aUI2B7/ewI
sKMdMQ+ra6tG+9rCmBfVZgc6kC7kRXh/yrpnviGD/qSqVoY5Cg/7Gef+Ek5OSlvc
ukzdb6hZnq3Uf1KdqAuryh8wX2EFZz9+FDlGJHtzc7Fv6GECQQDpVUZV5wk/UC5m
uKX5BpURW8WjQwz51VK3g5/idbpc91O1unINKsXSWUiImdTHJIdsMz7F85qhYVXZ
ihUTE+9ZAkEAue6UmXeq1vahNMo02ZJrMg2tRkwdH2TmxPQauiJvisAK1w8Jh8E3
b/PbZTACey2TyfG4YVpvP6VXCcLANYFRHQJAZMr3ZSg2MGlcgfcFizsyrZrtFwdh
1ZI29xset96PMJWOTZRKrDFr3t++m3OIHLZE4ZKJbU074LaBNUWWsPUNkQJAY8Dt
ttyuKsCNQr5N1oEow+T0lueVJFfFO9vfTwfUojNgXXty2IPAU28YwVQdsKqGRO1L
x+d2EkaJyPHUn6AuvQJAd0uaFhf0H7GbSDNAVv2QBjEhjVOld63ssbx1Xjk4akRZ
g9oFHdubikYTy8VTA2zmwZcw6GRK7wjqtlVqjABlJA==
-----END RSA PRIVATE KEY-----`

var PrivateRSAKey *rsa.PrivateKey
var PublicRSAKey *rsa.PublicKey

func init() {
	block, _ := pem.Decode([]byte(RSAPem))

	if block == nil {
		err := errors.New("failed to decode rsa pem block")
		panic(err)
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	PrivateRSAKey = privateKey
	PublicRSAKey = privateKey.Public().(*rsa.PublicKey)
}
