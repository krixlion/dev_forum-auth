package testdata

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	vault "github.com/hashicorp/vault/api"
	"github.com/krixlion/dev_forum-auth/pkg/entity"
	"github.com/krixlion/dev_forum-lib/env"
)

var ErrFailedToParseKey = errors.New("failed to parse key")

func init() {
	initRSA()
	initECDSA()
}

func Seed() error {
	env.Load("app")

	host := os.Getenv("VAULT_HOST")
	port := os.Getenv("VAULT_PORT")
	mountPath := os.Getenv("VAULT_MOUNT_PATH")
	token := os.Getenv("VAULT_TOKEN")

	client, err := vault.NewClient(&vault.Config{
		Address: fmt.Sprintf("http://%s:%s", host, port),
	})
	if err != nil {
		return err
	}

	client.SetToken(token)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	paths, err := list(ctx, client, mountPath)
	if err != nil {
		return fmt.Errorf("failed to list: %w", err)
	}

	kvv2 := client.KVv2(mountPath)

	for _, path := range paths {
		if err := kvv2.DeleteMetadata(ctx, path); err != nil {
			return fmt.Errorf("failed to delete versions: %w", err)
		}
	}

	testKeyData := map[string]interface{}{
		"private":   RSA.PrivPem,
		"algorithm": string(entity.RS256),
		"keyType":   string(entity.RSA),
	}

	if _, err := kvv2.Put(ctx, RSA.Id, testKeyData); err != nil {
		return fmt.Errorf("failed to put key: %w", err)
	}

	testKeyData = map[string]interface{}{
		"private":   ECDSA.PrivPem,
		"algorithm": string(entity.ES256),
		"keyType":   string(entity.ECDSA),
	}

	if _, err := kvv2.Put(ctx, ECDSA.Id, testKeyData); err != nil {
		return fmt.Errorf("failed to put key: %w", err)
	}

	return nil
}

func list(ctx context.Context, client *vault.Client, path string) ([]string, error) {
	secret, err := client.Logical().ListWithContext(ctx, path+"/metadata/")
	if err != nil {
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
			return nil, err
		}

		for _, path := range pathList {
			path, ok := path.(string)
			if !ok {
				err := ErrFailedToParseKey
				return nil, err
			}

			paths = append(paths, path)
		}
	}

	return paths, nil
}
