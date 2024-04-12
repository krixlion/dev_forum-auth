// TODO: add tests
package translator

import (
	"context"
	"testing"
	"time"

	"github.com/krixlion/dev_forum-auth/pkg/grpc/server/mocks"
	pb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1"
	"github.com/stretchr/testify/mock"
)

func Test_NewTranslator(t *testing.T) {
	t.Run("Test all channels are initialized with valid buffer size", func(t *testing.T) {
		queueSize := 5
		translator := NewTranslator(nil, Config{JobQueueSize: queueSize})

		if got := cap(translator.jobs); got != queueSize {
			t.Errorf("NewTranslator():\n got = %v\n want = %v\n", got, queueSize)
		}

		if got, want := cap(translator.renewStreamC), 1; got != want {
			t.Errorf("NewTranslator():\n got = %v\n want = %v\n", got, want)
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
			t.Errorf("Run did not stop on context cancellation. Time needed for func to return: %v", time.Since(before).Seconds())
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
				Req:     &pb.TranslateAccessTokenRequest{OpaqueAccessToken: "test-opaque"},
				ResultC: make(chan result),
			}},
			want: result{
				Resp: &pb.TranslateAccessTokenResponse{AccessToken: "test-token"},
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

			if got.Resp.AccessToken != tt.want.Resp.AccessToken {
				t.Errorf("Translator.handleJobs():\n got = %v\n want = %v", got.Resp, tt.want.Resp)
			}
		})
	}
}

func Test_isStreamRenewable(t *testing.T) {
	type args struct {
		err              error
		currentBufferLen int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Test returns false when error is nil",
			args: args{
				err:              nil,
				currentBufferLen: 0,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isStreamRenewable(tt.args.err, tt.args.currentBufferLen); got != tt.want {
				t.Errorf("isStreamRenewable():\n got = %v\n want = %v", got, tt.want)
			}
		})
	}
}
