package tokens

import "github.com/lestrrat-go/jwx/jwa"

type StdTokenManager struct {
	issuer string
	config Config
}

type Config struct {
	SignatureAlgorithm jwa.SignatureAlgorithm
}

func MakeTokenManager(issuer string, config Config) StdTokenManager {
	return StdTokenManager{
		issuer: issuer,
		config: config,
	}
}
