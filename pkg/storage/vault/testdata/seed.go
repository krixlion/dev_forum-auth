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
		metadataSlice, err := kvv2.GetVersionsAsList(ctx, path)
		if err != nil {
			return fmt.Errorf("failed to get versions: %w", err)
		}

		versions := make([]int, 0, len(metadataSlice))

		for _, metadata := range metadataSlice {
			versions = append(versions, metadata.Version)
		}

		if err := kvv2.DeleteVersions(ctx, path, versions); err != nil {
			return fmt.Errorf("failed to delete versions: %w", err)
		}

		if err := kvv2.Destroy(ctx, path, versions); err != nil {
			return fmt.Errorf("failed to destroy: %w", err)
		}
	}

	testKeyData := map[string]interface{}{
		"private":   RSAPem,
		"algorithm": string(entity.RS256),
		"keyType":   string(entity.RSA),
	}

	if _, err := kvv2.Put(ctx, Id, testKeyData); err != nil {
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
			return nil, errors.New("failed to parse key")
		}

		path, ok := pathList[0].(string)
		if !ok {
			return nil, errors.New("failed to parse key")
		}

		paths = append(paths, path)
	}

	return paths, nil
}
