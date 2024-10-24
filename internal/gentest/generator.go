package gentest

import (
	"encoding/json"
	"math/rand/v2"
	"time"

	"github.com/gofrs/uuid"
	"github.com/krixlion/dev_forum-auth/pkg/entity"
	"github.com/krixlion/dev_forum-auth/pkg/tokens"
	"github.com/krixlion/dev_forum-auth/pkg/tokens/manager"
)

func RandomString(length int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	v := make([]rune, length)
	for i := range v {
		v[i] = letters[rand.IntN(len(letters))]
	}
	return string(v)
}

// RandomToken panics on hardware error.
// It should be used ONLY for testing.
func RandomToken(tokenType entity.TokenType) entity.Token {
	userId := uuid.Must(uuid.NewV4())

	prefix := tokens.RefreshToken
	if tokenType == entity.AccessToken {
		prefix = tokens.AccessToken
	}

	_, id, err := manager.MakeManager(manager.Config{Issuer: "gentest"}).GenerateOpaque(prefix)
	if err != nil {
		panic(err)
	}

	return entity.Token{
		Id:        id,
		UserId:    userId.String(),
		Type:      tokenType,
		ExpiresAt: time.Now().Add(time.Minute),
		IssuedAt:  time.Now(),
	}
}

// Randomauth returns a random auth marshaled
// to JSON and panics on error.
// It should be used ONLY for testing.
func RandomJSONAuth(tokenType entity.TokenType) []byte {
	auth := RandomToken(tokenType)
	json, err := json.Marshal(auth)
	if err != nil {
		panic(err)
	}
	return json
}
