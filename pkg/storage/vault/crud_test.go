package vault

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/krixlion/dev_forum-auth/pkg/entity"
	"github.com/krixlion/dev_forum-auth/pkg/grpc/protokey"
	"github.com/krixlion/dev_forum-auth/pkg/grpc/protokey/testdata"
	vaultdata "github.com/krixlion/dev_forum-auth/pkg/storage/vault/testdata"
	"github.com/krixlion/dev_forum-lib/env"
	"github.com/krixlion/dev_forum-lib/nulls"
)

func setUpVault() Vault {
	env.Load("app")

	if err := vaultdata.Seed(); err != nil {
		panic(err)
	}

	host := os.Getenv("VAULT_HOST")
	port := os.Getenv("VAULT_PORT")
	mountPath := os.Getenv("VAULT_MOUNT_PATH")
	token := os.Getenv("VAULT_TOKEN")
	config := Config{
		MountPath: mountPath,
	}

	vault, err := Make(host, port, token, config, nulls.NullTracer{}, nulls.NullLogger{})
	if err != nil {
		panic(err)
	}

	return vault
}

func TestVault_GetKeySet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Vault.GetKeySet integration test...")
	}

	tests := []struct {
		name    string
		want    []entity.Key
		wantErr bool
	}{
		{
			name: "Test if retrieves beforehand seeded set",
			want: []entity.Key{
				{
					Id:        testdata.ECDSA.Id,
					Algorithm: entity.ES256,
					Raw: func() *ecdsa.PrivateKey {
						key, err := DecodeECDSA(testdata.ECDSA.PrivPem)
						if err != nil {
							panic(err)
						}
						return key
					}(),
					EncodeFunc: protokey.SerializeECDSA,
				},
				{
					Id:        testdata.RSA.Id,
					Algorithm: entity.RS256,
					Raw: func() *rsa.PrivateKey {
						key, err := DecodeRSA(testdata.RSA.PrivPem)
						if err != nil {
							panic(err)
						}
						return key
					}(),
					EncodeFunc: protokey.SerializeRSA,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setUpVault()

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()

			got, err := db.GetKeySet(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Vault.GetKeySet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !cmp.Equal(got, tt.want, cmpopts.IgnoreFields(entity.Key{}, "EncodeFunc")) {
				t.Errorf("Vault.GetKeySet():\n got = %v\n want = %v\n %v", got, tt.want, cmp.Diff(got, tt.want))
			}
		})
	}
}

func TestVault_list(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Vault.list integration test...")
	}

	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name:    "Test if retrieves beforehand seeded path",
			args:    args{path: "secret"},
			want:    []string{testdata.ECDSA.Id, testdata.RSA.Id},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setUpVault()

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()

			got, err := db.list(ctx, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Vault.list() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !cmp.Equal(got, tt.want) {
				t.Errorf("Vault.list():\n got = %v\n want = %v\n %v", got, tt.want, cmp.Diff(got, tt.want))
			}
		})
	}
}

func TestVault_purge(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Vault.purge integration test...")
	}

	t.Run("Test if Vault.list() returns an empty slice after purge", func(t *testing.T) {
		db := setUpVault()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()

		if err := db.purge(ctx); err != nil {
			t.Errorf("Vault.purge() error = %v", err)
			return
		}

		paths, err := db.list(ctx, db.config.MountPath)
		if err != nil {
			t.Errorf("Vault.purge(): failed to list: error = %v", err)
			return
		}

		if paths != nil {
			t.Errorf("Vault.purge(): list returned paths after purge: %v", paths)
		}
	})
}

func TestVault_create(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Vault.create integration test...")
	}

	type args struct {
		secret secretData
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test if a key is created without errors.",
			args: args{secret: secretData{
				algorithm:  entity.ES256,
				keyType:    entity.ECDSA,
				encodedKey: testdata.ECDSA.PrivPem,
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setUpVault()

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()

			if err := db.create(ctx, tt.args.secret); (err != nil) != tt.wantErr {
				t.Errorf("Vault.create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
