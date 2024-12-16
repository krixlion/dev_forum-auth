package testdata

import (
	"time"

	"github.com/krixlion/dev_forum-auth/pkg/entity"
)

const Issuer = "test"

// Signed valid JWT tokens.
const (
	AccessJWToken  = "eyJhbGciOiJIUzI1NiIsImtpZCI6InRlc3QiLCJ0eXAiOiJKV1QifQ.eyJleHAiOjE2ODI1MTc1NzIsImlhdCI6MTY4MjUxNzI4NiwiaXNzIjoidGVzdCIsImp0aSI6InRlc3QiLCJzdWIiOiJ0ZXN0LWlkIiwidHlwZSI6ImFjY2Vzcy10b2tlbiJ9.wxoMBhYMLxZo_0il-EeQOnfcYUXfyuGWI--3IiYupbY"
	RefreshJWToken = "eyJhbGciOiJIUzI1NiIsImtpZCI6InRlc3QiLCJ0eXAiOiJKV1QifQ.eyJleHAiOjE2ODI1MTc1NzIsImlhdCI6MTY4MjUxNzI4NiwiaXNzIjoidGVzdCIsImp0aSI6InRlc3QiLCJzdWIiOiJ0ZXN0LWlkIiwidHlwZSI6InJlZnJlc2gtdG9rZW4ifQ.uiDFSRVO5urzRb5u4aXD4fn15hmNZN9w8ArDDdbLC5Q"
)

var (
	HMACKey = []byte("key")

	TestKey = entity.Key{
		Id:        "test",
		Algorithm: entity.HS256,
		Raw:       HMACKey,
	}

	AccessToken = entity.Token{
		Id:        "test",
		UserId:    "test-id",
		Type:      entity.AccessToken,
		ExpiresAt: time.Unix(1682517572, 0).In(time.UTC),
		IssuedAt:  time.Unix(1682517286, 0).In(time.UTC),
	}

	TestRefreshToken = entity.Token{
		Id:        "test",
		UserId:    "test-id",
		Type:      entity.RefreshToken,
		ExpiresAt: time.Unix(1682517572, 0).In(time.UTC),
		IssuedAt:  time.Unix(1682517286, 0).In(time.UTC),
	}
)
