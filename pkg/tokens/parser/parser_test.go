package parser

import (
	"context"
	"testing"
	"time"

	"github.com/krixlion/dev_forum-auth/internal/gentest"
	"github.com/krixlion/dev_forum-auth/pkg/entity"
	"github.com/krixlion/dev_forum-auth/pkg/entity/testdata"
	"github.com/lestrrat-go/jwx/jwt"
)

var testClockFunc = jwt.ClockFunc(func() time.Time {
	return time.Unix(1682517486, 0)
})

var testKey = Key{
	Id:        "test",
	Algorithm: "HS256",
	Raw:       testdata.HMACKey,
}

func setUpTokenParser(ctx context.Context, refreshFunc RefreshFunc, clockFunc jwt.Clock) *JWTParser {
	v, err := NewParser(testdata.Issuer, refreshFunc, WithClock(clockFunc))
	if err != nil {
		panic(err)
	}

	go func() {
		v.Run(ctx)
	}()

	// Wait for the goroutine to start up.
	time.Sleep(time.Millisecond * 10)

	return v
}

func TestJWTParser_ParseToken(t *testing.T) {
	type args struct {
		token string
	}
	tests := []struct {
		name        string
		args        args
		refreshFunc RefreshFunc
		clockFunc   jwt.Clock
		want        entity.Token
		wantErr     bool
	}{
		{
			name: "Test if correctly parses a valid token",
			args: args{
				token: testdata.AccessJWToken,
			},
			refreshFunc: func(ctx context.Context) ([]Key, error) {
				return []Key{testKey}, nil
			},
			clockFunc: testClockFunc,
			want:      testdata.AccessToken,
			wantErr:   false,
		},
		{
			name: "Test if fails on invalid token type",
			args: args{
				token: testdata.RefreshJWToken,
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
				token: testdata.AccessJWToken,
			},
			refreshFunc: func(ctx context.Context) ([]Key, error) {
				return []Key{{
					Id:        "test",
					Type:      "HMAC",
					Algorithm: "RS256",
					Raw:       testdata.HMACKey,
				}}, nil
			},
			clockFunc: testClockFunc,
			wantErr:   true,
		},
		{
			name: "Test if fails on expired token",
			args: args{
				token: testdata.AccessJWToken,
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
			parser := setUpTokenParser(ctx, tt.refreshFunc, tt.clockFunc)
			got, err := parser.ParseToken(tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Fatalf("JWTParser.ParseToken() error = %v, wantErr = %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if got != tt.want {
				t.Fatalf("JWTParser.ParseToken():\n got = %v\n want = %v", got, tt.want)
			}
		})
	}
}

func Test_NewTokenParser(t *testing.T) {
	type args struct {
		Issuer      string
		RefreshFunc RefreshFunc
		options     []Option
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test if returns an error on nil RefreshFunc",
			args: args{
				Issuer:      testdata.Issuer,
				options:     []Option{WithClock(testClockFunc)},
				RefreshFunc: nil,
			},
			wantErr: true,
		},
		{
			name: "Test if does not return an err on nil Clock",
			args: args{
				Issuer:      testdata.Issuer,
				options:     []Option{WithClock(nil)},
				RefreshFunc: func(ctx context.Context) ([]Key, error) { return nil, nil },
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewParser(tt.args.Issuer, tt.args.RefreshFunc, tt.args.options...); (err != nil) != tt.wantErr {
				t.Fatalf("MakeTokenParser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestJWTParser_RunReturnsOnContextCancellation(t *testing.T) {
	parser, err := NewParser("", func(ctx context.Context) ([]Key, error) { return []Key{testKey}, nil })
	if err != nil {
		t.Fatalf("JWTParser.Run() unexpected error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})

	go func() {
		parser.Run(ctx)
		done <- struct{}{}
	}()
	cancel()

	ctxT, cancelT := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancelT()

	select {
	case <-ctxT.Done():
		t.Fatalf("JWTParser.Run() did not return on context cancellation")
	case <-done:
		return
	}
}
