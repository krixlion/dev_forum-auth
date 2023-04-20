package vault

import (
	vault "github.com/hashicorp/vault/api"
	"github.com/krixlion/dev_forum-auth/pkg/entity"
)

type secretData struct {
	algorithm  Algorithm
	encodedKey string
}

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

func validateSecret(secret *vault.KVSecret) (secretData, error) {
	key, ok := secret.Data["private"]
	if !ok {
		return secretData{}, ErrKeyEmpty
	}

	encodedKey, ok := key.(string)
	if !ok {
		return secretData{}, ErrInvalidKeyFormat
	}

	algo, ok := secret.CustomMetadata["algorithm"]
	if !ok {
		return secretData{}, ErrInvalidAlgorithm
	}

	algorithm, ok := algo.(string)
	if !ok {
		return secretData{}, ErrInvalidAlgorithm
	}

	return secretData{
		algorithm:  Algorithm(algorithm),
		encodedKey: encodedKey,
	}, nil
}
