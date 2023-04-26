package tokens

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/krixlion/dev_forum-auth/pkg/entity"
	"github.com/krixlion/dev_forum-auth/pkg/tokens/testdata"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
)

var clockFunc jwt.Clock = jwt.ClockFunc(time.Now)

func setUpTokenManager() StdTokenManager {
	m := MakeTokenManager(Config{
		Issuer: testdata.TestIssuer,
		Clock:  clockFunc,
	})
	return m
}

func TestTokenManager_Encode(t *testing.T) {
	type args struct {
		privateKey entity.Key
		token      entity.Token
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Test if correctly encodes and signes a token struct",
			args: args{
				privateKey: testdata.TestKey,
				token:      testdata.TestToken,
			},
			want: testdata.SignedJWT,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := setUpTokenManager()
			got, err := m.Encode(tt.args.privateKey, tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("TokenManager.Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if string(got) != tt.want {
				t.Errorf("TokenManager.Encode():\n got = %v\n want = %v\n", string(got), tt.want)
			}
		})
	}
}

func TestTokenManager_GenerateOpaque(t *testing.T) {
	type args struct {
		prefixType OpaqueTokenPrefix
	}
	tests := []struct {
		name        string
		args        args
		want        string
		wantTokenId string
		wantErr     bool
	}{
		{
			name: "Test if raises an error on invalid prefix",
			args: args{
				prefixType: OpaqueTokenPrefix(999999999999),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := setUpTokenManager()

			got, gotTokenId, err := m.GenerateOpaque(tt.args.prefixType)
			if (err != nil) != tt.wantErr {
				t.Errorf("TokenManager.GenerateOpaqueToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("TokenManager.GenerateOpaqueToken():\n got = %v\n want = %v\n", got, tt.want)
				return
			}

			if gotTokenId != tt.wantTokenId {
				t.Errorf("TokenManager.GenerateOpaqueToken():\n gotTokenId = %v\n want = %v\n", gotTokenId, tt.wantTokenId)
			}
		})
	}
}

func TestTokenManager_DecodeOpaque(t *testing.T) {
	type args struct {
		typ                OpaqueTokenPrefix
		encodedOpaqueToken string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Test if correctly decodes a valid token",
			args: args{
				typ:                AccessToken,
				encodedOpaqueToken: "dfa_c2VHWmJVVWhKTWpVYnNlR19mYWJjNTJiYQ==",
			},
			want: "seGZbUUhJMjUbseG",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := setUpTokenManager()
			got, err := m.DecodeOpaque(tt.args.typ, tt.args.encodedOpaqueToken)
			if (err != nil) != tt.wantErr {
				t.Errorf("TokenManager.DecodeOpaqueToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("TokenManager.DecodeOpaqueToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_decodeAndValidateOpaque(t *testing.T) {
	type args struct {
		rawToken string
	}
	tests := []struct {
		algo    jwa.SignatureAlgorithm
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Test if correctly decodes a valid token",
			algo: jwa.HS256,
			args: args{
				rawToken: "eVdGUmxGcFdGVGZDUld5V18yMmRmNDkyMQ==",
			},
			want: "yWFRlFpWFTfCRWyW",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeAndValidateOpaque(tt.args.rawToken)
			if (err != nil) != tt.wantErr {
				t.Errorf("TokenManager.decodeAndValidateOpaque() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("TokenManager.decodeAndValidateOpaque() = %v, want %v", got, tt.want)
			}
		})
	}
}
func Test_verifyAlgorithm(t *testing.T) {
	type args struct {
		algo entity.Algorithm
	}
	tests := []struct {
		name    string
		args    args
		want    jwa.SignatureAlgorithm
		wantErr bool
	}{
		{
			name: "Test if fails on empty algorithm",
			args: args{
				algo: "",
			},
			wantErr: true,
		},
		{
			name: "Test if recognizes RSA",
			args: args{
				algo: entity.RS256,
			},
			want: jwa.RS256,
		},
		{
			name: "Test if recognizes HMAC",
			args: args{
				algo: entity.HS256,
			},
			want: jwa.HS256,
		},
		{
			name: "Test if recognizes ECDSA",
			args: args{
				algo: entity.ES256,
			},
			want: jwa.ES256,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := verifyAlgorithm(tt.args.algo)
			if (err != nil) != tt.wantErr {
				t.Errorf("verifyAlgorithm() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("verifyAlgorithm() = %v, want %v", got, tt.want)
			}
		})
	}
}
