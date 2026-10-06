package api

import (
	paymentshttp "github.com/primandproper/dinnerdonebetter/backend/internal/services/payments/http"

	analyticscfg "github.com/primandproper/primitives-go/v2/analytics/config"

	"github.com/samber/do/v2"
)

// RegisterHTTPServerServices registers the providers the HTTP API server needs beyond
// what the gRPC API injector already provides. It is safe to call on the shared gRPC
// API injector: none of these registrations overlap with that container's contents.
//
// The server itself is not among them. service.Register builds it from the HTTPServer block,
// with the encoder it serves through and the health registry it mounts at /readyz — a registry
// that already checks the database and the broker, because those are what Register registered.
// What is left is what the server serves: the payment processors' webhooks, the platform
// surfaces, and the router they are all mounted on, which this application builds itself.
func RegisterHTTPServerServices(i do.Injector) {
	analyticscfg.RegisterEventReporter(i)

	// services
	paymentshttp.RegisterPaymentsHTTP(i)

	// routes
	RegisterPlatformSurfaces(i)
	RegisterAPIRouter(i)
}
