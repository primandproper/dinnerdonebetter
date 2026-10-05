package notificationsstore

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"

	platformnotifications "github.com/primandproper/platform-go/v15/notifications"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterNotificationsStore registers the inbox and the device registry, each
// with this application's recording hung off its writes.
//
// The store is not registered under its own type. Everything here that writes a
// notification should write an audit entry with it, and a second registration
// would be a way to skip that by naming the wrong dependency.
func RegisterNotificationsStore(i do.Injector) {
	do.Provide[platformnotifications.Inbox](i, func(i do.Injector) (platformnotifications.Inbox, error) {
		return newStore(i)
	})

	do.Provide[platformnotifications.Registry](i, func(i do.Injector) (platformnotifications.Registry, error) {
		return newStore(i)
	})
}

// newStore builds the hooked platform store from the injector.
//
// Built twice rather than shared, because the two halves are registered
// separately and a store holds no state worth sharing: it is a querier, a
// dialect and its hooks, and all three are settled from the same config either
// way.
func newStore(i do.Injector) (platformnotifications.Store, error) {
	return provideStore(
		do.MustInvoke[context.Context](i),
		do.MustInvoke[logging.Logger](i),
		do.MustInvoke[tracing.Provider](i),
		do.MustInvoke[metrics.Provider](i),
		do.MustInvoke[audit.Repository](i),
		do.MustInvoke[*events.Emitter](i),
		do.MustInvoke[database.Client](i),
	)
}
