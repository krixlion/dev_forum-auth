package vault

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/krixlion/dev_forum-auth/pkg/entity"
	ecpb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1/ec"
	rsapb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1/rsa"
	"github.com/krixlion/dev_forum-auth/pkg/storage/vault/testdata"
	"google.golang.org/protobuf/proto"
)

func TestDecode(t *testing.T) {
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
			_, _, err := Decode(tt.args.algorithm, tt.args.encodedKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
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

func TestEncodeRSA(t *testing.T) {
	type args struct {
		key crypto.PrivateKey
	}
	tests := []struct {
		name    string
		args    args
		want    proto.Message
		wantErr bool
	}{
		{
			name: "Test if valid RSA private key is marshaled into correct public key",
			args: args{
				key: testdata.RSA.PrivKey,
			},
			want: &rsapb.RSA{
				N: testdata.RSA.N,
				E: testdata.RSA.E,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EncodeRSA(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeRSA() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want, cmpopts.IgnoreUnexported(rsapb.RSA{})) {
				t.Errorf("EncodeRSA() = %v, want %v", got, tt.want)
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

func TestEncodeECDSA(t *testing.T) {
	type args struct {
		key crypto.PrivateKey
	}
	tests := []struct {
		name    string
		args    args
		want    proto.Message
		wantErr bool
	}{
		{
			name: "Test if valid ECDSA private key is marshaled into correct public key",
			args: args{
				key: testdata.ECDSA.PrivKey,
			},
			want: &ecpb.EC{
				Crv: testdata.ECDSA.Crv,
				X:   testdata.ECDSA.X,
				Y:   testdata.ECDSA.Y,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EncodeECDSA(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeECDSA() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want, cmpopts.IgnoreUnexported(ecpb.EC{})) {
				t.Errorf("EncodeECDSA():\n got = %v\n want = %v", got, tt.want)
			}
		})
	}
}
