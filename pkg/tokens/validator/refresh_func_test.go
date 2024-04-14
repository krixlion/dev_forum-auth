package validator

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/krixlion/dev_forum-auth/pkg/entity"
	"github.com/krixlion/dev_forum-auth/pkg/grpc/server/servertest"
	"github.com/krixlion/dev_forum-auth/pkg/storage/storagemocks"
	"github.com/krixlion/dev_forum-auth/pkg/storage/vault"
	"github.com/krixlion/dev_forum-lib/nulls"
	"github.com/stretchr/testify/mock"
)

func TestDefaultRefreshFunc(t *testing.T) {
	rsaPrivKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("Failed to generate rsa private key: %s", err)
	}
	ecdsaPrivKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate ecdsa private key: %s", err)
	}

	tests := []struct {
		name    string
		deps    servertest.Deps
		want    []Key
		wantErr bool
	}{
		{
			name: "Test keys are parsed and returned as expected on valid flow",
			deps: servertest.Deps{
				Vault: func() storagemocks.Vault {
					keys := []entity.Key{
						{
							Id:         "test-rsa-id",
							Type:       entity.RSA,
							Algorithm:  entity.RS256,
							Raw:        rsaPrivKey,
							EncodeFunc: vault.EncodeRSA,
						},
						{
							Id:         "test-ecdsa-id",
							Type:       entity.ECDSA,
							Algorithm:  entity.ES256,
							Raw:        ecdsaPrivKey,
							EncodeFunc: vault.EncodeECDSA,
						},
					}
					m := storagemocks.NewVault()
					m.On("GetKeySet", mock.Anything).Return(keys, nil)
					return m
				}(),
			},
			want: []Key{
				{
					Id:        "test-rsa-id",
					Algorithm: "RS256",
					Type:      "RSA",
					Raw:       rsaPrivKey.Public(),
				},
				{
					Id:        "test-ecdsa-id",
					Algorithm: "ES256",
					Type:      "ECDSA",
					Raw:       ecdsaPrivKey.Public(),
				},
			},
			wantErr: false,
		},
		{
			name: "Test internal error is forwarded",
			deps: servertest.Deps{
				Vault: func() storagemocks.Vault {
					m := storagemocks.NewVault()
					m.On("GetKeySet", mock.Anything).Return([]entity.Key(nil), errors.New("test err"))
					return m
				}(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()

			fn := DefaultRefreshFunc(servertest.NewServer(ctx, tt.deps), nulls.NullTracer{})
			got, err := fn(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("DefaultRefreshFunc() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}

			if !cmp.Equal(got, tt.want) {
				t.Errorf("DefaultRefreshFunc():\n got = %v\n want = %v", got, tt.want)
			}
		})
	}
}

func Test_keySetFromKeys(t *testing.T) {
	type args struct {
		keys []Key
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test if returns an error on nil keys",
			args: args{
				keys: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := keySetFromKeys(tt.args.keys)
			if (err != nil) != tt.wantErr {
				t.Errorf("keySetFromKeys() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
		})
	}
}
