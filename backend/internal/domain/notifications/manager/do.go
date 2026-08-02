package manager

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"
	notificationsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/notifications"

	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterNotificationsDataManager registers the notifications data manager with the injector.
func RegisterNotificationsDataManager(i do.Injector) {
	// Register the repo provider (was included in wire.NewSet)
	notificationsrepo.RegisterNotificationsRepository(i)

	// Bind *notificationsrepo.Repository to the notificationsRepo interface
	do.Provide[notificationsRepo](i, func(i do.Injector) (notificationsRepo, error) {
		return do.MustInvoke[*notificationsrepo.Repository](i), nil
	})

	do.Provide[NotificationsDataManager](i, func(i do.Injector) (NotificationsDataManager, error) {
		return NewNotificationsDataManager(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[notificationsRepo](i),
		)
	})

	// Bind NotificationsDataManager to notifications.Repository
	do.Provide[notifications.Repository](i, func(i do.Injector) (notifications.Repository, error) {
		return do.MustInvoke[NotificationsDataManager](i), nil
	})
}
