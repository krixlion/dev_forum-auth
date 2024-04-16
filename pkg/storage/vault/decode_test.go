package vault

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/krixlion/dev_forum-auth/pkg/entity"
)

func TestDecodeKey(t *testing.T) {
	type args struct {
		algorithm  entity.Algorithm
		encodedKey string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test if returns an error on unknown algorithm",
			args: args{
				algorithm: "",
			},
			wantErr: true,
		},
		{
			name: "Test if returns an error on unsupported algorithm",
			args: args{
				algorithm:  entity.HS256,
				encodedKey: "test",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := DecodeKey(tt.args.algorithm, tt.args.encodedKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestDecodeRSA(t *testing.T) {
	type args struct {
		encodedKey string
	}
	tests := []struct {
		name    string
		args    args
		want    rsa.PrivateKey
		wantErr bool
	}{
		{
			name: "Test if returns an error on invalid key",
			args: args{
				encodedKey: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeRSA(tt.args.encodedKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeRSA() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			public, ok := got.Public().(*rsa.PublicKey)
			if !ok {
				t.Errorf("DecodeRSA(): public key is not *rsa.PublicKey")
				return
			}

			if !got.Equal(tt.want) || !public.Equal(tt.want.Public()) {
				t.Errorf("DecodeRSA(): keys are not equal\n got = %v\n want = %v\n %v", got, tt.want, cmp.Diff(got, tt.want))
			}
		})
	}
}

func TestDecodeECDSA(t *testing.T) {
	type args struct {
		encodedKey string
	}
	tests := []struct {
		name    string
		args    args
		want    ecdsa.PrivateKey
		wantErr bool
	}{
		{
			name: "Test if returns an error on invalid key",
			args: args{
				encodedKey: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeECDSA(tt.args.encodedKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeECDSA() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			public, ok := got.Public().(*ecdsa.PublicKey)
			if !ok {
				t.Errorf("DecodeECDSA(): public key is not *ecdsa.PublicKey")
				return
			}

			if !got.Equal(tt.want) || !public.Equal(tt.want.Public()) {
				t.Errorf("DecodeECDSA(): keys are not equal\n got = %v\n want = %v\n %v", got, tt.want, cmp.Diff(got, tt.want))
			}
		})
	}
}
