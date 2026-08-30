package api

import (
	"context"
	"fmt"
	"time"

	grpcapi "github.com/primandproper/dinnerdonebetter/backend/internal/build/services/api/grpc"
	httpapi "github.com/primandproper/dinnerdonebetter/backend/internal/build/services/api/http"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"

	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/profiling"
	"github.com/primandproper/platform-go/v13/service"

	"github.com/samber/do/v2"
)

const (
	// serviceName is what this process calls itself. It is the name the platform HTTP
	// server was already registered under, so the service and the server it serves from
	// report the same thing.
	serviceName = "api_server"

	// shutdownTimeout bounds the whole of service.Shutdown — draining both servers,
	// closing every client — and, after it, the release of what the service does not
	// own: the profiler and the DI container.
	shutdownTimeout = 10 * time.Second
)

type Server struct {
	logger logging.Logger

	// svc is the lifecycle platform's service package assembled from the container:
	// both servers, every client that holds a connection, and the shutdown ordering
	// between them. It is what used to be written out longhand in Run.
	svc *service.Service

	// profiler is the one pillar this process builds outside the container — main
	// builds the pillars before the container exists so a config that fails to boot
	// still logs — so it is the one pillar service.Run does not start and
	// service.Shutdown does not stop. Handing it over would mean registering it with
	// the container too, and then both the container's shutdown and Pillars.Shutdown
	// would stop it: pyroscope's uploader closes a channel in Stop, so the second one
	// panics the process on its way out.
	profiler profiling.Provider

	// shutdownContainer releases the DI container's resources at shutdown. service.Service
	// holds an ordering, not the injector, so everything registered with the container that
	// the platform's own walk does not name — the container's observability providers, the
	// multi-source analytics reporter — is still the container's to release. Storing it as a
	// plain error-returning func keeps samber/do confined to the assembly in NewServer.
	shutdownContainer func(ctx context.Context) error
}

func NewServer(ctx context.Context, pillars *observability.Pillars, cfg *config.APIServiceConfig) (*Server, error) {
	// Both servers share one DI container, so expensive singletons (DB pool, message
	// queue connections, observability stack) are built once and migrations run exactly
	// once per boot.
	injector := grpcapi.BuildInjector(ctx, cfg)
	httpapi.RegisterHTTPServerServices(injector)

	provideServiceConfig(injector)

	// New is eager: it builds the database client, the queue providers, and both servers
	// here rather than at the first request, so a misconfigured dependency is a startup
	// error. It resolves the observability pillars from the container as well, which is
	// what everything else in this process is instrumented with — deliberately not the
	// pillars above, which are a second set built before the container existed.
	svc, err := service.New(injector)
	if err != nil {
		return nil, fmt.Errorf("could not assemble the service: %w", err)
	}

	return &Server{
		logger:   logging.EnsureLogger(pillars.Logger),
		svc:      svc,
		profiler: pillars.Profiler,
		shutdownContainer: func(ctx context.Context) error {
			if report := injector.ShutdownWithContext(ctx); report != nil && !report.Succeed {
				return report
			}
			return nil
		},
	}, nil
}

// provideServiceConfig registers the *service.Config service.New reads.
//
// It is not the composition root's config: this service is composed by BuildInjector,
// which registers every subsystem itself, so the only two fields that mean anything
// here are the two New reads — the name it logs under and the budget its shutdown gets.
func provideServiceConfig(i do.Injector) {
	do.ProvideValue(i, &service.Config{
		Name:            serviceName,
		ShutdownTimeout: shutdownTimeout,
	})
}

// Run serves until a shutdown signal arrives or either server stops serving, then
// gracefully shuts everything down. A non-nil error means the server stopped on its own
// rather than by signal, so the caller should exit non-zero.
//
// service.Run is what serves and what takes the process down in the order the drains
// need — ingress first, then the background loops in reverse, then the clients, then the
// pillars. That covers the case grpcServeExited used to: a listener that fails to bind
// returns from Serve, and a server that stops serving takes the service down with it and
// reports its error, rather than leaving the pod Ready with the API down.
//
// It also narrows which signals shut the process down. This used to trap SIGHUP, SIGINT,
// SIGQUIT, and SIGTERM; service.Run traps SIGINT and SIGTERM, and the other two get their
// default disposition back. Nothing sends the API server a SIGHUP — it has no config to
// reload — and SIGQUIT's default, a stack dump, is more useful from a process that will
// not stop than a graceful shutdown that hides why.
func (s *Server) Run(ctx context.Context) error {
	if s.profiler != nil {
		if err := s.profiler.Start(ctx); err != nil {
			s.logger.Error("starting profiling provider", err)
		}
	}

	runErr := s.svc.Run(ctx)

	// By here the service has drained both servers, released its clients, and flushed the
	// pillars. What is left is what it does not own, and it is released on a context free
	// of the cancellation that ended Run — draining on a cancelled context cancels every
	// drain it is made of.
	releaseCtx, cancelRelease := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancelRelease()

	if s.profiler != nil {
		if err := s.profiler.Shutdown(releaseCtx); err != nil {
			s.logger.Error("shutting down profiling provider", err)
		}
	}

	// Last, so services (DB pool, message queue connections, telemetry) release their
	// resources after the servers have stopped using them.
	if s.shutdownContainer != nil {
		if err := s.shutdownContainer(releaseCtx); err != nil {
			s.logger.Error("shutting down DI container", err)
		}
	}

	return runErr
}
