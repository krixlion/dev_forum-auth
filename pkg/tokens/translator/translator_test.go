package translator

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/krixlion/dev_forum-auth/pkg/grpc/mocks"
	pb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1"
	"github.com/stretchr/testify/mock"
)

func Test_NewTranslator(t *testing.T) {
	t.Run("Test all channels are initialized with valid buffer size", func(t *testing.T) {
		queueSize := 5
		translator := NewTranslator(nil, Config{JobQueueSize: queueSize})

		if got := translator.jobs; got == nil {
			t.Errorf("NewTranslator(): a chan was not initialized\n got = %v\n", got)
		}

		if got := cap(translator.jobs); got != queueSize {
			t.Errorf("NewTranslator():\n got = %v\n want = %v\n", got, queueSize)
		}

		if got := translator.streamAborted; got == nil {
			t.Errorf("NewTranslator(): a chan was not initialized\n got = %v\n", got)
		}
	})
	t.Run("Test given options can mutate the final struct", func(t *testing.T) {
		translator := NewTranslator(mocks.NewAuthClient(), Config{}, optionFunc(func(t *Translator) {
			t.grpcClient = nil
		}))

		if got := translator.grpcClient; got != nil {
			t.Errorf("NewTranslator():\n got = %v\n want = %v\n", got, nil)
		}
	})
}

func TestTranslator_Run(t *testing.T) {
	t.Run("Test Run returns on context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		clientMock := mocks.NewAuthClient()
		clientMock.On("TranslateAccessToken", ctx).Return(nil, nil).Once()
		tr := NewTranslator(clientMock, Config{})

		finished := make(chan bool)
		go func() {
			tr.Run(ctx)
			finished <- true
		}()

		before := time.Now()

		// Shutdown the translator.
		cancel()

		select {
		case <-time.After(time.Millisecond):
			t.Errorf("Run did not stop on context cancellation. Time needed for func to return: %vs", time.Since(before).Seconds())
		case <-finished:
			return
		}
	})
}

func TestTranslator_handleJobs(t *testing.T) {
	type fields struct {
		grpcClient pb.AuthServiceClient
		config     Config
	}
	type args struct {
		job job
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   result
	}{
		{
			name: "Test jobs are being executed",
			fields: fields{
				grpcClient: func() mocks.AuthClient {
					s := mocks.NewAuthStreamClient()
					s.On("Send", &pb.TranslateAccessTokenRequest{OpaqueAccessToken: "test-opaque"}).Return(nil).Once()
					s.On("Recv").Return(&pb.TranslateAccessTokenResponse{AccessToken: "test-token"}, nil).Once()

					m := mocks.NewAuthClient()
					m.On("TranslateAccessToken", mock.Anything, mock.Anything).Return(s, nil).Once()
					return m
				}(),
			},
			args: args{job: job{
				OpaqueAccessToken: "test-opaque",
				ResultC:           make(chan result),
			}},
			want: result{
				TranslatedAccessToken: "test-token",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()

			tr := NewTranslator(tt.fields.grpcClient, tt.fields.config)
			go tr.Run(ctx)

			tr.jobs <- tt.args.job
			got := <-tt.args.job.ResultC

			if (got.Err != nil) != (tt.want.Err != nil) {
				t.Errorf("Translator.handleJobs():\n error = %v\n wantErr = %v\n", got.Err, tt.want.Err)
				return
			}

			if got.TranslatedAccessToken != tt.want.TranslatedAccessToken {
				t.Errorf("Translator.handleJobs():\n got = %v\n want = %v", got.TranslatedAccessToken, tt.want.TranslatedAccessToken)
			}
		})
	}
}

func Test_isStreamRenewable(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Test returns false when error is nil",
			args: args{
				err: nil,
			},
			want: false,
		},
		{
			name: "Test returns false when error is io.EOF",
			args: args{
				err: io.EOF,
			},
			want: false,
		},
		{
			name: "Test returns true when error is wrapped io.EOF",
			args: args{
				err: fmt.Errorf("%w", io.EOF),
			},
			want: true,
		},
		{
			name: "Test returns true on valid error",
			args: args{
				err: errors.New("test err"),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isStreamRenewable(tt.args.err); got != tt.want {
				t.Errorf("isStreamRenewable():\n got = %v\n want = %v", got, tt.want)
			}
		})
	}
}

func TestTranslator_TranslateAccessToken(t *testing.T) {
	type fields struct {
		grpcClient pb.AuthServiceClient
		config     Config
	}
	type args struct {
		opaqueAccessToken string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Test token is returned with no error on happy path",
			fields: fields{
				grpcClient: func() mocks.AuthClient {
					s := mocks.NewAuthStreamClient()
					s.On("Send", &pb.TranslateAccessTokenRequest{OpaqueAccessToken: "test-opaque", Metadata: map[string]string{}}).Return(nil).Once()
					s.On("Recv").Return(&pb.TranslateAccessTokenResponse{AccessToken: "test-translated-token", Metadata: map[string]string{}}, nil).Once()

					m := mocks.NewAuthClient()
					m.On("TranslateAccessToken", mock.Anything, mock.Anything).Return(s, nil).Once()
					return m
				}(),
			},
			args:    args{opaqueAccessToken: "test-opaque"},
			want:    "test-translated-token",
			wantErr: false,
		},
		{
			name: "Test error is returned on Recv error",
			fields: fields{
				grpcClient: func() mocks.AuthClient {
					s := mocks.NewAuthStreamClient()
					s.On("Send", &pb.TranslateAccessTokenRequest{OpaqueAccessToken: "test-opaque", Metadata: map[string]string{}}).Return(nil).Once()
					s.On("Recv").Return((*pb.TranslateAccessTokenResponse)(nil), errors.New("test-err")).Once()

					m := mocks.NewAuthClient()
					// Stream will be renewed on second call.
					m.On("TranslateAccessToken", mock.Anything, mock.Anything).Return(s, nil).Twice()
					return m
				}(),
			},
			args:    args{opaqueAccessToken: "test-opaque"},
			wantErr: true,
		},
		{
			name: "Test error is returned on Send error",
			fields: fields{
				grpcClient: func() mocks.AuthClient {
					s := mocks.NewAuthStreamClient()
					s.On("Send", &pb.TranslateAccessTokenRequest{OpaqueAccessToken: "test-opaque", Metadata: map[string]string{}}).Return(errors.New("test-err")).Once()

					m := mocks.NewAuthClient()
					// Stream will be renewed on second call.
					m.On("TranslateAccessToken", mock.Anything, mock.Anything).Return(s, nil).Twice()
					return m
				}(),
			},
			args:    args{opaqueAccessToken: "test-opaque"},
			wantErr: true,
		},
		{
			name: "Test job is retried on io.EOF until result is returned",
			fields: fields{
				grpcClient: func() mocks.AuthClient {
					s := mocks.NewAuthStreamClient()
					s.On("Send", &pb.TranslateAccessTokenRequest{OpaqueAccessToken: "test-opaque", Metadata: map[string]string{}}).Return(io.EOF).Once()
					s.On("Send", &pb.TranslateAccessTokenRequest{OpaqueAccessToken: "test-opaque", Metadata: map[string]string{}}).Return(nil).Once()
					s.On("Recv").Return(&pb.TranslateAccessTokenResponse{AccessToken: "test-token"}, nil).Once()

					m := mocks.NewAuthClient()
					m.On("TranslateAccessToken", mock.Anything, mock.Anything).Return(s, nil).Once()
					return m
				}(),
				config: Config{JobQueueSize: 2},
			},
			args:    args{opaqueAccessToken: "test-opaque"},
			want:    "test-token",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()

			tr := NewTranslator(tt.fields.grpcClient, tt.fields.config)
			go tr.Run(ctx)

			got, err := tr.TranslateAccessToken(ctx, tt.args.opaqueAccessToken)
			if (err != nil) != tt.wantErr {
				t.Errorf("Translator.TranslateAccessToken():\n error = %v\n wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				return
			}

			if got != tt.want {
				t.Errorf("Translator.TranslateAccessToken():\n got = %v\n want = %v", got, tt.want)
			}
		})
	}
}

func TestTranslator_maybeSendRenewStreamSig(t *testing.T) {
	t.Run("Test a streamAborted signal is sent on an unknown error", func(t *testing.T) {
		testErr := errors.New("test-error")
		tr := &Translator{
			streamAborted: make(chan struct{}),
		}

		up := make(chan struct{})
		done := make(chan struct{})
		go func() {
			up <- struct{}{}
			<-tr.streamAborted
			done <- struct{}{}
		}()

		// Wait for the goroutine to start up.
		<-up

		before := time.Now()

		tr.maybeSendRenewStreamSig(testErr)

		select {
		case <-time.After(time.Millisecond):
			t.Errorf("Translator.maybeSendRenewStreamSig(): Func did not send stream renewal signal. Time passed: %vs", time.Since(before).Seconds())
		case <-done:
			return
		}
	})

	t.Run("Test a streamAborted signal is sent on an unknown error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		testErr := errors.New("test-error")
		tr := &Translator{
			streamAborted: make(chan struct{}),
		}

		up := make(chan struct{})
		go func() {
			up <- struct{}{}
			select {
			case tr.streamAborted <- struct{}{}:
				return
			case <-ctx.Done():
				return
			}
		}()

		// Wait for the goroutine to start up.
		<-up

		up = make(chan struct{})
		done := make(chan struct{})
		go func() {
			up <- struct{}{}
			tr.maybeSendRenewStreamSig(testErr)
			done <- struct{}{}
		}()

		// Wait for the goroutine to start up.
		<-up
		before := time.Now()

		select {
		case <-time.After(time.Millisecond):
			t.Errorf("Translator.maybeSendRenewStreamSig(): Func is blocking. Time passed: %vs", time.Since(before).Seconds())
		case <-done:
			return
		}
	})
}

func Test_makeResult(t *testing.T) {
	type args struct {
		accessToken string
		metadata    map[string]string
		err         error
	}
	tests := []struct {
		name string
		args args
		want result
	}{
		{
			name: "Test fields are assigned",
			args: args{
				accessToken: "test-token",
				err:         io.EOF,
			},
			want: result{
				TranslatedAccessToken: "test-token",
				Err:                   io.EOF,
				Metadata:              nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeResult(tt.args.accessToken, tt.args.metadata, tt.args.err); !cmp.Equal(got, tt.want, cmp.Comparer(func(err, err2 error) bool { return err == err2 })) {
				t.Errorf("makeResult():\n got = %v\n want = %v", got, tt.want)
			}
		})
	}
}
