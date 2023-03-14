package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/krixlion/dev_forum-auth/pkg/grpc/server"
	"github.com/krixlion/dev_forum-auth/pkg/service"
	"github.com/krixlion/dev_forum-auth/pkg/storage/db"
	"github.com/krixlion/dev_forum-auth/pkg/tokens"
	"github.com/krixlion/dev_forum-lib/env"
	"github.com/krixlion/dev_forum-lib/event/broker"
	"github.com/krixlion/dev_forum-lib/event/dispatcher"
	"github.com/krixlion/dev_forum-lib/logging"
	"github.com/krixlion/dev_forum-lib/tracing"
	"github.com/krixlion/dev_forum-proto/auth_service/pb"
	userPb "github.com/krixlion/dev_forum-proto/user_service/pb"
	rabbitmq "github.com/krixlion/dev_forum-rabbitmq"
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

	service := service.NewAuthService(port, getServiceDependencies())
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

// getServiceDependencies is a Composition root.
// Panics on any non-nil error.
func getServiceDependencies() service.Dependencies {
	tracer := otel.Tracer(serviceName)

	logger, err := logging.NewLogger()
	if err != nil {
		panic(err)
	}

	userServiceAccessSecret := os.Getenv("USER_SERVICE_ACCESS_SECRET")

	dbPort := os.Getenv("DB_WRITE_PORT")
	dbHost := os.Getenv("DB_WRITE_HOST")
	dbUser := os.Getenv("DB_WRITE_USER")
	dbPass := os.Getenv("DB_WRITE_PASS")
	storage, err := db.MakeDB(dbUser, dbPass, dbHost, dbPort, logger, tracer)
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
	dispatcher := dispatcher.NewDispatcher(broker, 20)

	for eType, handlers := range storage.EventHandlers() {
		dispatcher.Subscribe(eType, handlers...)
	}

	tokenManager := tokens.MakeTokenManager(issuer, nil, nil)

	conn, err := grpc.Dial("user-service:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			otelgrpc.UnaryClientInterceptor(),
		),
	)
	if err != nil {
		panic(err)
	}
	userClient := userPb.NewUserServiceClient(conn)

	authConfig := server.Config{
		AccessTokenValidityTime:  time.Minute * 5,
		RefreshTokenValidityTime: time.Minute * 5,
	}

	authDependencies := server.Dependencies{
		Services: server.Services{
			User: userClient,
		},
		Storage:      storage,
		Logger:       logger,
		TokenManager: tokenManager,
		Dispatcher:   dispatcher,
	}

	authServer := server.NewAuthServer(authDependencies, authConfig, server.Secrets{
		UserServiceAccessToken: userServiceAccessSecret,
		// PrivateKey:             secretKey,
		// PublicKey:              secretKey,
	})

	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),

		grpc_middleware.WithUnaryServerChain(
			grpc_recovery.UnaryServerInterceptor(),
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
			return authServer.Close()
		},
	}
}
