package servertest

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/krixlion/dev_forum-auth/pkg/grpc/server"
	pb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1"
	"github.com/krixlion/dev_forum-auth/pkg/storage"
	"github.com/krixlion/dev_forum-auth/pkg/tokens"
	"github.com/krixlion/dev_forum-lib/event/dispatcher"
	"github.com/krixlion/dev_forum-lib/nulls"
	userPb "github.com/krixlion/dev_forum-user/pkg/grpc/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

// Struct for server mock dependencies.
type Deps struct {
	VerifyClientCert bool
	Now              func() time.Time
	Storage          storage.Storage
	Vault            storage.Vault
	UserClient       userPb.UserServiceClient
	TokenManager     tokens.Manager
}

// NewServer initializes and runs in the background a gRPC
// server allowing only for local calls for testing.
// Returns a client to interact with the server.
// The server is shutdown and the client is closed when provided
// context is cancelled. No interceptors are registered.
func NewServer(ctx context.Context, d Deps) pb.AuthServiceClient {
	// bufconn allows the server to call itself
	// great for testing across whole infrastructure
	lis := bufconn.Listen(1024 * 1024)
	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	config := server.Config{
		VerifyClientCert:         d.VerifyClientCert,
		AccessTokenValidityTime:  time.Minute,
		RefreshTokenValidityTime: time.Minute,
		Now:                      d.Now,
	}

	deps := server.Dependencies{
		Services: server.Services{
			User: d.UserClient,
		},
		Vault:        d.Vault,
		TokenManager: d.TokenManager,
		Storage:      d.Storage,
		Dispatcher:   dispatcher.NewDispatcher(0),
		Logger:       nulls.NullLogger{},
		Tracer:       nulls.NullTracer{},
	}

	s := grpc.NewServer()
	pb.RegisterAuthServiceServer(s, server.MakeAuthServer(deps, config))

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with an error: %v", err)
		}
	}()

	conn, err := grpc.NewClient("passthrough://bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial bufnet: %v", err)
	}

	go func() {
		<-ctx.Done()
		s.Stop()
		if err := conn.Close(); err != nil {
			log.Fatalf("Failed to close client conn, err: %v", err)
		}
	}()

	return pb.NewAuthServiceClient(conn)
}
