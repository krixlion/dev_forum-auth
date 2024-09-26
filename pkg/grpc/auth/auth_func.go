package auth

import (
	"context"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"

	"github.com/krixlion/dev_forum-auth/pkg/tokens"
	"github.com/krixlion/dev_forum-lib/tracing"
	"go.opentelemetry.io/otel/trace"
)

// NewAuthFunc returns a callback to be used with grpc_auth interceptor.
// It reads the Bearer token from the context of an incoming request
// and verifies it using given tokens.Validator.
// If the validator fails to verify the token an error is returned.
// Otherwise the context is returned unaltered.
func NewAuthFunc(tokenValidator tokens.Validator, tracer trace.Tracer) grpc_auth.AuthFunc {
	return func(ctx context.Context) (_ context.Context, err error) {
		ctx, span := tracer.Start(ctx, "server.AuthFunc")
		defer span.End()
		defer tracing.SetSpanErr(span, err)

		token, err := grpc_auth.AuthFromMD(ctx, "Bearer")
		if err != nil {
			return nil, err
		}

		if err := tokenValidator.ValidateToken(token); err != nil {
			return nil, err
		}

		return ctx, nil
	}
}
