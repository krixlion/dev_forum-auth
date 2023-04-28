// Opaque Tokens are generated from a random string with appended 8 digit
// crc32 hex checksum and encoded in base64 with a prefix depending on their type.
package manager

type StdTokenManager struct {
	config Config
}

type Config struct {
	Issuer string
}

func MakeTokenManager(config Config) StdTokenManager {
	return StdTokenManager{
		config: config,
	}
}
