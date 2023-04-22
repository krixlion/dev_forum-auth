package vault

import (
	vault "github.com/hashicorp/vault/api"
	"github.com/krixlion/dev_forum-auth/pkg/entity"
)

// secretData is a convenience struct used to contain parsed KVV2 keys.
type secretData struct {
	algorithm  Algorithm
	encodedKey string
}

func parseSecret(secret *vault.KVSecret) (secretData, error) {
	if secret == nil {
		return secretData{}, ErrKeyMissing
	}

	algo, ok := secret.Data["algorithm"]
	if !ok {
		return secretData{}, ErrInvalidAlgorithm
	}

	algorithm, ok := algo.(string)
	if !ok {
		return secretData{}, ErrInvalidAlgorithm
	}

	key, ok := secret.Data["private"]
	if !ok {
		return secretData{}, ErrKeyMissing
	}

	encodedKey, ok := key.(string)
	if !ok {
		return secretData{}, ErrInvalidKeyFormat
	}

	return secretData{
		algorithm:  Algorithm(algorithm),
		encodedKey: encodedKey,
	}, nil
}

// makeKey is a convenience func used to make an entity.Key
// correctly decoded with correct encodeFunc assigned.
func makeKey(id string, validated secretData) (entity.Key, error) {
	privateKey, encodeFunc, err := Decode(validated.algorithm, validated.encodedKey)
	if err != nil {
		return entity.Key{}, err
	}

	return entity.Key{
		Id:         id,
		Type:       string(validated.algorithm),
		Raw:        privateKey,
		EncodeFunc: encodeFunc,
	}, nil
}
