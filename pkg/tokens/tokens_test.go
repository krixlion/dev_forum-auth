package tokens

import (
	"reflect"
	"testing"

	"github.com/golang-jwt/jwt"
	"github.com/krixlion/dev_forum-auth/pkg/entity"
)

func setUpTokenManager() TokenManager {
	m := MakeTokenManager("testing", nil, nil)
	return m
}

func TestTokenManager_Parse(t *testing.T) {
	type args struct {
		token string
	}
	tests := []struct {
		name    string
		args    args
		want    entity.Token
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := setUpTokenManager()
			got, err := m.Parse(tt.args.token)
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
		userPassword string
		token        entity.Token
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := setUpTokenManager()
			got, err := m.Encode(tt.args.userPassword, tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("TokenManager.Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("TokenManager.Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTokenManager_GenerateOpaqueToken(t *testing.T) {
	type args struct {
		prefixType OpaqueTokenPrefix
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := setUpTokenManager()
			got, got1, err := m.GenerateOpaqueToken(tt.args.prefixType)
			if (err != nil) != tt.wantErr {
				t.Errorf("TokenManager.GenerateOpaqueToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("TokenManager.GenerateOpaqueToken() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("TokenManager.GenerateOpaqueToken() got1 = %v, want %v", got1, tt.want1)
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
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := setUpTokenManager()
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

func TestTokenManager_isValid(t *testing.T) {
	type args struct {
		t *jwt.Token
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := setUpTokenManager()
			got, err := m.isValid(tt.args.t)
			if (err != nil) != tt.wantErr {
				t.Errorf("TokenManager.isValid() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TokenManager.isValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTokenManager_EncryptAsJWE(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := setUpTokenManager()
			got, err := m.EncryptAsJWE(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("TokenManager.EncryptAsJWE() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("TokenManager.EncryptAsJWE() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTokenManager_DecryptJWE(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := setUpTokenManager()
			got, err := m.DecryptJWE(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("TokenManager.DecryptJWE() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TokenManager.DecryptJWE() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTokenManager_decodeAndValidateOpaque(t *testing.T) {
	type args struct {
		rawToken string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := setUpTokenManager()
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := randomAlphaString(tt.args.length)
			if (err != nil) != tt.wantErr {
				t.Errorf("randomAlphaString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("randomAlphaString() = %v, want %v", got, tt.want)
			}
		})
	}
}
