package oauth2clientsstore

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddboauth "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"

	platformoauth2clients "github.com/primandproper/platform-go/v14/authentication/oauth2clients"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterOAuth2ClientsStore registers the decorated store and the service over it.
//
// The undecorated store is not registered under its own type, for webhooksstore's reason:
// every write here owes an audit entry and a data change event, and a second registration
// would be a way to skip both by naming the wrong dependency.
//
// The service is built over the decorated store deliberately. Minting a credential is the
// one write in this package that also produces a plaintext secret, and a registration the
// log does not know about is the registration it most matters to know about.
func RegisterOAuth2ClientsStore(i do.Injector) {
	do.Provide[platformoauth2clients.Store](i, func(i do.Injector) (platformoauth2clients.Store, error) {
		logger := do.MustInvoke[logging.Logger](i)
		tracerProvider := do.MustInvoke[tracing.Provider](i)

		inner, err := platformoauth2clients.NewSQLStore(
			do.MustInvoke[database.Client](i),
			// The same namespace the four protocol tables carry, so one application's
			// oauth2 tables sort together in a database that may hold another's.
			platformoauth2clients.WithTablePrefix(ddboauth.TablePrefix),
			platformoauth2clients.WithStoreLogger(logger),
			platformoauth2clients.WithStoreTracerProvider(tracerProvider),
		)
		if err != nil {
			return nil, err
		}

		return ProvideStore(
			logger,
			tracerProvider,
			do.MustInvoke[audit.Repository](i),
			do.MustInvoke[*events.Emitter](i),
			inner,
		)
	})

	do.Provide[*platformoauth2clients.Service](i, func(i do.Injector) (*platformoauth2clients.Service, error) {
		return platformoauth2clients.NewService(
			do.MustInvoke[database.Client](i),
			do.MustInvoke[platformoauth2clients.Store](i),
			platformoauth2clients.WithServiceLogger(do.MustInvoke[logging.Logger](i)),
			platformoauth2clients.WithServiceTracerProvider(do.MustInvoke[tracing.Provider](i)),
			platformoauth2clients.WithServiceMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}
