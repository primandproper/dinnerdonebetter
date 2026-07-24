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

	"github.com/samber/do/v2"
)

type Server struct {
	logger logging.Logger
	// shutdownContainer releases the DI container's resources (DB pool, message queue
	// connections, telemetry) at shutdown. It is the only part of the container the
	// running server needs; storing it as a plain error-returning func keeps samber/do
	// confined to the assembly in NewServer.
	shutdownContainer func(ctx context.Context) error
	grpcServer        *grpcapi.GRPCService
	httpServer        http.Server
	profilingProvider profiling.Provider
}

func NewServer(ctx context.Context, pillars *observability.Pillars, cfg *config.APIServiceConfig) (*Server, error) {
	// Both servers share one DI container, so expensive singletons (DB pool, message
	// queue connections, observability stack) are built once and migrations run exactly
	// once per boot.
	injector := grpcapi.BuildInjector(ctx, cfg)
	httpapi.RegisterHTTPServerServices(injector)

	httpServer, err := do.Invoke[http.Server](injector)
	if err != nil {
		return nil, fmt.Errorf("could not create http server: %w", err)
	}

	grpcServer, err := do.Invoke[*grpcapi.GRPCService](injector)
	if err != nil {
		return nil, fmt.Errorf("could not create grpc server: %w", err)
	}

	return &Server{
		logger: logging.EnsureLogger(pillars.Logger),
		shutdownContainer: func(ctx context.Context) error {
			if report := injector.ShutdownWithContext(ctx); report != nil && !report.Succeed {
				return report
			}
			return nil
		},
		grpcServer:        grpcServer,
		httpServer:        httpServer,
		profilingProvider: pillars.Profiler,
	}, nil
}

// Run serves until a shutdown signal arrives or the gRPC serve loop dies, then gracefully shuts
// everything down. A non-nil error means the server stopped on its own rather than by signal, so
// the caller should exit non-zero.
func (s *Server) Run(ctx context.Context) error {
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

	// grpcServeExited is signaled when the gRPC serve loop returns. Because the platform Serve
	// only logs and returns when the listener fails to bind, a dead listener would otherwise leave
	// the pod Ready (K8s probes only hit HTTP :8000) with the primary API down.
	grpcServeExited := make(chan struct{}, 1)

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
		defer func() {
			select {
			case grpcServeExited <- struct{}{}:
			default:
			}
		}()
		s.grpcServer.Serve(ctx)
	}()

	// Wait for a shutdown signal or an unexpected gRPC serve exit (e.g. bind failure).
	var runErr error
	select {
	case sig := <-signalChan:
		s.logger.WithValue("signal", sig.String()).Info("received shutdown signal")
	case <-grpcServeExited:
		runErr = fmt.Errorf("gRPC serve loop exited before shutdown")
		s.logger.Error("gRPC server stopped serving unexpectedly", runErr)
	}

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

	// Shut down the DI container last, so services (DB pool, message queue connections,
	// telemetry) release their resources after the servers have stopped using them.
	if s.shutdownContainer != nil {
		if err := s.shutdownContainer(cancelCtx); err != nil {
			s.logger.Error("shutting down DI container", err)
		}
	}

	return runErr
}
