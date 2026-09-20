package identitystore

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"

	platformidentity "github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterIdentityStore registers platform's identity store and the service over it.
//
// The store is registered undecorated here, unlike every other adopted domain, and that is
// the seam being different rather than the rule being broken. A store decorator cannot
// record an identity operation — registering somebody is three writes in one transaction —
// so the recording is the Hooks the service is built with, and anything holding the store
// directly is holding it for reads.
func RegisterIdentityStore(i do.Injector) {
	do.Provide[platformidentity.Store](i, func(i do.Injector) (platformidentity.Store, error) {
		return platformidentity.NewSQLStore(
			do.MustInvoke[database.Client](i),
			platformidentity.WithTablePrefix(ddbidentity.TablePrefix),
			platformidentity.WithStoreLogger(do.MustInvoke[logging.Logger](i)),
			platformidentity.WithStoreTracerProvider(do.MustInvoke[tracing.Provider](i)),
			platformidentity.WithStoreMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})

	do.Provide[*Hooks](i, func(i do.Injector) (*Hooks, error) {
		return ProvideHooks(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[audit.Repository](i),
			do.MustInvoke[*events.Emitter](i),
		), nil
	})

	do.Provide[*platformidentity.Service](i, func(i do.Injector) (*platformidentity.Service, error) {
		return platformidentity.NewService(
			do.MustInvoke[database.Client](i),
			do.MustInvoke[platformidentity.Store](i),
			// Without this every identity write is a write nothing recorded, and the
			// failure is silent: NoopHooks satisfies the interface.
			platformidentity.WithHooks(do.MustInvoke[*Hooks](i)),
			platformidentity.WithServiceLogger(do.MustInvoke[logging.Logger](i)),
			platformidentity.WithServiceTracerProvider(do.MustInvoke[tracing.Provider](i)),
			platformidentity.WithServiceMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}
