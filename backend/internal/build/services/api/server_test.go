package api

import (
	"context"
	"errors"
	"testing"

	grpcapi "github.com/primandproper/dinnerdonebetter/backend/internal/build/services/api/grpc"
	httpapi "github.com/primandproper/dinnerdonebetter/backend/internal/build/services/api/http"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"

	"github.com/primandproper/platform-go/v13/jobs"
	"github.com/primandproper/platform-go/v13/metering"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	"github.com/primandproper/platform-go/v13/operations"
	"github.com/primandproper/platform-go/v13/outbox"
	"github.com/primandproper/platform-go/v13/saga"
	"github.com/primandproper/platform-go/v13/server/grpc"
	"github.com/primandproper/platform-go/v13/server/http"
	"github.com/primandproper/platform-go/v13/service"
	"github.com/primandproper/platform-go/v13/webhooks"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSharedInjector_HTTPAndGRPCServersShareOneContainer is a regression test for the
// combined API server's DI wiring: both servers must come from a single container so
// the DB pool, message queue connections, and observability stack are built once and
// migrations run exactly once per boot. samber/do panics on duplicate registration, so
// registering the HTTP-specific services onto the gRPC injector both proves the sets
// are disjoint and guards against a future registration drifting into both builders.
func TestSharedInjector_HTTPAndGRPCServersShareOneContainer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := &config.APIServiceConfig{}

	injector := grpcapi.BuildInjector(ctx, cfg)

	require.NotPanics(t, func() {
		httpapi.RegisterHTTPServerServices(injector)
	}, "HTTP-specific registrations must not overlap the shared gRPC injector's contents")

	services := injector.ListProvidedServices()
	provided := make(map[string]struct{}, len(services))
	for _, svc := range services {
		provided[svc.Service] = struct{}{}
	}

	_, hasHTTPServer := provided[do.NameOf[http.Server]()]
	assert.True(t, hasHTTPServer, "shared injector should provide the HTTP server")

	_, hasGRPCServer := provided[do.NameOf[*grpcapi.GRPCService]()]
	assert.True(t, hasGRPCServer, "shared injector should provide the gRPC server")
}

// TestAPIInjector_ProvidesTheIngressServiceResolves pins the two servers platform's
// service.New reaches for when it walks the container. It resolves the platform types —
// http.Server and *grpc.Server — not the *GRPCService wrapper the old bootstrap invoked,
// so a registration that stopped providing either would leave the service with no ingress
// and nothing to say about it: New resolves servers optionally, and a service with no
// servers starts, reports nothing, and serves nothing.
func TestAPIInjector_ProvidesTheIngressServiceResolves(t *testing.T) {
	t.Parallel()

	provided := providedServices(t)

	_, hasHTTPServer := provided[do.NameOf[http.Server]()]
	assert.True(t, hasHTTPServer, "service.New resolves http.Server as ingress")

	_, hasGRPCServer := provided[do.NameOf[*grpc.Server]()]
	assert.True(t, hasGRPCServer, "service.New resolves *grpc.Server as ingress")
}

// TestAPIInjector_RegistersNoBackgroundRunners is the check issue #1354 asked for by name:
// under service's ordering the outbox relay outlives ingress, and this process has no relay
// to order. The API server writes data-change events into the outbox table in the same
// transaction as the row that caused them, and the scheduler drains them — so there is
// nothing here whose last cycle has to run after the last request commits.
//
// Every loop below is one service.New would find and start if it were registered. This
// process registers none of them, and the assertion is what keeps that a decision rather
// than an accident: a runner that drifts into the API container would silently start
// running here the moment service.New walked past it.
func TestAPIInjector_RegistersNoBackgroundRunners(t *testing.T) {
	t.Parallel()

	provided := providedServices(t)

	for name, runner := range map[string]string{
		"outbox relay":      do.NameOf[*outbox.Relay](),
		"jobs pool":         do.NameOf[*jobs.Pool](),
		"jobs scheduler":    do.NameOf[*jobs.Scheduler](),
		"saga worker":       do.NameOf[*saga.Worker](),
		"webhooks worker":   do.NameOf[*webhooks.Worker](),
		"operations worker": do.NameOf[*operations.Worker](),
		"metering flusher":  do.NameOf[*metering.Flusher](),
	} {
		_, registered := provided[runner]
		assert.False(t, registered, "the API server does not run the %s; that belongs to the scheduler", name)
	}
}

// TestProvideServiceConfig_NamesTheServiceAndBoundsItsShutdown covers the one thing
// service.New reads a *service.Config for. Without it New fails outright, and with a zero
// ShutdownTimeout every shutdown would start on an already-expired deadline.
func TestProvideServiceConfig_NamesTheServiceAndBoundsItsShutdown(t *testing.T) {
	t.Parallel()

	injector := do.New()
	provideServiceConfig(injector)

	cfg, err := do.Invoke[*service.Config](injector)
	require.NoError(t, err)

	assert.Equal(t, serviceName, cfg.Name)
	assert.Equal(t, shutdownTimeout, cfg.ShutdownTimeout)
	assert.Positive(t, cfg.ShutdownTimeout, "a zero budget makes every shutdown an expired deadline")
}

// TestServer_Run_ReleasesWhatTheServiceDoesNotOwn covers the half of Run that is still
// this package's: the profiler main built outside the container, and the container itself.
// Both have to happen after service.Run has finished draining, and the container has to
// happen last.
func TestServer_Run_ReleasesWhatTheServiceDoesNotOwn(t *testing.T) {
	t.Parallel()

	profiler := &recordingProfiler{}

	var order []string

	profiler.onStart = func() { order = append(order, "profiler start") }
	profiler.onShutdown = func() { order = append(order, "profiler shutdown") }

	srv := &Server{
		logger:   loggingnoop.NewLogger(),
		svc:      emptyService(t),
		profiler: profiler,
		shutdownContainer: func(context.Context) error {
			order = append(order, "container shutdown")
			return nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	require.NoError(t, srv.Run(ctx))

	assert.Equal(t, []string{"profiler start", "profiler shutdown", "container shutdown"}, order)
}

// TestServer_Run_SurvivesAProfilerThatWillNotStart keeps a profiling agent that cannot
// reach its collector from being the reason the API does not serve. It was logged and
// ignored before the port and it still is.
func TestServer_Run_SurvivesAProfilerThatWillNotStart(t *testing.T) {
	t.Parallel()

	profiler := &recordingProfiler{startErr: errors.New("no collector"), shutdownErr: errors.New("nothing to stop")}

	var containerShutDown bool

	srv := &Server{
		logger:   loggingnoop.NewLogger(),
		svc:      emptyService(t),
		profiler: profiler,
		shutdownContainer: func(context.Context) error {
			containerShutDown = true
			return nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	require.NoError(t, srv.Run(ctx))
	assert.True(t, containerShutDown, "a profiler that fails to shut down must not skip the container")
}

// TestServer_Run_ReleasesOnAContextTheShutdownDidNotCancel guards the reason Run rebuilds
// the context it releases on: Run returns because ctx was cancelled, and releasing on that
// same context would cancel every release it is made of.
func TestServer_Run_ReleasesOnAContextTheShutdownDidNotCancel(t *testing.T) {
	t.Parallel()

	profiler := &recordingProfiler{}

	var (
		profilerCtxErr  error
		containerCtxErr error
	)

	profiler.onShutdownCtx = func(ctx context.Context) { profilerCtxErr = ctx.Err() }

	srv := &Server{
		logger:   loggingnoop.NewLogger(),
		svc:      emptyService(t),
		profiler: profiler,
		shutdownContainer: func(ctx context.Context) error {
			containerCtxErr = ctx.Err()
			return nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	require.NoError(t, srv.Run(ctx))

	assert.NoError(t, profilerCtxErr)
	assert.NoError(t, containerCtxErr)
}

// providedServices returns the names of every service the combined API container provides.
func providedServices(t *testing.T) map[string]struct{} {
	t.Helper()

	injector := grpcapi.BuildInjector(context.Background(), &config.APIServiceConfig{})
	httpapi.RegisterHTTPServerServices(injector)

	services := injector.ListProvidedServices()
	provided := make(map[string]struct{}, len(services))
	for idx := range services {
		provided[services[idx].Service] = struct{}{}
	}

	return provided
}

// emptyService is a service.Service with nothing registered: no servers, no loops, no
// clients. Run on it does the whole of its lifecycle without binding anything, which is
// what makes the ordering around it testable.
func emptyService(t *testing.T) *service.Service {
	t.Helper()

	injector := do.New()
	provideServiceConfig(injector)

	svc, err := service.New(injector)
	require.NoError(t, err)

	return svc
}

type recordingProfiler struct {
	onStart       func()
	onShutdown    func()
	onShutdownCtx func(context.Context)
	startErr      error
	shutdownErr   error
}

func (p *recordingProfiler) Start(context.Context) error {
	if p.onStart != nil {
		p.onStart()
	}

	return p.startErr
}

func (p *recordingProfiler) Shutdown(ctx context.Context) error {
	if p.onShutdown != nil {
		p.onShutdown()
	}

	if p.onShutdownCtx != nil {
		p.onShutdownCtx(ctx)
	}

	return p.shutdownErr
}
