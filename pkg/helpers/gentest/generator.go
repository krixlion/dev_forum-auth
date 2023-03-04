package gentest

import (
	"encoding/json"
	"math/rand"

	"github.com/gofrs/uuid"
	"github.com/krixlion/dev_forum-auth/pkg/entity"
)

func RandomString(length int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	v := make([]rune, length)
	for i := range v {
		v[i] = letters[rand.Intn(len(letters))]
	}
	return string(v)
}

// Randomauth panics on hardware error.
// It should be used ONLY for testing.
func Randomauth(titleLen, bodyLen int) entity.Token {
	id := uuid.Must(uuid.NewV4())
	userId := uuid.Must(uuid.NewV4())

	return entity.Token{
		Id:     id.String(),
		UserId: userId.String(),
		Title:  RandomString(titleLen),
		Body:   RandomString(bodyLen),
	}
}

// Randomauth returns a random auth marshaled
// to JSON and panics on error.
// It should be used ONLY for testing.
func RandomJSONauth(titleLen, bodyLen int) []byte {
	auth := Randomauth(titleLen, bodyLen)
	json, err := json.Marshal(auth)
	if err != nil {
		panic(err)
	}
	return json
}
