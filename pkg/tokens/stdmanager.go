package tokens

import "github.com/lestrrat-go/jwx/jwt"

type StdTokenManager struct {
	config Config
}

type Config struct {
	Issuer string
	Clock  jwt.Clock
}

func MakeTokenManager(config Config) StdTokenManager {
	return StdTokenManager{
		config: config,
	}
}
