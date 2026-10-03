package oauth2clientsstore

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"

	platformoauth2clients "github.com/primandproper/platform-go/v14/authentication/oauth2clients"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterOAuth2ClientsStore registers platform's client registry and the service over it,
// with this application's recording hung off the service's writes.
//
// The recording is on the service rather than the store because the service is the one
// that opens the transaction, and every registration, revision and withdrawal in this
// application goes through it. Minting a credential is the one write here that also
// produces a plaintext secret, and a registration the log does not know about is the
// registration it most matters to know about.
func RegisterOAuth2ClientsStore(i do.Injector) {
	do.Provide[platformoauth2clients.Store](i, func(i do.Injector) (platformoauth2clients.Store, error) {
		logger := do.MustInvoke[logging.Logger](i)

		return platformoauth2clients.NewSQLStore(
			do.MustInvoke[database.Client](i),
			// The same namespace the four protocol tables carry, so one application's
			// oauth2 tables sort together in a database that may hold another's.
			platformoauth2clients.WithTablePrefix(branding.TablePrefix),
			platformoauth2clients.WithStoreLogger(logger),
			platformoauth2clients.WithStoreTracerProvider(do.MustInvoke[tracing.Provider](i)),
		)
	})

	do.Provide[*platformoauth2clients.Service](i, func(i do.Injector) (*platformoauth2clients.Service, error) {
		logger := do.MustInvoke[logging.Logger](i)
		tracerProvider := do.MustInvoke[tracing.Provider](i)

		return platformoauth2clients.NewService(
			do.MustInvoke[database.Client](i),
			do.MustInvoke[platformoauth2clients.Store](i),
			platformoauth2clients.WithHooks(ProvideHooks(
				logger,
				tracerProvider,
				do.MustInvoke[audit.Repository](i),
				do.MustInvoke[*events.Emitter](i),
			)),
			platformoauth2clients.WithServiceLogger(logger),
			platformoauth2clients.WithServiceTracerProvider(tracerProvider),
			platformoauth2clients.WithServiceMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}
