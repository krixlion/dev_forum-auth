package gentest

import (
	"math/rand"
)

func RandomString(length int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	v := make([]rune, length)
	for i := range v {
		v[i] = letters[rand.Intn(len(letters))]
	}
	return string(v)
}

// // RandomAuth panics on hardware error.
// // It should be used ONLY for testing.
// func RandomAuth(titleLen, bodyLen int) entity.Token {
// 	id := uuid.Must(uuid.NewV4())
// 	userId := uuid.Must(uuid.NewV4())

// 	return entity.Token{
// 		Id:     id.String(),
// 		UserId: userId.String(),
// 		Title:  RandomString(titleLen),
// 		Body:   RandomString(bodyLen),
// 	}
// }

// // Randomauth returns a random auth marshaled
// // to JSON and panics on error.
// // It should be used ONLY for testing.
// func RandomJSONauth(titleLen, bodyLen int) []byte {
// 	auth := Randomauth(titleLen, bodyLen)
// 	json, err := json.Marshal(auth)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return json
// }
