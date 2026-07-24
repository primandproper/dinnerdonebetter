package manager

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/notifications"
	notificationsrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/notifications"

	"github.com/primandproper/platform-go/v6/messagequeue"
	msgconfig "github.com/primandproper/platform-go/v6/messagequeue/config"
	"github.com/primandproper/platform-go/v6/observability/logging"
	"github.com/primandproper/platform-go/v6/observability/tracing"

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
			do.MustInvoke[*msgconfig.QueuesConfig](i),
			do.MustInvoke[messagequeue.PublisherProvider](i),
		)
	})

	// Bind NotificationsDataManager to notifications.Repository
	do.Provide[notifications.Repository](i, func(i do.Injector) (notifications.Repository, error) {
		return do.MustInvoke[NotificationsDataManager](i), nil
	})
}
