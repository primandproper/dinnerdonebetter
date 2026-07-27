package datachangemessagehandler

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/auth"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/dataprivacy"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/internalops"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	notificationsmanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/notifications/manager"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks"
	identityindexing "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/identity/indexing"
	mealplanningindexing "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/indexing"

	"github.com/primandproper/platform-go/v7/analytics"
	"github.com/primandproper/platform-go/v7/email"
	"github.com/primandproper/platform-go/v7/encoding"
	"github.com/primandproper/platform-go/v7/messagequeue"
	notifications "github.com/primandproper/platform-go/v7/notifications/mobile"
	"github.com/primandproper/platform-go/v7/observability/logging"
	"github.com/primandproper/platform-go/v7/observability/metrics"
	"github.com/primandproper/platform-go/v7/observability/tracing"
	"github.com/primandproper/platform-go/v7/uploads"

	"github.com/samber/do/v2"
)

// RegisterAsyncDataChangeMessageHandler registers the async data change message handler with the injector.
func RegisterAsyncDataChangeMessageHandler(i do.Injector) {
	do.Provide[*AsyncDataChangeMessageHandler](i, func(i do.Injector) (*AsyncDataChangeMessageHandler, error) {
		return NewAsyncDataChangeMessageHandler(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[*config.AsyncMessageHandlerConfig](i),
			do.MustInvoke[identity.Repository](i),
			do.MustInvoke[dataprivacy.Repository](i),
			do.MustInvoke[webhooks.Repository](i),
			do.MustInvoke[internalops.InternalOpsDataManager](i),
			do.MustInvoke[messagequeue.ConsumerProvider](i),
			do.MustInvoke[messagequeue.PublisherProvider](i),
			do.MustInvoke[analytics.EventReporter](i),
			do.MustInvoke[email.Emailer](i),
			do.MustInvoke[uploads.UploadManager](i),
			do.MustInvoke[metrics.Provider](i),
			do.MustInvoke[encoding.ServerEncoderDecoder](i),
			do.MustInvoke[*identityindexing.UserDataIndexer](i),
			do.MustInvoke[*mealplanningindexing.MealPlanningDataIndexer](i),
			do.MustInvoke[mealplanning.Repository](i),
			do.MustInvoke[auth.PasswordResetTokenDataManager](i),
			do.MustInvoke[notificationsmanager.NotificationsDataManager](i),
			do.MustInvoke[notifications.PushNotificationSender](i),
		)
	})
}
