package api

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpcapi "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/build/services/api/grpc"
	httpapi "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/build/services/api/http"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"

	"github.com/primandproper/platform-go/v5/observability"
	"github.com/primandproper/platform-go/v5/observability/logging"
	"github.com/primandproper/platform-go/v5/observability/profiling"
	"github.com/primandproper/platform-go/v5/server/http"
)

type Server struct {
	logger            logging.Logger
	grpcServer        *grpcapi.GRPCService
	httpServer        http.Server
	profilingProvider profiling.Provider
}

func NewServer(ctx context.Context, pillars *observability.Pillars, cfg *config.APIServiceConfig) (*Server, error) {
	// build our server struct.
	httpServer, err := httpapi.Build(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("could not create http server: %w", err)
	}

	grpcServer, err := grpcapi.Build(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("could not create grpc server: %w", err)
	}

	return &Server{
		logger:            logging.EnsureLogger(pillars.Logger),
		grpcServer:        grpcServer,
		httpServer:        httpServer,
		profilingProvider: pillars.Profiler,
	}, nil
}

func (s *Server) Run(ctx context.Context) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(
		signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)

	if err := s.profilingProvider.Start(ctx); err != nil {
		s.logger.Error("starting profiling provider", err)
	}

	// Run servers
	go func() {
		defer func() {
			if err := recover(); err != nil {
				s.logger.Error("HTTP server panic", fmt.Errorf("%v", err))
				panic(err)
			}
		}()
		s.httpServer.Serve()
	}()
	go func() {
		defer func() {
			if err := recover(); err != nil {
				s.logger.Error("gRPC server panic", fmt.Errorf("%v", err))
				panic(err)
			}
		}()
		s.grpcServer.Serve(ctx)
	}()

	// Wait for shutdown signal
	sig := <-signalChan
	s.logger.WithValue("signal", sig.String()).Info("received shutdown signal")

	cancelCtx, cancelShutdown := context.WithTimeout(ctx, 10*time.Second)
	defer cancelShutdown()

	s.logger.Info("shutting down")

	if err := s.profilingProvider.Shutdown(cancelCtx); err != nil {
		s.logger.Error("shutting down profiling provider", err)
	}

	if err := s.httpServer.Shutdown(cancelCtx); err != nil {
		s.logger.Error("shutting down HTTP server", err)
	}

	s.grpcServer.Shutdown(cancelCtx)
}
