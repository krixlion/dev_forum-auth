package validator

import (
	"context"
	"io"

	"github.com/krixlion/dev_forum-auth/pkg/grpc/deserialize"
	pb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1"
	"github.com/krixlion/dev_forum-lib/nulls"
	"github.com/krixlion/dev_forum-lib/tracing"
	"github.com/lestrrat-go/jwx/jwk"
	"go.opentelemetry.io/otel/trace"
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
// Tracing is disabled if no tracer is provided.
func DefaultRefreshFunc(authClient pb.AuthServiceClient, tracer trace.Tracer) RefreshFunc {
	if tracer == nil {
		tracer = nulls.NullTracer{}
	}

	return func(ctx context.Context) ([]Key, error) {
		ctx, span := tracer.Start(ctx, "refreshFunc")
		defer span.End()

		stream, err := authClient.GetValidationKeySet(ctx, &emptypb.Empty{})
		if err != nil {
			tracing.SetSpanErr(span, err)
			return nil, err
		}

		keyset := []Key{}

		for {
			if err := ctx.Err(); err != nil {
				tracing.SetSpanErr(span, err)
				return nil, err
			}

			jwk, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					break
				}
				tracing.SetSpanErr(span, err)
				return nil, err
			}

			rawMessage, err := jwk.Key.UnmarshalNew()
			if err != nil {
				tracing.SetSpanErr(span, err)
				return nil, err
			}

			raw, err := deserialize.Key(rawMessage)
			if err != nil {
				tracing.SetSpanErr(span, err)
				return nil, err
			}

			key := Key{
				Id:        jwk.GetKid(),
				Algorithm: jwk.GetAlg(),
				Type:      jwk.GetKty(),
				Raw:       raw,
			}

			keyset = append(keyset, key)
		}

		return keyset, nil
	}
}

// keySetFromKeys copies provided keys to a new keyset and returns it.
func keySetFromKeys(keys []Key) (jwk.Set, error) {
	if keys == nil {
		return nil, ErrKeysNotReceived
	}

	keySet := jwk.NewSet()

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
