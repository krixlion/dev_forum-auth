package validator

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/krixlion/dev_forum-auth/pkg/tokens"
	"github.com/krixlion/dev_forum-lib/logging"
	"github.com/krixlion/dev_forum-lib/nulls"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
)

var _ tokens.Validator = (*JWTValidator)(nil)

var (
	ErrKeysNotReceived        = errors.New("no keys were received")
	ErrKeySetNotFound         = errors.New("key set not found")
	ErrRefreshFuncNotProvided = errors.New("no refreshFunc was provided to refresh the keyset")
)

type JWTValidator struct {
	// Expected tokens issuer, used to validate JWTs.
	issuer string

	// refreshFunc is used to retrieve a fresh keyset.
	// It's used by TokenValidator to refresh the keyset used for JWT validation
	// each time it fails to find an expected key.
	refreshFunc RefreshFunc

	// clock is used to return current time when validating JWTs.
	// Defaults to time.Now(). Useful for testing.
	clock jwt.Clock

	logger logging.Logger

	// keySetExpired is a channel which notifies when the current keyset is outdated
	keySetExpired chan struct{}

	keySetMutex   sync.RWMutex
	lastRefreshed time.Time
	keySet        jwk.Set
}

type Option interface {
	apply(*JWTValidator)
}

// NewValidator returns a new instance or a non-nil error if provided RefreshFunc is nil.
// If no Clock is provided time.Now() is used by default.
// If no logger is provided then logging is disabled by default.
//
// Make sure to invoke Run() before verifying tokens to start fetching keysets.
func NewValidator(issuer string, refreshFunc RefreshFunc, options ...Option) (*JWTValidator, error) {
	if refreshFunc == nil {
		return nil, ErrRefreshFuncNotProvided
	}

	v := &JWTValidator{
		issuer:        issuer,
		refreshFunc:   refreshFunc,
		keySetExpired: make(chan struct{}, 16),
		keySetMutex:   sync.RWMutex{},
	}

	for _, option := range options {
		option.apply(v)
	}

	if v.clock == nil {
		v.clock = jwt.ClockFunc(time.Now)
	}

	if v.logger == nil {
		v.logger = nulls.NullLogger{}
	}

	return v, nil
}

// Run starts up the validator to refresh the its keySet automatically using its RefreshFunc.
// This function will block until provided context is cancelled or the validator
// fails to fetch a new keyset.
func (validator *JWTValidator) Run(ctx context.Context) {
	// Set keySet on start.
	validator.keySetExpired <- struct{}{}

	for {
		select {
		case <-validator.keySetExpired:
			isTooEarly := validator.lastRefreshed.Sub(validator.clock.Now()) < time.Second
			isNotFirstInit := validator.lastRefreshed != time.Time{}

			if isTooEarly && isNotFirstInit {
				continue
			}

			if err := validator.fetchKeySet(ctx); err != nil {
				validator.logger.Log(ctx, "Failed to fetch a new keyset", "err", err)
			}

		case <-ctx.Done():
			validator.logger.Log(ctx, "Shutting down")
			return
		}
	}
}

// VerifyToken returns a non-nil error if the token is expired,
// signature is invalid or any of the token's claims are different than expected.
// Eg. token was issued in the future or specified 'kid' does not exist.
//
// Note that if the keyset expires, this method will not wait for a new keyset to be fetched
// and instead it will return an error and will continue to do so until
// an updated keyset is successfully retrieved.
func (validator *JWTValidator) VerifyToken(token string) error {
	jwToken, err := jwt.ParseString(token, jwt.WithKeySetProvider(validator.keySetProvider()))
	if err != nil {
		return err
	}

	validateOptions := []jwt.ValidateOption{
		jwt.WithIssuer(validator.issuer),
		jwt.WithClock(validator.clock),
	}

	if err := jwt.Validate(jwToken, validateOptions...); err != nil {
		return err
	}

	if tokenType, ok := jwToken.Get("type"); !ok || tokenType != "access-token" {
		return tokens.ErrInvalidTokenType
	}

	return nil
}

type optionFunc func(*JWTValidator)

func (fn optionFunc) apply(validator *JWTValidator) {
	fn(validator)
}

func WithClock(clock jwt.Clock) Option {
	return optionFunc(func(validator *JWTValidator) {
		validator.clock = clock
	})
}

func WithLogger(logger logging.Logger) Option {
	return optionFunc(func(validator *JWTValidator) {
		validator.logger = logger
	})
}

// fetchKeySet invokes the RefreshFunc and serializes keys into validator's keySet.
// Safe for concurrent use.
func (validator *JWTValidator) fetchKeySet(ctx context.Context) (err error) {
	defer func() {
		err = fmt.Errorf("failed to fetch keyset: %w", err)
	}()

	keys, err := validator.refreshFunc(ctx)
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
	validator.lastRefreshed = validator.clock.Now()

	return nil
}

// keySetProvider returns a callback that safely returns the keyset for the library to use when verifying a JWS.
// Safe for concurrent use.
func (validator *JWTValidator) keySetProvider() jwt.KeySetProvider {
	return jwt.KeySetProviderFunc(func(jwt.Token) (jwk.Set, error) {
		validator.keySetMutex.RLock()
		defer validator.keySetMutex.RUnlock()

		if validator.keySet == nil {
			// Keyset hasn't been fetched yet.
			validator.keySetExpired <- struct{}{}
			return nil, ErrKeySetNotFound
		}

		// Keyset is handled internally and does not need to be derived from, or compared against
		// the token, so it can just be copied so that the lestrrat-go library won't cause a data
		// race when reading keys from it.
		return validator.keySet.Clone()
	})
}
