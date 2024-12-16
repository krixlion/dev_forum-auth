package parser

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/krixlion/dev_forum-auth/pkg/entity"
	"github.com/krixlion/dev_forum-auth/pkg/tokens"
	"github.com/krixlion/dev_forum-lib/event"
	"github.com/krixlion/dev_forum-lib/event/dispatcher"
	"github.com/krixlion/dev_forum-lib/logging"
	"github.com/krixlion/dev_forum-lib/nulls"
	"github.com/krixlion/dev_forum-lib/tracing"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
)

var _ tokens.Parser = (*JWTParser)(nil)
var _ dispatcher.Listener = (*JWTParser)(nil)

var (
	ErrKeysNotReceived        = errors.New("no keys were received")
	ErrKeySetNotFound         = errors.New("key set not found")
	ErrRefreshFuncNotProvided = errors.New("no refreshFunc was provided to refresh the keyset")
)

type JWTParser struct {
	// Expected tokens issuer, used to validate JWTs.
	issuer string

	// refreshFunc is used to refresh the keyset used for JWT
	// validation each time it fails to find an expected key.
	refreshFunc RefreshFunc

	// clock is used to return current time when validating JWTs.
	// Defaults to time.Now(). Useful for testing.
	clock jwt.Clock

	logger logging.Logger

	// keySetExpired is a channel which notifies when the current keyset is outdated.
	// The map contains the event's metadata, e.g TraceId.
	keySetExpired chan map[string]string

	keySetMutex   sync.RWMutex
	lastRefreshed time.Time
	keySet        jwk.Set
}

type Option interface {
	apply(*JWTParser)
}

// NewParser returns a new instance or a non-nil error if provided RefreshFunc is nil.
// If no Clock is provided time.Now() is used by default.
// If no logger is provided then logging is disabled by default.
//
// Make sure to invoke Run() before verifying tokens to start fetching keysets.
func NewParser(issuer string, refreshFunc RefreshFunc, options ...Option) (*JWTParser, error) {
	if refreshFunc == nil {
		return nil, ErrRefreshFuncNotProvided
	}

	v := &JWTParser{
		issuer:        issuer,
		refreshFunc:   refreshFunc,
		keySetExpired: make(chan map[string]string, 1),
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

// Run starts up the parser to refresh its keySet automatically
// using its RefreshFunc. This function will block until provided
// context is cancelled or the parser fails to fetch a new keyset.
func (parser *JWTParser) Run(ctx context.Context) {
	// Set keySet on start.
	parser.keySetExpired <- nil

	for {
		select {
		case metadata := <-parser.keySetExpired:
			isTooEarly := parser.clock.Now().Sub(parser.lastRefreshed) < time.Second
			isNotInit := parser.lastRefreshed != time.Time{}

			if isTooEarly && isNotInit {
				continue
			}

			if err := parser.fetchKeySet(tracing.InjectMetadataIntoContext(ctx, metadata)); err != nil {
				parser.logger.Log(ctx, "Failed to fetch a new keyset", "err", err)
			}

		case <-ctx.Done():
			parser.logger.Log(ctx, "Shutting down JWT parser")
			return
		}
	}
}

// ParseToken returns a non-nil error if the token is expired, signature
// is invalid or any of the token's claims are invalid.
// Eg. token was issued in the future or specified 'kid' does not exist.
//
// Note that if the keyset expires, this method will not wait for a new keyset
// to be fetched and instead it will return an error and it will continue to do
// so until an updated keyset is successfully retrieved.
func (parser *JWTParser) ParseToken(s string) (entity.Token, error) {
	jwToken, err := jwt.ParseString(s, jwt.WithKeySetProvider(parser.keySetProvider()))
	if err != nil {
		return entity.Token{}, err
	}

	validateOptions := []jwt.ValidateOption{
		jwt.WithIssuer(parser.issuer),
		jwt.WithClock(parser.clock),
	}

	if err := jwt.Validate(jwToken, validateOptions...); err != nil {
		return entity.Token{}, err
	}

	if tokenType, ok := jwToken.Get("type"); !ok || tokenType != "access-token" {
		return entity.Token{}, tokens.ErrInvalidTokenType
	}

	return entity.Token{
		Id:        jwToken.JwtID(),
		UserId:    jwToken.Subject(),
		Type:      entity.AccessToken,
		ExpiresAt: jwToken.Expiration(),
		IssuedAt:  jwToken.IssuedAt(),
	}, nil
}

func (parser *JWTParser) EventHandlers() map[event.EventType][]event.Handler {
	return map[event.EventType][]event.Handler{
		event.KeySetUpdated: {
			event.HandlerFunc(func(e event.Event) {
				parser.keySetExpired <- e.Metadata
			}),
		}}
}

type optionFunc func(*JWTParser)

func (fn optionFunc) apply(parser *JWTParser) {
	fn(parser)
}

func WithClock(clock jwt.Clock) Option {
	return optionFunc(func(parser *JWTParser) {
		parser.clock = clock
	})
}

func WithLogger(logger logging.Logger) Option {
	return optionFunc(func(parser *JWTParser) {
		parser.logger = logger
	})
}

// fetchKeySet invokes the RefreshFunc and serializes keys
// into parser's keySet. Safe for concurrent use.
func (parser *JWTParser) fetchKeySet(ctx context.Context) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to fetch keyset: %w", err)
		}
	}()

	keys, err := parser.refreshFunc(ctx)
	if err != nil {
		return err
	}

	keySet, err := keySetFromKeys(keys)
	if err != nil {
		return err
	}

	parser.keySetMutex.Lock()
	defer parser.keySetMutex.Unlock()

	parser.keySet = keySet
	parser.lastRefreshed = parser.clock.Now()

	return nil
}

// keySetProvider returns a callback that safely returns the keyset for
// the library to use when verifying a JWS. Safe for concurrent use.
func (parser *JWTParser) keySetProvider() jwt.KeySetProvider {
	return jwt.KeySetProviderFunc(func(jwt.Token) (jwk.Set, error) {
		parser.keySetMutex.RLock()
		defer parser.keySetMutex.RUnlock()

		if parser.keySet == nil {
			// Keyset hasn't been fetched yet.
			parser.keySetExpired <- nil
			return nil, ErrKeySetNotFound
		}

		// Clone the keyset so that the jwx library won't cause a data
		// race when reading keys from it while they are updated.
		return parser.keySet.Clone()
	})
}
