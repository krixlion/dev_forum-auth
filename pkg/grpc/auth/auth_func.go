package auth

import (
	"context"
	"errors"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"

	"github.com/krixlion/dev_forum-auth/pkg/entity"
	"github.com/krixlion/dev_forum-auth/pkg/tokens"
	"github.com/krixlion/dev_forum-lib/tracing"
	"go.opentelemetry.io/otel/trace"
)

type ctxTokenKey struct{}

// NewAuthFunc returns a callback to be used with grpc_auth interceptor.
// It reads the Bearer token from the context of an incoming request
// and verifies it using given tokens.Validator.
// If the validator fails to verify the token an error is returned.
// Otherwise the context is returned unaltered.
func NewAuthFunc(tokenParser tokens.Parser, tracer trace.Tracer) grpc_auth.AuthFunc {
	return func(ctx context.Context) (_ context.Context, err error) {
		ctx, span := tracer.Start(ctx, "server.AuthFunc")
		defer span.End()
		defer tracing.SetSpanErr(span, err)

		token, err := grpc_auth.AuthFromMD(ctx, "Bearer")
		if err != nil {
			return nil, err
		}

		t, err := tokenParser.ParseToken(token)
		if err != nil {
			return nil, err
		}

		return context.WithValue(ctx, ctxTokenKey{}, t), nil
	}
}

func GetTokenFromCtx(ctx context.Context) (entity.Token, error) {
	t, ok := ctx.Value(ctxTokenKey{}).(entity.Token)
	if !ok || (t == entity.Token{}) {
		return entity.Token{}, errors.New("token is missing or is of an unexpected type")
	}

	return t, nil
}
