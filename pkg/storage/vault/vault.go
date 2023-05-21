package vault

import (
	"context"
	"errors"
	"fmt"
	"time"

	vault "github.com/hashicorp/vault/api"
	"github.com/krixlion/dev_forum-lib/logging"
	"github.com/krixlion/dev_forum-lib/nulls"
	"go.opentelemetry.io/otel/trace"
)

var (
	ErrInvalidKeyType        = errors.New("invalid key type")
	ErrInvalidKeyFormat      = errors.New("invalid key format")
	ErrAlgorithmNotSupported = errors.New("key's algorithm is not supported")
	ErrInvalidAlgorithm      = errors.New("key's algorithm is missing or invalid")
	ErrKeyMissing            = errors.New("key does not contain a private key")
	ErrFailedToParseKey      = errors.New("failed to parse key")
)

type Vault struct {
	vault  *vault.KVv2
	client *vault.Client
	config Config
	tracer trace.Tracer
	logger logging.Logger
}

type Config struct {
	// Path in the Vault that the client will mount on.
	MountPath          string
	KeyCount           int
	KeyRefreshInterval time.Duration
}

// Make takes in a Token used to connect to Vault and returns a DB instance or a non nil error.
func Make(host, port, token string, config Config, tracer trace.Tracer, logger logging.Logger) (Vault, error) {
	if tracer == nil {
		tracer = nulls.NullTracer{}
	}

	if logger == nil {
		logger = nulls.NullLogger{}
	}

	client, err := vault.NewClient(&vault.Config{
		Address: fmt.Sprintf("http://%s:%s", host, port),
	})
	if err != nil {
		return Vault{}, err
	}

	client.SetToken(token)

	return Vault{
		client: client,
		vault:  client.KVv2(config.MountPath),
		tracer: tracer,
		config: config,
		logger: logger,
	}, nil
}

// Run blocks until provided context is cancelled.
// When invoked Vault starts to periodically purge the vault and write a new
// set of keys in amount specified in the config.
func (db *Vault) Run(ctx context.Context) {
	// Refresh the vault on start.
	db.logger.Log(ctx, "refreshing keys")

	if err := db.refreshKeys(ctx); err != nil {
		db.logger.Log(ctx, "failed to refresh keys", "err", err)
	}

	ticker := time.NewTicker(db.config.KeyRefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			db.logger.Log(ctx, "refreshing keys")

			if err := db.refreshKeys(ctx); err != nil {
				db.logger.Log(ctx, "failed to refresh keys", "err", err)
			}

		case <-ctx.Done():
			return
		}

	}
}
