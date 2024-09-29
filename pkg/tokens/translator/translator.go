package translator

import (
	"context"
	"io"
	"time"

	pb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1"
	"github.com/krixlion/dev_forum-lib/logging"
	"github.com/krixlion/dev_forum-lib/nulls"
	"github.com/krixlion/dev_forum-lib/tracing"
	sync "github.com/sasha-s/go-deadlock"
	"go.opentelemetry.io/otel/trace"
)

type Translator struct {
	grpcClient pb.AuthServiceClient

	mu     *sync.RWMutex // Protects the stream.
	stream pb.AuthService_TranslateAccessTokenClient

	// Receives signals when a stream is aborted and needs to be renewed.
	streamAborted chan struct{}

	jobs chan job

	tracer trace.Tracer
	logger logging.Logger
	config Config
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
		grpcClient:    grpcClient,
		mu:            &sync.RWMutex{},
		stream:        nil,
		streamAborted: make(chan struct{}),
		jobs:          make(chan job, config.JobQueueSize),
		logger:        nulls.NullLogger{},
		tracer:        nulls.NullTracer{},
		config:        config,
	}

	for _, opt := range opts {
		opt.apply(t)
	}

	return t
}

// Run starts up necessary goroutines for automatic stream renewals
// and job handling. Blocks until given context is cancelled.
// It is intended to be invoked in a separate goroutine.
func (t *Translator) Run(ctx context.Context) {
	// Init stream on start.
	t.renewStream(ctx)

	go t.handleStreamRenewals(ctx)
	t.handleJobs(ctx)
}

// TranslateAccessToken takes in an opaqueAccessToken and translates it to an
// encoded JWT token or returns a non-nil error.
func (t *Translator) TranslateAccessToken(ctx context.Context, opaqueAccessToken string) (_ string, err error) {
	ctx, span := t.tracer.Start(ctx, "translator.TranslateAccessToken")
	defer span.End()
	defer tracing.SetSpanErr(span, err)

	job := job{
		OpaqueAccessToken: opaqueAccessToken,
		ResultC:           make(chan result),
		Metadata:          tracing.ExtractMetadataFromContext(ctx),
	}

	t.jobs <- job
	res := <-job.ResultC
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
	Metadata          map[string]string
}

// result contains either a translated token or a non-nil error.
// Always check if the Err is nil and if it is then discard
// the response and handle the error.
type result struct {
	TranslatedAccessToken string
	Err                   error
	Metadata              map[string]string
}

// handleJobs blocks until given context is cancelled.
// It reads incoming jobs and executes them, optionally triggering a stream
// renewal on error. It is intended to be invoked in a separate goroutine.
func (t *Translator) handleJobs(ctx context.Context) {
	for {
		select {
		case job := <-t.jobs:
			func() {
				t.mu.RLock()
				defer t.mu.RUnlock()

				if err := t.stream.Send(&pb.TranslateAccessTokenRequest{OpaqueAccessToken: job.OpaqueAccessToken, Metadata: job.Metadata}); err != nil {
					t.maybeSendRenewStreamSig(err)
					if err == io.EOF {
						t.jobs <- job
						return
					}
					job.ResultC <- makeResult("", job.Metadata, err)
					close(job.ResultC)
					return
				}

				resp, err := t.stream.Recv()
				t.maybeSendRenewStreamSig(err)

				job.ResultC <- makeResult(resp.GetAccessToken(), resp.GetMetadata(), err)
				close(job.ResultC)
			}()
		case <-ctx.Done():
			return
		}
	}
}

// makeResult construct a result to respond with to a job.
// Takes an access token and err returned by the gRPC client.
// If the error is io.EOF then it will not be assigned to the result.
func makeResult(accessToken string, metadata map[string]string, err error) result {
	return result{
		TranslatedAccessToken: accessToken,
		Metadata:              metadata,
		Err:                   err,
	}
}

// maybeSendRenewStreamSig sends a signal to Translator if
// the following conditions are met:
//   - given error is not nil,
//   - given error is not io.EOF,
//   - Translator is currently not renewing the stream.
//
// Use this func to determine whether the error returned by grpc.ClientStream
// methods indicates that the stream was aborted and needs to be renewed.
func (t *Translator) maybeSendRenewStreamSig(err error) {
	if isStreamRenewable(err) {
		select {
		case t.streamAborted <- struct{}{}:
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

	if err == io.EOF {
		// Stream was closed naturally and does not need to be renewed.
		return false
	}

	return true
}

// handleStreamRenewals listens for Translator signals
// and renews the stream once a signal is received.
// It blocks until given context is cancelled.
// It is intended to be invoked in a separate goroutine.
func (t *Translator) handleStreamRenewals(ctx context.Context) {
	for {
		select {
		case <-t.streamAborted:
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
