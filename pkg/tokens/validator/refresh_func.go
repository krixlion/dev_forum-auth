package validator

import (
	"context"
	"io"

	"github.com/krixlion/dev_forum-auth/pkg/grpc/serialize"
	pb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1"
	"github.com/lestrrat-go/jwx/jwk"
	"google.golang.org/protobuf/types/known/emptypb"
)

type RefreshFunc func(ctx context.Context) ([]Key, error)

// Key is a struct for data necessary to register a key in a keyset.
type Key struct {
	Id        string
	Algorithm string
	Type      string
	Raw       interface{}
}

// DefaultRefreshFunc returns a callback that uses the auth service as the
// keyset source and fetches the keyset using provided gRPC client.
func DefaultRefreshFunc(authClient pb.AuthServiceClient) RefreshFunc {
	return func(ctx context.Context) ([]Key, error) {
		stream, err := authClient.GetValidationKeySet(ctx, &emptypb.Empty{})
		if err != nil {
			return nil, err
		}

		keyset := []Key{}

		for {
			if err := ctx.Err(); err != nil {
				return nil, err
			}

			jwk, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					break
				}
				return nil, err
			}

			rawMessage, err := jwk.Key.UnmarshalNew()
			if err != nil {
				return nil, err
			}

			raw, err := serialize.Key(rawMessage)
			if err != nil {
				return nil, err
			}

			keyset = append(keyset, Key{
				Id:        jwk.GetKid(),
				Algorithm: jwk.GetAlg(),
				Type:      jwk.GetKty(),
				Raw:       raw,
			})
		}

		return keyset, nil
	}
}

// keySetFromKeys copies provided keys to a new keyset and returns it.
func keySetFromKeys(keys []Key) (jwk.Set, error) {
	keySet := jwk.NewSet()

	if keys == nil {
		return nil, ErrKeysNotReceived
	}

	for _, key := range keys {
		jwKey, err := jwk.New(key.Raw)
		if err != nil {
			return nil, err
		}

		if err := jwKey.Set(jwk.KeyIDKey, key.Id); err != nil {
			return nil, err
		}

		if err := jwKey.Set(jwk.KeyTypeKey, key.Type); err != nil {
			return nil, err
		}

		if err := jwKey.Set(jwk.AlgorithmKey, key.Algorithm); err != nil {
			return nil, err
		}

		keySet.Add(jwKey)
	}

	return keySet, nil
}
