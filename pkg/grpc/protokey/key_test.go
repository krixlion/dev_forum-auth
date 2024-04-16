package protokey

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/krixlion/dev_forum-auth/pkg/grpc/protokey/testdata"
	ecpb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1/ec"
	rsapb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1/rsa"
	"google.golang.org/protobuf/proto"
)

func TestDeserializeRSA(t *testing.T) {
	type args struct {
		input *rsapb.RSA
	}
	tests := []struct {
		name    string
		args    args
		want    *rsa.PublicKey
		wantErr bool
	}{
		{
			name: "Test a valid public key is correctly deserialized",
			args: args{
				input: &rsapb.RSA{
					E: "AQAB",
					N: "sZGVa39dSmJ5c7mbOsJZaq62MVjPD3xNPb-Aw3VJznk6piF5GGgdMoQmAjNmANVBBpPUyQU2SEHgXQvp6j52E662umdV2xU-1ETzn2dW23jtdTFPHRG4BFZz7m14MXX9i0QqgWVnTRy-DD5VITkFZvBqCEzWjT_y47DYD2Dod-U",
				},
			},
			want: &rsa.PublicKey{
				E: 65537,
				// Big Endian 124692971944797177402996703053303877641609106436730124136075828918287037758927191447826707233876916396730936365584704201525802806009892366608834910101419219957891196104538322266555160652329444921468362525907130134965311064068870381940624996449410632960760491317833379253431879193412822078872504618021680609253
				N: new(big.Int).SetBytes([]byte{0xB1, 0x91, 0x95, 0x6B, 0x7F, 0x5D, 0x4A, 0x62, 0x79, 0x73, 0xB9, 0x9B, 0x3A, 0xC2, 0x59, 0x6A, 0xAE, 0xB6, 0x31, 0x58, 0xCF, 0x0F, 0x7C, 0x4D, 0x3D, 0xBF, 0x80, 0xC3, 0x75, 0x49, 0xCE, 0x79, 0x3A, 0xA6, 0x21, 0x79, 0x18, 0x68, 0x1D, 0x32, 0x84, 0x26, 0x02, 0x33, 0x66, 0x00, 0xD5, 0x41, 0x06, 0x93, 0xD4, 0xC9, 0x05, 0x36, 0x48, 0x41, 0xE0, 0x5D, 0x0B, 0xE9, 0xEA, 0x3E, 0x76, 0x13, 0xAE, 0xB6, 0xBA, 0x67, 0x55, 0xDB, 0x15, 0x3E, 0xD4, 0x44, 0xF3, 0x9F, 0x67, 0x56, 0xDB, 0x78, 0xED, 0x75, 0x31, 0x4F, 0x1D, 0x11, 0xB8, 0x04, 0x56, 0x73, 0xEE, 0x6D, 0x78, 0x31, 0x75, 0xFD, 0x8B, 0x44, 0x2A, 0x81, 0x65, 0x67, 0x4D, 0x1C, 0xBE, 0x0C, 0x3E, 0x55, 0x21, 0x39, 0x05, 0x66, 0xF0, 0x6A, 0x08, 0x4C, 0xD6, 0x8D, 0x3F, 0xF2, 0xE3, 0xB0, 0xD8, 0x0F, 0x60, 0xE8, 0x77, 0xE5}),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DeserializeRSA(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("RSA(): error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !cmp.Equal(got, tt.want) {
				t.Errorf("RSA():\n got = %v\n want = %v", got, tt.want)
			}
		})
	}
}

func TestDeserializeECDSA(t *testing.T) {
	type args struct {
		input *ecpb.EC
	}
	tests := []struct {
		name    string
		args    args
		want    *ecdsa.PublicKey
		wantErr bool
	}{
		{
			name: "Test a valid public key is correctly deserialized",
			args: args{
				input: &ecpb.EC{
					Crv: ecpb.ECType_P256,
					X:   "m7joBnA3H0VQi1-PWZRqfE3qSzojoDbPJMH0CZP0odo",
					Y:   "rRcW3ovWZOy0WWZI1yKkaFKT3iCMHS2pNhucunTD0ew",
				},
			},
			want: &ecdsa.PublicKey{
				Curve: elliptic.P256(),
				// Big Endian 70435192769055300932927002065149381078422509475567861355254674689345996497370
				X: new(big.Int).SetBytes([]byte{0x9B, 0xB8, 0xE8, 0x06, 0x70, 0x37, 0x1F, 0x45, 0x50, 0x8B, 0x5F, 0x8F, 0x59, 0x94, 0x6A, 0x7C, 0x4D, 0xEA, 0x4B, 0x3A, 0x23, 0xA0, 0x36, 0xCF, 0x24, 0xC1, 0xF4, 0x09, 0x93, 0xF4, 0xA1, 0xDA}),
				// Big Endian 78290918125649382743379572394002796782688715728647748645467977529752316793324
				Y: new(big.Int).SetBytes([]byte{0xAD, 0x17, 0x16, 0xDE, 0x8B, 0xD6, 0x64, 0xEC, 0xB4, 0x59, 0x66, 0x48, 0xD7, 0x22, 0xA4, 0x68, 0x52, 0x93, 0xDE, 0x20, 0x8C, 0x1D, 0x2D, 0xA9, 0x36, 0x1B, 0x9C, 0xBA, 0x74, 0xC3, 0xD1, 0xEC}),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DeserializeEC(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ECDSA(): error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("ECDSA():\n got = %v\n want = %v", got, tt.want)
			}
		})
	}
}

func TestSerializeRSA(t *testing.T) {
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
				key: testdata.RSA.PubKey,
			},
			want: &rsapb.RSA{
				N: testdata.RSA.N,
				E: testdata.RSA.E,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SerializeRSA(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("SerializeRSA() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want, cmpopts.IgnoreUnexported(rsapb.RSA{})) {
				t.Errorf("SerializeRSA() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSerializeECDSA(t *testing.T) {
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
				key: testdata.ECDSA.PubKey,
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
			got, err := SerializeECDSA(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("SerializeECDSA() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want, cmpopts.IgnoreUnexported(ecpb.EC{})) {
				t.Errorf("SerializeECDSA():\n got = %v\n want = %v", got, tt.want)
			}
		})
	}
}

func TestKeySerializationFlowCompatibilityForRSA(t *testing.T) {
	original, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Errorf("Failed to generate a RSA key: error = %v", err)
		return
	}

	serialized, err := SerializeRSA(original.PublicKey)
	if err != nil {
		t.Errorf("Failed to encode RSA key: error = %v", err)
		return
	}

	pubKey, err := DeserializeKey(serialized)
	if err != nil {
		t.Errorf("Failed to serialize RSA key: error = %v", err)
		return
	}

	want := original.PublicKey
	got, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		t.Errorf("Received key is of invalid type: got = %T, want = %T", pubKey, got)
		return
	}

	if !want.Equal(got) {
		t.Errorf("Public Keys are not equal: %v", cmp.Diff(want, got))
	}
}

func TestKeySerializationFlowCompatibilityForECDSA(t *testing.T) {
	original, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Errorf("Failed to generate a ECDSA key: error = %v", err)
		return
	}

	serialized, err := SerializeECDSA(original.PublicKey)
	if err != nil {
		t.Errorf("Failed to encode ECDSA key: error = %v", err)
		return
	}

	pubKey, err := DeserializeKey(serialized)
	if err != nil {
		t.Errorf("Failed to serialize ECDSA key: error = %v", err)
		return
	}

	want := original.PublicKey
	got, ok := pubKey.(*ecdsa.PublicKey)
	if !ok {
		t.Errorf("Received key is of invalid type: got = %T, want = %T", pubKey, got)
		return
	}

	if !want.Equal(got) {
		t.Errorf("Public Keys are not equal: %v", cmp.Diff(want, got))
	}
}
