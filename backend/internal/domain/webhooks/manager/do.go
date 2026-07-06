package manager

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks"

	"github.com/primandproper/platform-go/v3/messagequeue"
	msgconfig "github.com/primandproper/platform-go/v3/messagequeue/config"
	"github.com/primandproper/platform-go/v3/observability/logging"
	"github.com/primandproper/platform-go/v3/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterWebhookDataManager registers the webhook data manager with the injector.
func RegisterWebhookDataManager(i do.Injector) {
	do.Provide[WebhookDataManager](i, func(i do.Injector) (WebhookDataManager, error) {
		return NewWebhookDataManager(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[webhooks.Repository](i),
			do.MustInvoke[*msgconfig.QueuesConfig](i),
			do.MustInvoke[messagequeue.PublisherProvider](i),
		)
	})
}
