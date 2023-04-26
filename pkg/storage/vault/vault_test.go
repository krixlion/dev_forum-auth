package vault

import (
	"context"
	"crypto/rsa"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/krixlion/dev_forum-auth/pkg/entity"
	"github.com/krixlion/dev_forum-auth/pkg/storage/vault/testdata"
	"github.com/krixlion/dev_forum-lib/env"
	"github.com/krixlion/dev_forum-lib/nulls"
)

func setUpVault() Vault {
	env.Load("app")
	vaultHost := os.Getenv("VAULT_HOST")
	vaultPort := os.Getenv("VAULT_PORT")
	vaultMountPath := os.Getenv("VAULT_MOUNT_PATH")
	vaultToken := os.Getenv("VAULT_TOKEN")
	vaultConfig := Config{
		VaultPath: vaultMountPath,
	}

	vault, err := Make(vaultHost, vaultPort, vaultToken, vaultConfig, nulls.NullTracer{}, nulls.NullLogger{})
	if err != nil {
		panic(err)
	}

	return vault
}

func TestVault_GetKeySet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping vault.GetKeySet integration test...")
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
					Id:        "test",
					Algorithm: entity.RS256,
					Raw: func() *rsa.PrivateKey {
						key, err := DecodeRSA(testdata.RSAPem)
						if err != nil {
							panic(err)
						}
						return key
					}(),
					EncodeFunc: EncodeRSA,
				},
			},
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
		t.Skip("Skipping vault.list integration test...")
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
			name: "Test if retrieves beforehand seeded path",
			args: args{path: "secret"},
			want: []string{"test"},
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
