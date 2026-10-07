package auth

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"

	"github.com/primandproper/platform-go/v15/authentication/signin/devices"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

const (
	signInDevicesO11yName = "sign_in_devices_db_client"
)

// ProvideSignInDevicesSQLStore builds the platform's sign-in device store over this
// deployment's database: where each login was last renewed from, which "where you're signed
// in" shows beside it. See internal/authentication/devices for what is read off a request.
//
// It is the store itself rather than the devices.Store seam, for the reason the refresh token
// store is: the db-cleaner job wants Sweep, which the seam does not carry. And, like that store,
// no sweeper goroutine is started — one scheduled sweep for the deployment rather than one per
// replica.
func ProvideSignInDevicesSQLStore(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	client database.Client,
) (*devices.SQLStore, error) {
	return devices.NewSQLStore(
		&devices.Config{TablePrefix: branding.TablePrefix},
		client,
		devices.WithLogger(logging.NewNamedLogger(logger, signInDevicesO11yName)),
		devices.WithTracerProvider(tracerProvider),
		devices.WithMetricsProvider(metricsProvider),
	)
}
