package validator

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/krixlion/dev_forum-auth/pkg/tokens"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
)

var (
	ErrKeySetNotFound = errors.New("key set not found")
	ErrRefreshFuncNil = errors.New("RefreshFunc is nil")
)

type JWTValidator struct {
	config Config
	// keySetExpired is a channel which notifies
	// when the current keyset is outdated
	keySetExpired chan struct{}

	keySetMutex sync.Mutex
	keySet      jwk.Set
}

type Config struct {
	// Expected tokens issuer, used to validate JWTs.
	Issuer string

	// Clock is used to return current time when validating JWTs.
	// Useful for testing.
	Clock jwt.Clock

	// RefreshFunc is used to retrieve a fresh keyset.
	// It's used by TokenValidator to refresh the keyset used for JWT validation
	// each time it fails to find an expected key.
	RefreshFunc func(ctx context.Context) ([]Key, error)
}

type RefreshFunc func(ctx context.Context) ([]Key, error)

type Key struct {
	Id        string
	Algorithm string
	Type      string
	Raw       interface{}
}

// MakeValidator returns a new instance
// or a non-nil error if provided RefreshFunc is nil.
// If no Clock is provided in the config time.Now() is used by default.
//
// Make sure to invoke Run() before verifying tokens to start fetching keysets.
func MakeValidator(config Config) (*JWTValidator, error) {
	if config.RefreshFunc == nil {
		return nil, ErrRefreshFuncNil
	}

	if config.Clock == nil {
		config.Clock = jwt.ClockFunc(time.Now)
	}

	v := &JWTValidator{
		config:        config,
		keySetExpired: make(chan struct{}, 16),
		keySetMutex:   sync.Mutex{},
	}
	return v, nil
}

// Run starts validator to refresh keySet automatically using RefreshFunc.
// This function will block until provided context is cancelled or the validator
// fails to fetch a new keyset.
func (validator *JWTValidator) Run(ctx context.Context) error {
	validator.keySetExpired <- struct{}{}

	for {
		select {
		case <-validator.keySetExpired:
			if err := validator.fetchKeySet(ctx); err != nil {
				return err
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// VerifyToken returns a non-nil error if the token is expired,
// signature is invalid or any of the token's claims are different than expected.
// Eg. token was issued in the future or specified 'kid' does not exist.
//
// Note that if the keyset expires, this method will not wait for a new keyset to be fetched
// and instead it will return an error and will continue to do so until
// an updated keyset is succesfully retrieved.
func (validator *JWTValidator) VerifyToken(token string) error {
	jwToken, err := jwt.ParseString(token, jwt.WithKeySetProvider(validator.keySetProvider()))
	if err != nil {
		return err
	}

	validateOptions := []jwt.ValidateOption{
		jwt.WithIssuer(validator.config.Issuer),
		jwt.WithClock(validator.config.Clock),
	}

	if err := jwt.Validate(jwToken, validateOptions...); err != nil {
		return err
	}

	tokenType, ok := jwToken.Get("type")
	if !ok || tokenType != "access-token" {
		return tokens.ErrInvalidTokenType
	}

	return nil
}

// fetchKeySet invokes the RefreshFunc and serializes keys into validator's keySet.
// Safe for concurrent use.
func (validator *JWTValidator) fetchKeySet(ctx context.Context) error {
	keys, err := validator.config.RefreshFunc(ctx)
	if err != nil {
		return err
	}

	keySet, err := keySetFromKeys(keys)
	if err != nil {
		return err
	}

	validator.keySetMutex.Lock()
	defer validator.keySetMutex.Unlock()

	validator.keySet = keySet

	return nil
}

// keySetProvider returns a callback that safely returns the keyset for the library to use when verifying a JWS.
// Safe for concurrent use.
func (validator *JWTValidator) keySetProvider() jwt.KeySetProvider {
	return jwt.KeySetProviderFunc(func(jwt.Token) (jwk.Set, error) {

		validator.keySetMutex.Lock()
		defer validator.keySetMutex.Unlock()

		if validator.keySet == nil {
			// Keyset hasn't been fetched yet.
			validator.keySetExpired <- struct{}{}
			return nil, ErrKeySetNotFound
		}

		// Keyset is handled internally and does not need to be derived from, or compared against
		// the token, so it can just be copied so that the library won't cause a data race when reading
		// keys from it since the keyset is accessed and updated concurrently.
		return validator.keySet.Clone()
	})
}

// keySetFromKeys copies provided keys to a new keyset and returns it.
func keySetFromKeys(keys []Key) (jwk.Set, error) {
	keySet := jwk.NewSet()

	for _, key := range keys {
		jwKey, err := jwk.New(key.Raw)
		if err != nil {
			return nil, err
		}

		if err := jwKey.Set(jwk.KeyIDKey, key.Id); err != nil {
			return nil, err
		}

		if err := jwKey.Set(jwk.KeyTypeKey, key.Type); err != nil {
			return nil, err
		}

		if err := jwKey.Set(jwk.AlgorithmKey, key.Algorithm); err != nil {
			return nil, err
		}

		keySet.Add(jwKey)
	}

	return keySet, nil
}
