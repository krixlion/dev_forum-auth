package main

import (
	"context"
	"errors"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/krixlion/dev_forum-auth/pkg/grpc/server"
	pb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1"
	"github.com/krixlion/dev_forum-auth/pkg/service"
	"github.com/krixlion/dev_forum-auth/pkg/storage/mongo"
	"github.com/krixlion/dev_forum-auth/pkg/storage/vault"
	"github.com/krixlion/dev_forum-auth/pkg/tokens/manager"
	"github.com/krixlion/dev_forum-lib/cert"
	"github.com/krixlion/dev_forum-lib/env"
	"github.com/krixlion/dev_forum-lib/event/broker"
	"github.com/krixlion/dev_forum-lib/event/dispatcher"
	"github.com/krixlion/dev_forum-lib/logging"
	"github.com/krixlion/dev_forum-lib/tracing"
	rabbitmq "github.com/krixlion/dev_forum-rabbitmq"
	userPb "github.com/krixlion/dev_forum-user/pkg/grpc/v1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

// Hardcoded root dir name.
const projectDir = "app"
const serviceName = "auth-service"
const issuer = "http://auth-service"

var port int
var isTLS bool

func init() {
	portFlag := flag.Int("p", 50051, "The gRPC server port")
	insecureFlag := flag.Bool("insecure", false, "Whether to not use TLS over gRPC")
	flag.Parse()
	port = *portFlag
	isTLS = !(*insecureFlag)
}

func main() {
	env.Load(projectDir)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	deps, err := getServiceDependencies(ctx, serviceName, isTLS)
	if err != nil {
		logging.Log("Failed to initialize service dependencies", "err", err)
		return
	}

	service := service.NewAuthService(port, deps)
	service.Run(ctx)

	<-ctx.Done()
	logging.Log("Service shutting down")

	defer func() {
		cancel()

		if err := service.Close(); err != nil {
			logging.Log("Failed to shutdown service", "err", err)
			return
		}

		logging.Log("Service shutdown successful")
	}()
}

// getServiceDependencies is the composition root.
// Panics on any non-nil error.
func getServiceDependencies(ctx context.Context, serviceName string, isTLS bool) (service.Dependencies, error) {
	userClientCreds := insecure.NewCredentials()
	serverCreds := insecure.NewCredentials()
	if isTLS {
		caCertPool, err := cert.LoadCaPool(os.Getenv("TLS_CA_PATH"))
		if err != nil {
			return service.Dependencies{}, err
		}

		serverCert, err := cert.LoadX509KeyPair(os.Getenv("TLS_CERT_PATH"), os.Getenv("TLS_KEY_PATH"))
		if err != nil {
			return service.Dependencies{}, err
		}

		serverCreds = cert.NewServerOptionalMTLSCreds(caCertPool, serverCert)

		userServiceClientCert, err := cert.LoadX509KeyPair(os.Getenv("TLS_USER_SERVICE_CLIENT_CERT_PATH"), os.Getenv("TLS_USER_SERVICE_CLIENT_KEY_PATH"))
		if err != nil {
			return service.Dependencies{}, err
		}

		userClientCreds = cert.NewClientMTLSCreds(caCertPool, userServiceClientCert)
	}

	shutdownTracing, err := tracing.InitProvider(ctx, serviceName)
	if err != nil {
		return service.Dependencies{}, err
	}

	tracer := otel.Tracer(serviceName)

	logger, err := logging.NewLogger()
	if err != nil {
		return service.Dependencies{}, err
	}

	storage, err := mongo.Make(os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"), logger, tracer)
	if err != nil {
		return service.Dependencies{}, err
	}

	mqConfig := rabbitmq.Config{
		QueueSize:         100,
		MaxWorkers:        100,
		ReconnectInterval: time.Second * 2,
		MaxRequests:       30,
		ClearInterval:     time.Second * 5,
		ClosedTimeout:     time.Second * 15,
	}

	mq := rabbitmq.NewRabbitMQ(serviceName, os.Getenv("MQ_USER"), os.Getenv("MQ_PASS"), os.Getenv("MQ_HOST"), os.Getenv("MQ_PORT"), mqConfig,
		rabbitmq.WithLogger(logger),
		rabbitmq.WithTracer(tracer),
	)
	broker := broker.NewBroker(mq, logger, tracer)
	dispatcher := dispatcher.NewDispatcher(20)

	dispatcher.Register(storage)

	tokenManager := manager.MakeManager(manager.Config{Issuer: issuer})

	userConn, err := grpc.DialContext(ctx, "user-service:50051",
		grpc.WithTransportCredentials(userClientCreds),
		grpc.WithChainUnaryInterceptor(
			otelgrpc.UnaryClientInterceptor(),
		),
	)
	if err != nil {
		return service.Dependencies{}, err
	}
	userClient := userPb.NewUserServiceClient(userConn)

	vaultConfig := vault.Config{
		MountPath:          os.Getenv("VAULT_MOUNT_PATH"),
		KeyCount:           10,
		KeyRefreshInterval: time.Hour * 24, // Daily
	}
	vault, err := vault.Make(os.Getenv("VAULT_HOST"), os.Getenv("VAULT_PORT"), os.Getenv("VAULT_TOKEN"), vaultConfig, tracer, logger)
	if err != nil {
		return service.Dependencies{}, err
	}

	go vault.Run(ctx)

	authConfig := server.Config{
		// AccessTokenValidityTime:  time.Minute * 15,
		VerifyClientCert:         isTLS,
		AccessTokenValidityTime:  time.Hour * 24 * 7, // One week
		RefreshTokenValidityTime: time.Hour * 24 * 7, // One week
	}

	authDependencies := server.Dependencies{
		Services: server.Services{
			User: userClient,
		},
		Storage:      storage,
		Vault:        vault,
		Logger:       logger,
		Tracer:       tracer,
		TokenManager: tokenManager,
		Dispatcher:   dispatcher,
	}

	authServer := server.MakeAuthServer(authDependencies, authConfig)

	grpcServer := grpc.NewServer(
		grpc.Creds(serverCreds),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
		grpc.ChainUnaryInterceptor(
			// grpc_auth.UnaryServerInterceptor(auth.Interceptor()),
			// grpc_recovery.UnaryServerInterceptor(),
			grpc_zap.UnaryServerInterceptor(zap.L()),
			otelgrpc.UnaryServerInterceptor(),
			authServer.ValidateRequestInterceptor(),
		),
	)

	reflection.Register(grpcServer)
	pb.RegisterAuthServiceServer(grpcServer, authServer)

	return service.Dependencies{
		Logger:     logger,
		Broker:     broker,
		GRPCServer: grpcServer,
		Storage:    storage,
		Dispatcher: dispatcher,
		ShutdownFunc: func() error {
			grpcServer.GracefulStop()
			shutdownTracing()

			return errors.Join(userConn.Close(), authServer.Close())
		},
	}, nil
}
