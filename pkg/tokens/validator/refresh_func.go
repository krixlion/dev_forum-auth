package validator

import (
	"context"
	"io"

	"github.com/krixlion/dev_forum-auth/pkg/grpc/serialize"
	pb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

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
