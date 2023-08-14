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
	"github.com/krixlion/dev_forum-lib/env"
	"github.com/krixlion/dev_forum-lib/event/broker"
	"github.com/krixlion/dev_forum-lib/event/dispatcher"
	"github.com/krixlion/dev_forum-lib/logging"
	"github.com/krixlion/dev_forum-lib/tls"
	"github.com/krixlion/dev_forum-lib/tracing"
	rabbitmq "github.com/krixlion/dev_forum-rabbitmq"
	userPb "github.com/krixlion/dev_forum-user/pkg/grpc/v1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Hardcoded root dir name.
const projectDir = "app"
const serviceName = "auth-service"
const issuer = "http://auth-service"

var port int

func init() {
	portFlag := flag.Int("p", 50051, "The gRPC server port")
	flag.Parse()
	port = *portFlag
}

func main() {
	env.Load(projectDir)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	shutdownTracing, err := tracing.InitProvider(ctx, serviceName)
	if err != nil {
		logging.Log("Failed to initialize tracing", "err", err)
	}

	service := service.NewAuthService(port, getServiceDependencies(ctx))
	service.Run(ctx)

	<-ctx.Done()
	logging.Log("Service shutting down")

	defer func() {
		cancel()
		shutdownTracing()
		err := service.Close()
		if err != nil {
			logging.Log("Failed to shutdown service", "err", err)
		} else {
			logging.Log("Service shutdown properly")
		}
	}()
}

// getServiceDependencies is the composition root.
// Panics on any non-nil error.
func getServiceDependencies(ctx context.Context) service.Dependencies {
	tracer := otel.Tracer(serviceName)

	logger, err := logging.NewLogger()
	if err != nil {
		panic(err)
	}

	dbPort := os.Getenv("DB_PORT")
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")
	storage, err := mongo.Make(dbUser, dbPass, dbHost, dbPort, dbName, logger, tracer)
	if err != nil {
		panic(err)
	}

	mqPort := os.Getenv("MQ_PORT")
	mqHost := os.Getenv("MQ_HOST")
	mqUser := os.Getenv("MQ_USER")
	mqPass := os.Getenv("MQ_PASS")

	mqConfig := rabbitmq.Config{
		QueueSize:         100,
		MaxWorkers:        100,
		ReconnectInterval: time.Second * 2,
		MaxRequests:       30,
		ClearInterval:     time.Second * 5,
		ClosedTimeout:     time.Second * 15,
	}

	mq := rabbitmq.NewRabbitMQ(serviceName, mqUser, mqPass, mqHost, mqPort, mqConfig, rabbitmq.WithLogger(logger), rabbitmq.WithTracer(tracer))
	broker := broker.NewBroker(mq, logger, tracer)
	dispatcher := dispatcher.NewDispatcher(20)

	for eType, handlers := range storage.EventHandlers() {
		dispatcher.Subscribe(eType, handlers...)
	}

	tokenManager := manager.MakeManager(manager.Config{
		Issuer: issuer,
	})

	tlsCaPath := os.Getenv("TLS_CA_PATH")
	clientCredentials, err := tls.LoadCA(tlsCaPath)
	if err != nil {
		panic(err)
	}

	userConn, err := grpc.DialContext(ctx, "user-service:50051",
		grpc.WithTransportCredentials(clientCredentials),
		grpc.WithChainUnaryInterceptor(
			otelgrpc.UnaryClientInterceptor(),
		),
	)
	if err != nil {
		panic(err)
	}
	userClient := userPb.NewUserServiceClient(userConn)

	vaultHost := os.Getenv("VAULT_HOST")
	vaultPort := os.Getenv("VAULT_PORT")
	vaultMountPath := os.Getenv("VAULT_MOUNT_PATH")
	vaultToken := os.Getenv("VAULT_TOKEN")
	vaultConfig := vault.Config{
		MountPath:          vaultMountPath,
		KeyCount:           10,
		KeyRefreshInterval: time.Hour * 24, // Daily
	}
	vault, err := vault.Make(vaultHost, vaultPort, vaultToken, vaultConfig, tracer, logger)
	if err != nil {
		panic(err)
	}

	go vault.Run(ctx)

	authConfig := server.Config{
		// AccessTokenValidityTime:  time.Minute * 15,
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

	tlsCertPath := os.Getenv("TLS_CERT_PATH")
	tlsKeyPath := os.Getenv("TLS_KEY_PATH")
	credentials, err := tls.LoadServerCredentials(tlsCertPath, tlsKeyPath)
	if err != nil {
		panic(err)
	}

	grpcServer := grpc.NewServer(
		grpc.Creds(credentials),
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

			return errors.Join(
				userConn.Close(),
				authServer.Close(),
			)
		},
	}
}
