package vault

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	vault "github.com/hashicorp/vault/api"
	"github.com/krixlion/dev_forum-auth/pkg/entity"
	"github.com/krixlion/dev_forum-lib/logging"
	"github.com/krixlion/dev_forum-lib/tracing"
	"go.opentelemetry.io/otel/trace"
)

var (
	ErrInvalidKeyType        = errors.New("invalid key type")
	ErrInvalidKeyFormat      = errors.New("invalid key format")
	ErrAlgorithmNotSupported = errors.New("key's algorithm is not supported")
	ErrInvalidAlgorithm      = errors.New("key's algorithm is missing or invalid")
	ErrKeyMissing            = errors.New("key does not contain a private key")
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
	VaultPath string
}

// Make takes in a Token used to connect to Vault and returns a DB instance or a non nil error.
func Make(host, port, token string, config Config, tracer trace.Tracer, logger logging.Logger) (Vault, error) {
	if tracer == nil {
		return Vault{}, errors.New("tracer not provided")
	}

	if logger == nil {
		return Vault{}, errors.New("logger not provided")
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
		vault:  client.KVv2(config.VaultPath),
		tracer: tracer,
		config: config,
		logger: logger,
	}, nil
}

// GetRandom returns a random existing private key from the Vault.
func (db Vault) GetRandom(ctx context.Context) (entity.Key, error) {
	ctx, span := db.tracer.Start(ctx, "vault.GetRandom")
	defer span.End()

	keyPaths, err := db.list(ctx, db.config.VaultPath)
	if err != nil {
		return entity.Key{}, err
	}

	if len(keyPaths) <= 0 {
		return entity.Key{}, errors.New("key not found")
	}

	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(keyPaths))))
	if err != nil {
		return entity.Key{}, err
	}

	randomPath := keyPaths[n.Int64()]

	secret, err := db.vault.Get(ctx, randomPath)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return entity.Key{}, err
	}

	parsed, err := parseSecret(secret)
	if err != nil {
		return entity.Key{}, err
	}

	return makeKey(randomPath, parsed)
}

// GetKeySet returns a slice of keys present in the Vault.
func (db Vault) GetKeySet(ctx context.Context) ([]entity.Key, error) {
	ctx, span := db.tracer.Start(ctx, "vault.GetKeySet")
	defer span.End()

	keyPaths, err := db.list(ctx, db.config.VaultPath)
	if err != nil {
		return nil, err
	}

	keys := []entity.Key{}

	for _, path := range keyPaths {
		secret, err := db.vault.Get(ctx, path)
		if err != nil {
			return nil, err
		}

		parsed, err := parseSecret(secret)
		if err != nil {
			return nil, err
		}

		key, err := makeKey(path, parsed)
		if err != nil {
			return nil, err
		}

		keys = append(keys, key)
	}

	return keys, nil
}

// list returns a slice containing all available paths in the Vault.
// They can be used to retrieve a key from the Vault.
func (db Vault) list(ctx context.Context, mountPath string) ([]string, error) {
	ctx, span := db.tracer.Start(ctx, "vault.list")
	defer span.End()

	secret, err := db.client.Logical().ListWithContext(ctx, mountPath+"/metadata/")
	if err != nil {
		tracing.SetSpanErr(span, err)
		return nil, err
	}

	// Check early in order to avoid unnecessary loop.
	if secret == nil || secret.Data == nil {
		return nil, nil
	}

	paths := make([]string, 0, len(secret.Data))

	for _, pathLists := range secret.Data {
		pathList, ok := pathLists.([]interface{})
		if !ok {
			err := errors.New("failed to parse key")
			tracing.SetSpanErr(span, err)
			return nil, err
		}

		path, ok := pathList[len(pathList)-1].(string)
		if !ok {
			err := errors.New("failed to parse key")
			tracing.SetSpanErr(span, err)
			return nil, err
		}

		paths = append(paths, path)
	}

	return paths, nil
}
