package tokens

import (
	"reflect"
	"testing"
	"time"

	"github.com/krixlion/dev_forum-auth/pkg/entity"
	"github.com/lestrrat-go/jwx/jwa"
)

func setUpTokenManager(algo jwa.SignatureAlgorithm) TokenManager {
	m := MakeTokenManager("test", Config{
		SignatureAlgorithm: algo,
	})
	return m
}

func TestTokenManager_Parse(t *testing.T) {
	type args struct {
		publicKey interface{}
		token     string
	}
	tests := []struct {
		name    string
		algo    jwa.SignatureAlgorithm
		args    args
		want    entity.Token
		wantErr bool
	}{
		// TODO add cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := setUpTokenManager(tt.algo)
			got, err := m.Parse(tt.args.publicKey, []byte(tt.args.token))
			if (err != nil) != tt.wantErr {
				t.Errorf("TokenManager.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TokenManager.Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTokenManager_Encode(t *testing.T) {
	type args struct {
		privateKey interface{}
		token      entity.Token
	}
	tests := []struct {
		algo    jwa.SignatureAlgorithm
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Test if correctly encodes and signes a token struct with HS256 algo",
			algo: jwa.HS256,
			args: args{
				privateKey: []byte("key"),
				token: entity.Token{
					Id:        "test",
					UserId:    "test-id",
					Type:      entity.AccessToken,
					ExpiresAt: time.Unix(1680358945, 0),
					IssuedAt:  time.Unix(1680358945, 0),
				},
			},
			want: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2ODAzNTg5NDUsImlhdCI6MTY4MDM1ODk0NSwiaXNzIjoidGVzdCIsImp0aSI6InRlc3QiLCJzdWIiOiJ0ZXN0LWlkIn0.POi3q7JKM71nG49W-UMDo81pAO3AQ0O7KtbOxqZynD4",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := setUpTokenManager(tt.algo)
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

func TestTokenManager_GenerateOpaqueToken(t *testing.T) {
	type args struct {
		prefixType OpaqueTokenPrefix
	}
	tests := []struct {
		algo        jwa.SignatureAlgorithm
		name        string
		args        args
		want        string
		wantTokenId string
		wantErr     bool
	}{
		{
			name: "Test if raises an error on invalid prefix",
			algo: jwa.HS256,
			args: args{
				prefixType: OpaqueTokenPrefix(999999999999),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := setUpTokenManager(tt.algo)

			got, gotTokenId, err := m.GenerateOpaqueToken(tt.args.prefixType)
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

func TestTokenManager_DecodeOpaqueToken(t *testing.T) {
	type args struct {
		typ                OpaqueTokenPrefix
		encodedOpaqueToken string
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
			args: args{
				typ:                AccessToken,
				encodedOpaqueToken: "dfa_c2VHWmJVVWhKTWpVYnNlR19mYWJjNTJiYQ==",
			},
			want: "seGZbUUhJMjUbseG",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := setUpTokenManager(tt.algo)
			got, err := m.DecodeOpaqueToken(tt.args.typ, tt.args.encodedOpaqueToken)
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

func TestTokenManager_decodeAndValidateOpaque(t *testing.T) {
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
			m := setUpTokenManager(tt.algo)
			got, err := m.decodeAndValidateOpaque(tt.args.rawToken)
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

func Test_randomAlphaString(t *testing.T) {
	type args struct {
		length int
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Test if returned string has correct length",
			args: args{
				length: 5,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := randomAlphaString(tt.args.length)
			if (err != nil) != tt.wantErr {
				t.Errorf("randomAlphaString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != tt.args.length {
				t.Errorf("randomAlphaString() invalid length: got = %v expected length %v", got, tt.args.length)
			}
		})
	}
}
