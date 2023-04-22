package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	vault "github.com/hashicorp/vault/api"
	"github.com/krixlion/dev_forum-auth/pkg/storage/vault/testdata"
	"github.com/krixlion/dev_forum-lib/env"
)

var testKeyData = map[string]interface{}{
	"private":   testdata.RSAPem,
	"algorithm": "RSA",
}

func main() {
	env.Load("app")
	vaultHost := os.Getenv("VAULT_HOST")
	vaultPort := os.Getenv("VAULT_PORT")
	vaultMountPath := os.Getenv("VAULT_MOUNT_PATH")
	vaultToken := os.Getenv("VAULT_TOKEN")

	client, err := vault.NewClient(&vault.Config{
		Address: fmt.Sprintf("http://%s:%s", vaultHost, vaultPort),
	})
	if err != nil {
		log.Fatal(err)
	}

	client.SetToken(vaultToken)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	paths, err := list(ctx, client, vaultMountPath)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to list: %w", err))
	}

	kvv2 := client.KVv2(vaultMountPath)

	for _, path := range paths {
		// Skip root
		if path == "" {
			continue
		}

		metadataSlice, err := kvv2.GetVersionsAsList(ctx, path)
		if err != nil {
			log.Fatal(fmt.Errorf("failed to get versions: %w", err))
		}

		versions := make([]int, len(metadataSlice))

		for _, metadata := range metadataSlice {
			versions = append(versions, metadata.Version)
		}

		if err := kvv2.DeleteVersions(ctx, path, versions); err != nil {
			log.Fatal(fmt.Errorf("failed to delete versions: %w", err))
		}

		if err := kvv2.Destroy(ctx, path, versions); err != nil {
			log.Fatal(fmt.Errorf("failed to destroy: %w", err))
		}
	}

	if _, err := kvv2.Put(ctx, testdata.Id, testKeyData); err != nil {
		log.Fatal(fmt.Errorf("failed to put key: %w", err))
	}
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

	paths := make([]string, len(secret.Data))

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
