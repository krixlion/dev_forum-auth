package vault

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"

	"github.com/krixlion/dev_forum-auth/pkg/entity"
	"github.com/krixlion/dev_forum-lib/str"
	"github.com/krixlion/dev_forum-lib/tracing"
)

// GetRandom returns a random existing private key from the Vault.
func (db Vault) GetRandom(ctx context.Context) (entity.Key, error) {
	ctx, span := db.tracer.Start(ctx, "vault.GetRandom")
	defer span.End()

	keyPaths, err := db.list(ctx, db.config.MountPath)
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

	keyPaths, err := db.list(ctx, db.config.MountPath)
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
			err := ErrFailedToParseKey
			tracing.SetSpanErr(span, err)
			return nil, err
		}

		for _, path := range pathList {
			path, ok := path.(string)
			if !ok {
				err := ErrFailedToParseKey
				tracing.SetSpanErr(span, err)
				return nil, err
			}

			paths = append(paths, path)
		}
	}

	return paths, nil
}

// refreshKeys wipes out all keys from the Vault and inserts new randomly
// generated valid keys in amount specified in config.
func (db Vault) refreshKeys(ctx context.Context) (err error) {
	ctx, span := db.tracer.Start(ctx, "vault.refreshKeys")
	defer span.End()

	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to refresh keys: %w", err)
		}
	}()

	if err := db.purge(ctx); err != nil {
		tracing.SetSpanErr(span, err)
		return err
	}

	for i := 0; i < db.config.KeyCount; i++ {
		ECPem, err := generateECDSAPem()
		if err != nil {
			tracing.SetSpanErr(span, err)
			return err
		}

		secretECDSA := secretData{
			algorithm:  entity.ES256,
			keyType:    entity.ECDSA,
			encodedKey: ECPem,
		}

		if err := db.create(ctx, secretECDSA); err != nil {
			tracing.SetSpanErr(span, err)
			return err
		}

		// RSAPem, err := generateRSAPem()
		// if err != nil {
		// 	tracing.SetSpanErr(span, err)
		// 	return err
		// }

		// secretRSA := secretData{
		// 	algorithm:  entity.RS256,
		// 	keyType:    entity.RSA,
		// 	encodedKey: RSAPem,
		// }

		// if err := db.create(ctx, secretRSA); err != nil {
		// 	tracing.SetSpanErr(span, err)
		// 	return err
		// }
	}

	return nil
}

// purge deletes all versions and metadata of all keys in the vault.
func (db Vault) purge(ctx context.Context) error {
	ctx, span := db.tracer.Start(ctx, "vault.purge")
	defer span.End()

	paths, err := db.list(ctx, db.config.MountPath)
	if err != nil {
		err = fmt.Errorf("failed to list keys: %w", err)
		tracing.SetSpanErr(span, err)
		return err
	}

	for _, path := range paths {
		if err := db.vault.DeleteMetadata(ctx, path); err != nil {
			err = fmt.Errorf("failed to delete metadata: %w", err)
			tracing.SetSpanErr(span, err)
			return err
		}
	}

	return nil
}

func (db Vault) create(ctx context.Context, secret secretData) error {
	ctx, span := db.tracer.Start(ctx, "vault.create")
	defer span.End()

	keyData := map[string]interface{}{
		"private":   secret.encodedKey,
		"algorithm": string(secret.algorithm),
		"keyType":   string(secret.keyType),
	}

	id, err := str.RandomAlphaString(50)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return err
	}

	if _, err := db.vault.Put(ctx, id, keyData); err != nil {
		err = fmt.Errorf("failed to create key: %w", err)
		tracing.SetSpanErr(span, err)
		return err
	}

	return nil
}

// func generateRSAPem() (string, error) {
// 	key, err := rsa.GenerateKey(rand.Reader, 4096)
// 	if err != nil {
// 		return "", err
// 	}

// 	pemData := pem.EncodeToMemory(&pem.Block{
// 		Type:  "RSA PRIVATE KEY",
// 		Bytes: x509.MarshalPKCS1PrivateKey(key),
// 	})

// 	return string(pemData), nil
// }

func generateECDSAPem() (string, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", err
	}

	marshaled, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return "", err
	}

	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: marshaled,
	})

	return string(pemData), nil
}
