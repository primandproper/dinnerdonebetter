package manager

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"
	notificationsstore "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/notificationsstore"

	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterNotificationsDataManager registers the notifications data manager with the injector.
func RegisterNotificationsDataManager(i do.Injector) {
	// The store is platform's now, with this application's recording around its
	// writes and an adapter presenting it in this application's vocabulary. The
	// manager did not change; see notificationsstore.Adapter for why the
	// translation is there rather than here.
	notificationsstore.RegisterNotificationsStore(i)
	notificationsstore.RegisterNotificationsRepository(i)

	do.Provide[notificationsRepo](i, func(i do.Injector) (notificationsRepo, error) {
		return do.MustInvoke[notifications.Repository](i), nil
	})

	do.Provide[NotificationsDataManager](i, func(i do.Injector) (NotificationsDataManager, error) {
		return NewNotificationsDataManager(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[notificationsRepo](i),
		)
	})

	// Bind NotificationsDataManager to notifications.Repository
	do.Provide[notifications.Repository](i, func(i do.Injector) (notifications.Repository, error) {
		return do.MustInvoke[NotificationsDataManager](i), nil
	})
}
