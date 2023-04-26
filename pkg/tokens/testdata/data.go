package testdata

import (
	"time"

	"github.com/krixlion/dev_forum-auth/pkg/entity"
	"github.com/lestrrat-go/jwx/jwt"
)

const (
	SignedJWT  = "eyJhbGciOiJIUzI1NiIsImtpZCI6InRlc3QiLCJ0eXAiOiJKV1QifQ.eyJleHAiOjE2ODI1MTc1NzIsImlhdCI6MTY4MjUxNzI4NiwiaXNzIjoidGVzdCIsImp0aSI6InRlc3QiLCJzdWIiOiJ0ZXN0LWlkIiwidHlwZSI6ImFjY2Vzcy10b2tlbiJ9.wxoMBhYMLxZo_0il-EeQOnfcYUXfyuGWI--3IiYupbY"
	TestIssuer = "test"
)

var (
	TestHMACKey = []byte("key")

	TestClockFunc = jwt.ClockFunc(func() time.Time {
		return time.Unix(1682517486, 0)
	})

	TestToken = entity.Token{
		Id:        "test",
		UserId:    "test-id",
		Type:      entity.AccessToken,
		ExpiresAt: time.Unix(1682517572, 0),
		IssuedAt:  time.Unix(1682517286, 0),
	}

	TestKey = entity.Key{
		Id:        "test",
		Raw:       TestHMACKey,
		Algorithm: entity.HS256,
	}
)
