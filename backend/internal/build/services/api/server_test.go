package api

import (
	"context"
	"testing"

	grpcapi "github.com/primandproper/dinnerdonebetter/backend/internal/build/services/api/grpc"
	httpapi "github.com/primandproper/dinnerdonebetter/backend/internal/build/services/api/http"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"

	"github.com/primandproper/platform-go/v12/server/http"

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
