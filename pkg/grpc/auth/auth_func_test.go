package auth

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/metadata"
	"github.com/krixlion/dev_forum-auth/pkg/tokens"
	"github.com/krixlion/dev_forum-auth/pkg/tokens/tokensmocks"
	"github.com/krixlion/dev_forum-lib/nulls"
	"go.opentelemetry.io/otel/trace"
)

func TestNewAuthFunc(t *testing.T) {
	type args struct {
		tokenValidator tokens.Validator
		tracer         trace.Tracer
		ctx            context.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test no error is returned on valid token",
			args: args{
				tokenValidator: func() tokensmocks.TokenValidator {
					m := tokensmocks.NewTokenValidator()
					m.On("ValidateToken", "test-token").Return(nil).Once()
					return m
				}(),
				tracer: nulls.NullTracer{},
				ctx:    metadata.MD{}.Add("authorization", "Bearer test-token").ToIncoming(context.Background()),
			},
			wantErr: false,
		},
		{
			name: "Test error is returned if the context does not contain a Bearer token",
			args: args{
				tokenValidator: tokensmocks.NewTokenValidator(),
				tracer:         nulls.NullTracer{},
				ctx:            context.Background(),
			},
			wantErr: true,
		},
		{
			name: "Test error is returned if the validator fails to validate the token",
			args: args{
				tokenValidator: func() tokensmocks.TokenValidator {
					m := tokensmocks.NewTokenValidator()
					m.On("ValidateToken", "test-token").Return(errors.New("test-err")).Once()
					return m
				}(),
				tracer: nulls.NullTracer{},
				ctx:    metadata.MD{}.Add("authorization", "Bearer test-token").ToIncoming(context.Background()),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAuthFunc(tt.args.tokenValidator, tt.args.tracer)(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAuthFunc():\n error = %v\n wantErr = %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if !reflect.DeepEqual(got, tt.args.ctx) {
				t.Errorf("NewAuthFunc():\n got = %v\n want = %v", got, tt.args.ctx)
			}
		})
	}
}
