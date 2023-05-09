package validator

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/krixlion/dev_forum-auth/internal/gentest"
	"github.com/lestrrat-go/jwx/jwt"
)

const testIssuer = "test"

// Signed valid JWT token.
const (
	testAccessToken  = "eyJhbGciOiJIUzI1NiIsImtpZCI6InRlc3QiLCJ0eXAiOiJKV1QifQ.eyJleHAiOjE2ODI1MTc1NzIsImlhdCI6MTY4MjUxNzI4NiwiaXNzIjoidGVzdCIsImp0aSI6InRlc3QiLCJzdWIiOiJ0ZXN0LWlkIiwidHlwZSI6ImFjY2Vzcy10b2tlbiJ9.wxoMBhYMLxZo_0il-EeQOnfcYUXfyuGWI--3IiYupbY"
	testRefreshToken = "eyJhbGciOiJIUzI1NiIsImtpZCI6InRlc3QiLCJ0eXAiOiJKV1QifQ.eyJleHAiOjE2ODI1MTc1NzIsImlhdCI6MTY4MjUxNzI4NiwiaXNzIjoidGVzdCIsImp0aSI6InRlc3QiLCJzdWIiOiJ0ZXN0LWlkIiwidHlwZSI6InJlZnJlc2gtdG9rZW4ifQ.uiDFSRVO5urzRb5u4aXD4fn15hmNZN9w8ArDDdbLC5Q"
)

var (
	testHMACKey = []byte("key")

	testClockFunc = jwt.ClockFunc(func() time.Time {
		return time.Unix(1682517486, 0)
	})

	testKey = Key{
		Id:        "test",
		Algorithm: "HS256",
		Raw:       testHMACKey,
	}
)

func setUpTokenValidator(ctx context.Context, refresher RefreshFunc, clockFunc jwt.Clock) *JWTValidator {
	v, err := NewValidator(Config{
		Issuer:      testIssuer,
		Clock:       clockFunc,
		RefreshFunc: refresher,
	})
	if err != nil {
		panic(err)
	}

	go func() {
		if err := v.Run(ctx); err != nil && !(errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)) {
			panic(err)
		}
	}()

	// Wait for the goroutine to start up.
	time.Sleep(time.Millisecond)

	return v
}

func TestTokenValidator_VerifyJWT(t *testing.T) {
	type args struct {
		token string
	}
	tests := []struct {
		name        string
		args        args
		refreshFunc RefreshFunc
		clockFunc   jwt.Clock
		wantErr     bool
	}{
		{
			name: "Test if correctly parses a valid token",
			args: args{
				token: testAccessToken,
			},
			refreshFunc: func(ctx context.Context) ([]Key, error) {
				return []Key{testKey}, nil
			},
			clockFunc: testClockFunc,
			wantErr:   false,
		},
		{
			name: "Test if fails on invalid token type",
			args: args{
				token: testRefreshToken,
			},
			refreshFunc: func(ctx context.Context) ([]Key, error) {
				return []Key{testKey}, nil
			},
			clockFunc: testClockFunc,
			wantErr:   true,
		},
		{
			name: "Test if fails on invalid algorithm",
			args: args{
				token: testRefreshToken,
			},
			refreshFunc: func(ctx context.Context) ([]Key, error) {
				return []Key{{
					Id:        "test",
					Type:      "HMAC",
					Algorithm: "HS256",
					Raw:       testHMACKey,
				}}, nil
			},
			clockFunc: testClockFunc,
			wantErr:   true,
		},
		{
			name: "Test if fails on expired token",
			args: args{
				token: testRefreshToken,
			},
			refreshFunc: func(ctx context.Context) ([]Key, error) {
				return []Key{testKey}, nil
			},
			clockFunc: jwt.ClockFunc(func() time.Time {
				return time.Now().Add(time.Hour * 24)
			}),
			wantErr: true,
		},
		{
			name: "Test if fails on malformed token",
			args: args{
				token: gentest.RandomString(50),
			},
			refreshFunc: func(ctx context.Context) ([]Key, error) {
				return []Key{testKey}, nil
			},
			clockFunc: jwt.ClockFunc(func() time.Time {
				return time.Now().Add(time.Hour * 24)
			}),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			v := setUpTokenValidator(ctx, tt.refreshFunc, tt.clockFunc)

			if err := v.VerifyToken(tt.args.token); (err != nil) != tt.wantErr {
				t.Errorf("TokenManager.VerifyJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_NewTokenValidator(t *testing.T) {
	type args struct {
		config Config
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test if returns an error on nil RefreshFunc",
			args: args{config: Config{
				Issuer:      testIssuer,
				Clock:       testClockFunc,
				RefreshFunc: nil,
			}},
			wantErr: true,
		},
		{
			name: "Test if does not return an err on nil Clock",
			args: args{config: Config{
				Issuer:      testIssuer,
				Clock:       nil,
				RefreshFunc: func(ctx context.Context) ([]Key, error) { return nil, nil },
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewValidator(tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("MakeTokenValidator() error = %v, wantErr %v", err, tt.wantErr)
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
