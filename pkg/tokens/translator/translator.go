// TODO: add tests
package translator

import (
	"context"
	"errors"
	"io"
	"time"

	pb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1"
	"github.com/krixlion/dev_forum-lib/logging"
	"github.com/krixlion/dev_forum-lib/nulls"
	sync "github.com/sasha-s/go-deadlock"
)

type Translator struct {
	grpcClient   pb.AuthServiceClient
	mu           *sync.RWMutex // Protects the stream.
	stream       pb.AuthService_TranslateAccessTokenClient
	renewStreamC chan struct{}
	jobs         chan job
	logger       logging.Logger
	config       Config
}

type Config struct {
	// Duration between stream renewal attempts.
	StreamRenewalInterval time.Duration
	JobQueueSize          int
}

// NewTranslator returns a new, initialized instance of the Translator.
// Run() has to be invoked before use. Logging is disabled by default
// unless a logger option is given.
func NewTranslator(grpcClient pb.AuthServiceClient, config Config, opts ...Option) *Translator {
	t := &Translator{
		grpcClient:   grpcClient,
		mu:           &sync.RWMutex{},
		stream:       nil,
		renewStreamC: make(chan struct{}, 1),
		jobs:         make(chan job, config.JobQueueSize),
		logger:       nulls.NullLogger{},
		config:       config,
	}

	for _, opt := range opts {
		opt.apply(t)
	}

	return t
}

// Run starts up necessary goroutines for automatic stream renewals
// and job handling. Blocks until given context is cancelled.
// It is intended to be invoked in a seperate goroutine.
func (t *Translator) Run(ctx context.Context) {
	// Init stream on start.
	t.renewStream(ctx)

	go t.handleStreamRenewals(ctx)
	t.handleJobs(ctx)
}

// TranslateAccessToken takes in an opaqueAccessToken and translates it to an
// encoded JWT token or returns a non-nil error.
func (t *Translator) TranslateAccessToken(opaqueAccessToken string) (string, error) {
	resultC := make(chan result)
	job := job{
		OpaqueAccessToken: opaqueAccessToken,
		ResultC:           make(chan result),
	}

	t.jobs <- job
	res := <-resultC
	return res.TranslatedAccessToken, res.Err
}

// job contains the request to be made and a channel to which the
// translated token or an error will be sent. Channel should be initialized
// by the caller. Only one result is sent through it, so
// no need for a buffer. Translator will automatically close
// the channel once it sends the result.
type job struct {
	OpaqueAccessToken string
	ResultC           chan result
}

// result contains either a translated token or a non-nil error.
// Always check if the Err is nil and if it is then discard
// the response and handle the error.
type result struct {
	TranslatedAccessToken string
	Err                   error
}

// handleJobs blocks until given context is cancelled.
// It reads incoming jobs and executes them, optionally triggering a stream
// renewal on error. It is intended to be invoked in a seperate goroutine.
func (t *Translator) handleJobs(ctx context.Context) {
	for {
		select {
		case job := <-t.jobs:
			func() {
				t.mu.RLock()
				defer t.mu.RUnlock()

				if err := t.stream.Send(&pb.TranslateAccessTokenRequest{OpaqueAccessToken: job.OpaqueAccessToken}); err != nil {
					t.maybeSendRenewStreamSig(err)
					job.ResultC <- result{Err: err}
					close(job.ResultC)
					return
				}

				resp, err := t.stream.Recv()
				t.maybeSendRenewStreamSig(err)

				job.ResultC <- result{TranslatedAccessToken: resp.AccessToken, Err: err}
				close(job.ResultC)
			}()
		case <-ctx.Done():
			// TODO: Handle stream.Context() cancellation too
			return
		}
	}
}

// maybeSendRenewStreamSig sends a signal to Translator.renewStreamC if
// the following conditions are met:
//   - given error is not nil
//   - given error is not wrapped with io.EOF,
//   - renewStreamC does not have any pending, buffered signals.
//
// Use this func to determine whether the error returned by grpc.ClientStream
// methods indicates that the stream was aborted and needs to be renewed.
func (t *Translator) maybeSendRenewStreamSig(err error) {
	if isStreamRenewable(err) {
		select {
		case t.renewStreamC <- struct{}{}:
		default:
			// Stream is being renewed or is going to be renewed shortly.
			// No need to bloat the buffer.
			return
		}
	}
}

// isStreamRenewable returns true if given error is non-nil and not io.EOF.
func isStreamRenewable(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, io.EOF) {
		// TODO: add desc
		return false
	}

	return true
}

// handleStreamRenewals listens for Translator.renewStreamC signals
// and renews the stream once a signal is received.
// It blocks until given context is cancelled.
// It is intended to be invoked in a seperate goroutine.
func (t *Translator) handleStreamRenewals(ctx context.Context) {
	for {
		select {
		case <-t.renewStreamC:
			t.renewStream(ctx)
		case <-ctx.Done():
			return
		}
	}
}

// renewStream attempts to renew the stream until it succeeds or the context is cancelled.
// Mutex protecting the stream remains locked until this func returns.
func (t *Translator) renewStream(ctx context.Context) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for {
		if err := ctx.Err(); err != nil {
			return
		}

		t.logger.Log(ctx, "Renewing the token translation stream")
		var err error
		t.stream, err = t.grpcClient.TranslateAccessToken(ctx)
		if err == nil {
			return
		}

		t.logger.Log(ctx, "Failed to renew the token translation stream", "err", err)

		time.Sleep(t.config.StreamRenewalInterval)
	}
}
