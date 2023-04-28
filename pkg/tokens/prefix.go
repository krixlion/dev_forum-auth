package tokens

type OpaqueTokenPrefix int

const (
	// Opaque Refresh tokens are prefixed with "dfr_"
	RefreshToken OpaqueTokenPrefix = iota
	// Opaque Access tokens are prefixed with "dfa_"
	AccessToken
)

func (t OpaqueTokenPrefix) String() (string, error) {
	switch t {
	case RefreshToken:
		return "dfr", nil
	case AccessToken:
		return "dfa", nil
	default:
		return "", ErrInvalidTokenType
	}
}
