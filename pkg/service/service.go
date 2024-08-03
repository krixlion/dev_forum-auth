package service

import (
	"context"
	"net"

	"fmt"

	"github.com/krixlion/dev_forum-auth/pkg/storage"
	"github.com/krixlion/dev_forum-lib/event"
	"github.com/krixlion/dev_forum-lib/event/dispatcher"
	"github.com/krixlion/dev_forum-lib/logging"
	"google.golang.org/grpc"
)

type AuthService struct {
	grpcPort   int
	grpcServer *grpc.Server

	broker     event.Broker
	dispatcher *dispatcher.Dispatcher
	logger     logging.Logger
	shutdown   func() error
}

type Dependencies struct {
	Logger       logging.Logger
	Broker       event.Broker
	GRPCServer   *grpc.Server
	Storage      storage.Storage
	Dispatcher   *dispatcher.Dispatcher
	ShutdownFunc func() error
}

func NewAuthService(grpcPort int, d Dependencies) *AuthService {
	s := &AuthService{
		grpcPort:   grpcPort,
		dispatcher: d.Dispatcher,
		grpcServer: d.GRPCServer,
		broker:     d.Broker,
		logger:     d.Logger,
		shutdown:   d.ShutdownFunc,
	}

	return s
}

func (s *AuthService) Run(ctx context.Context) {
	if err := ctx.Err(); err != nil {
		return
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", s.grpcPort))
	if err != nil {
		s.logger.Log(ctx, "failed to create a listener", "transport", "grpc", "err", err)
		return
	}

	providers, err := s.eventProviders(ctx)
	if err != nil {
		s.logger.Log(ctx, "failed to register event providers", "transport", "pubsub", "err", err)
		return
	}

	s.dispatcher.AddEventProviders(providers...)
	go s.dispatcher.Run(ctx)

	s.logger.Log(ctx, "listening", "transport", "grpc", "port", s.grpcPort)

	if err := s.grpcServer.Serve(lis); err != nil {
		s.logger.Log(ctx, "failed to serve", "transport", "grpc", "err", err)
	}
}

func (s *AuthService) eventProviders(ctx context.Context) ([]<-chan event.Event, error) {
	eTypes := map[string]event.EventType{
		"DeleteStaleTokens": event.UserDeleted,
	}

	chans := make([]<-chan event.Event, 0, len(eTypes))

	for queueName, eType := range eTypes {
		ch, err := s.broker.Consume(ctx, queueName, eType)
		if err != nil {
			return nil, err
		}

		chans = append(chans, ch)
	}

	return chans, nil
}

func (s *AuthService) Close() error {
	return s.shutdown()
}
